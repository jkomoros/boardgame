package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
	"github.com/sirupsen/logrus"
)

// absentThreshold is how long since the last heartbeat before a player is
// flagged absent (spec §9.1). The client sends heartbeats every 10s, so a
// 30s threshold tolerates one missed heartbeat + some jitter without
// flapping.
const absentThreshold = 30 * time.Second

// absentScanInterval is how often workLoop wakes to scan lastHeartbeat for
// stale entries. 5s is fine-grained enough that "Waiting for Alice (m:ss)"
// counts up smoothly without thrashing.
const absentScanInterval = 5 * time.Second

// animationLeadMS is the minimum delay between when the server broadcasts
// an idle-lane state change and when clients should start its animation.
// 500ms gives typical LAN/Wi-Fi enough time to fetch, render, and pre-arm
// feeling laggy (spec §8.4). Per-game override via the (future)
// CompanionAnimationLeadDelegate hook.
const animationLeadMS = 500

// Companion animation timing is an explicit protocol contract, not an
// incidental copy of one renderer's current duration. A synchronized cycle
// gets 600ms of visible motion and 200ms for the next queued bundle to render
// and pre-arm before its slot. Custom longer effects must opt out of version
// synchronization on the client.
const (
	companionAnimationDurationMS = 600
	companionAnimationPrepareMS  = 200
	companionAnimationSlotMS     = companionAnimationDurationMS + companionAnimationPrepareMS
)

// socketMessage is the JSON-framed WebSocket message format.
// Clients should feature-detect: if the message starts with "{", parse as
// JSON; otherwise treat as a raw version number (legacy).
type socketMessage struct {
	Type string      `json:"type"` // "version" or "chat"
	Data interface{} `json:"data"`
}

type chatNotification struct {
	Channel   string `json:"channel"`
	MessageID string `json:"messageId"`
}

const (
	maxMessageSize = 512
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
)

type gameVersionChanged struct {
	ID      string
	Version int
}

type chatBroadcast struct {
	gameID       string
	notification chatNotification
}

// heartbeatRecord is sent on versionNotifier.heartbeat when a connected
// socket receives an application-level heartbeat from the client. Carries
// the socket's gameID + playerIndex so the notifier can update its
// lastHeartbeat map for the right (game, player) cell.
type heartbeatRecord struct {
	gameID      string
	playerIndex boardgame.PlayerIndex
	ts          time.Time
}

// presenceChanged is broadcast on versionNotifier when the absent set for
// a game changes (a player went stale, or a stale player came back). The
// existing gameInfoHandler reads the latest absent set into the JSON state
// on the next fetch — this notification is the wake-up that tells clients
// to refetch.
type presenceChanged struct {
	gameID string
}

// modeChangedRecord is sent on versionNotifier.modeChanged from HTTP handler
// goroutines (e.g. switchToSoloHandler) and consumed by workLoop, which owns
// the v.sockets map. This avoids a data race: handler goroutines must not
// read v.sockets directly.
type modeChangedRecord struct {
	gameID  string
	newMode string
}

type versionNotifier struct {
	sockets       map[string]map[*socket]bool
	register      chan *socket
	unregister    chan *socket
	notifyVersion chan gameVersionChanged
	notifyChat    chan chatBroadcast
	heartbeat     chan heartbeatRecord
	modeChanged   chan modeChangedRecord
	// presenceChanged carries gameIDs whose presence should be broadcast.
	// Same discipline as modeChanged: HTTP handler goroutines enqueue here;
	// only workLoop (which owns v.sockets) performs the actual broadcast.
	presenceChanged chan string
	doneChan        chan bool
	server          *Server

	// lastHeartbeat tracks, per (gameID, playerIndex), the wall-clock
	// time of the most recent application heartbeat. Read + written ONLY
	// by workLoop (and the absent-scan ticker that delivers events back
	// into the same goroutine via presenceChanged channels). No mutex —
	// channel discipline guarantees single-goroutine access.
	lastHeartbeat map[string]map[boardgame.PlayerIndex]time.Time
	// absent is the derived per-game set of player indices considered
	// absent at the last scan. Cleared when a game enters Finished or
	// transitions back to live. Read by gameInfoHandler concurrently —
	// guarded by absentMu since reads happen on HTTP handler goroutines.
	absent   map[string]map[boardgame.PlayerIndex]bool
	absentMu sync.RWMutex

	// animationLaneTail is the tail of each game's server-owned animation
	// lane. workLoop is its sole reader/writer, matching the sockets map's
	// channel discipline and avoiding a second synchronization mechanism.
	animationLaneTail map[string]animationLaneEntry
}

// AbsentPlayers returns the current absent-player set for the given gameID,
// as a copy suitable for serialization into the JSON state response. Empty
// list if no players are absent or the game is unknown to the notifier.
func (v *versionNotifier) AbsentPlayers(gameID string) []boardgame.PlayerIndex {
	v.absentMu.RLock()
	defer v.absentMu.RUnlock()
	set, ok := v.absent[gameID]
	if !ok {
		return nil
	}
	out := make([]boardgame.PlayerIndex, 0, len(set))
	for pi := range set {
		out = append(out, pi)
	}
	return out
}

type socket struct {
	gameID   string
	notifier *versionNotifier
	conn     *websocket.Conn
	send     chan socketFrameBatch

	// playerIndex is the seat this socket is bound to, or
	// ObserverPlayerIndex for unseated viewers (Table view connections,
	// spectators on a shared game). Populated at handshake time from
	// effectivePlayerIndex(c). Heartbeats from this socket are recorded
	// against (gameID, playerIndex).
	playerIndex boardgame.PlayerIndex
	// initialVersion is sent by registerSocket inside the notifier workLoop,
	// where an existing canonical timing for this version can be reused.
	initialVersion int
}

// socketFrameBatch is one atomic queue item. The version and its timing frame
// must either both enter the socket queue or neither does; otherwise one slow
// companion silently abandons the canonical animation lane while staying
// connected.
type socketFrameBatch [][]byte

type animationLaneEntry struct {
	version      int
	serverPlayAt int64
	// pacesNext is true only for a real version transition. Registration
	// baselines get timing for atomic catch-up but must not delay the first
	// subsequent move by consuming an otherwise invisible slot.
	pacesNext bool
}

func (s *Server) checkOriginForSocket(r *http.Request) bool {
	origin := r.Header["Origin"]

	if len(origin) == 0 {
		s.logger.Warnln("No origin headers provided")
		return true
	}

	return s.config.OriginAllowed(origin[0])
}

func (s *Server) socketHandler(c *gin.Context) {

	game := s.getGame(c)

	renderer := s.newRenderer(c)

	if game == nil {
		renderer.Error(errors.New("No such game"))
		return
	}

	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		renderer.Error(errors.New("Couldn't upgrade socket: " + err.Error()))
		return
	}

	playerIndex := s.effectivePlayerIndex(c)

	socket := newSocket(game, conn, s.notifier, playerIndex)
	s.notifier.register <- socket
	socket.startPumps()

}

func newSocket(game *boardgame.Game, conn *websocket.Conn, notifier *versionNotifier, playerIndex boardgame.PlayerIndex) *socket {
	result := &socket{
		notifier:       notifier,
		conn:           conn,
		send:           make(chan socketFrameBatch, 512), // same 1024-frame headroom, but paired frames enqueue atomically
		gameID:         game.ID(),
		playerIndex:    playerIndex,
		initialVersion: game.Version(),
	}
	return result
}

func (s *socket) startPumps() {
	// Registration happens first, so readPump can never enqueue unregister
	// before workLoop has installed this socket in its bucket.
	go s.readPump()
	go s.writePump()
}

func (s *socket) readPump() {

	//Based on implementation from https://github.com/gorilla/websocket/blob/master/examples/chat/client.go

	defer func() {
		s.notifier.unregister <- s
		s.conn.Close()
	}()

	s.conn.SetReadLimit(maxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error { s.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				s.notifier.server.logger.Errorln("Unexpected socket close error: "+err.Error(), logrus.Fields{
					"Id": s.gameID,
				})
			}
			break
		}

		// Application-level heartbeat. {"type":"heartbeat"} — body is
		// ignored. Anything else is unexpected and logged. We don't bother
		// validating that playerIndex is a real seat (ObserverPlayerIndex
		// hits this path too, harmlessly; lastHeartbeat updates won't
		// affect the absent set since absent is only computed for seated
		// players in workLoop's scan).
		var msg socketMessage
		if jerr := json.Unmarshal(message, &msg); jerr == nil && msg.Type == "heartbeat" {
			s.notifier.heartbeat <- heartbeatRecord{
				gameID:      s.gameID,
				playerIndex: s.playerIndex,
				ts:          time.Now(),
			}
			continue
		}
		if msg.Type == "clock-sync" {
			data, ok := msg.Data.(map[string]interface{})
			clientSentAt, valid := data["clientSentAt"].(float64)
			if ok && valid {
				s.SendSocketMessage("clock-sync", map[string]interface{}{
					"clientSentAt": clientSentAt,
					"serverAt":     time.Now().UnixMilli(),
				})
				continue
			}
		}

		s.notifier.server.logger.Warnln("Unexpectedly got a message from client", logrus.Fields{
			"Message": message,
			"Id":      s.gameID,
		})
	}

}

func (s *socket) writePump() {

	//Based on implementation at https://github.com/gorilla/websocket/blob/master/examples/chat/client.go

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()
	for {
		select {
		case batch, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			for _, message := range batch {
				if err := s.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			}
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}

}

// SendMessage enqueues the pre-marshaled version + timing frames on this
// socket. Both payloads are built ONCE per broadcast in workLoop (see the
// notifyVersion case) so every client in the game receives byte-identical
// frames — in particular the same serverPlayAt instant, which is the whole
// point of the timing frame (spec §8.4). They are one queue item: a full
// buffer closes the connection so the client reconnects and catches up,
// rather than leaving it connected with version data but no canonical time.
func (s *socket) SendMessage(versionData, timingData []byte) {
	batch := socketFrameBatch{versionData}
	if timingData != nil {
		batch = append(batch, timingData)
	}
	select {
	case s.send <- batch:
	default:
		// Buffer full — close the connection so the client reconnects
		// and catches up from the current version (gorilla chat pattern).
		s.conn.Close()
		return
	}
}

// marshalVersionFrames builds the shared version + version-timing payloads
// for one broadcast. timingData may be nil if marshaling fails (callers
// tolerate it); versionData always marshals (falls back to the bare
// version number for pathological cases).
func marshalVersionFrames(version int, serverSentAt, serverPlayAt int64) (versionData, timingData []byte) {
	msg := socketMessage{Type: "version", Data: version}
	versionData, err := json.Marshal(msg)
	if err != nil {
		versionData = []byte(strconv.Itoa(version))
	}
	timing := socketMessage{
		Type: "version-timing",
		Data: map[string]interface{}{
			"version":                version,
			"serverSentAt":           serverSentAt,
			"serverPlayAt":           serverPlayAt,
			"slotDurationMs":         companionAnimationSlotMS,
			"maxAnimationDurationMs": companionAnimationDurationMS,
		},
	}
	timingData, terr := json.Marshal(timing)
	if terr != nil {
		timingData = nil
	}
	return versionData, timingData
}

// nextAnimationPlayAt reserves the next slot in a game's animation lane.
// Idle games restart at the normal short lead; bursty version notifications
// advance monotonically instead of all pointing at effectively the same time.
func nextAnimationPlayAt(now, previous int64) int64 {
	result := now + animationLeadMS
	if paced := previous + companionAnimationSlotMS; paced > result {
		result = paced
	}
	return result
}

func reserveAnimationLane(now int64, version int, tail animationLaneEntry, hasTail, paceFromTail bool) (animationLaneEntry, bool) {
	if hasTail && version < tail.version {
		return tail, false
	}
	if hasTail && version == tail.version {
		return tail, true
	}
	previous := int64(0)
	if hasTail && paceFromTail && tail.pacesNext {
		previous = tail.serverPlayAt
	}
	return animationLaneEntry{
		version: version, serverPlayAt: nextAnimationPlayAt(now, previous), pacesNext: true,
	}, true
}

func (s *socket) SendChatNotification(notification chatNotification) {
	s.SendSocketMessage("chat", notification)
}

func (s *socket) SendSocketMessage(msgType string, payload interface{}) {
	msg := socketMessage{Type: msgType, Data: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case s.send <- socketFrameBatch{data}:
	default:
		s.conn.Close()
	}
}

func newVersionNotifier(s *Server) *versionNotifier {
	result := &versionNotifier{
		sockets:           make(map[string]map[*socket]bool),
		register:          make(chan *socket),
		unregister:        make(chan *socket),
		notifyVersion:     make(chan gameVersionChanged),
		notifyChat:        make(chan chatBroadcast),
		heartbeat:         make(chan heartbeatRecord, 64), // small buffer absorbs bursty heartbeats
		modeChanged:       make(chan modeChangedRecord, 4),
		presenceChanged:   make(chan string, 16),
		doneChan:          make(chan bool),
		server:            s,
		lastHeartbeat:     make(map[string]map[boardgame.PlayerIndex]time.Time),
		absent:            make(map[string]map[boardgame.PlayerIndex]bool),
		animationLaneTail: make(map[string]animationLaneEntry),
	}
	go result.workLoop()
	return result
}

func (v *versionNotifier) gameChanged(game *boardgame.GameStorageRecord) {
	v.notifyVersion <- gameVersionChanged{
		ID:      game.ID,
		Version: game.Version,
	}
}

func (v *versionNotifier) chatMessageSent(gameID, channel, messageID string) {
	v.notifyChat <- chatBroadcast{
		gameID: gameID,
		notification: chatNotification{
			Channel:   channel,
			MessageID: messageID,
		},
	}
}

func (v *versionNotifier) done() {
	close(v.doneChan)
}

func (v *versionNotifier) workLoop() {
	scanTicker := time.NewTicker(absentScanInterval)
	defer scanTicker.Stop()

	for {
		select {
		case s := <-v.register:
			v.registerSocket(s)
		case s := <-v.unregister:
			v.unregisterSocket(s)
		case rec := <-v.notifyVersion:
			v.server.logger.Debugln("Sending socket message", logrus.Fields{
				"ID":      rec.ID,
				"Version": rec.Version,
			})
			// Reserve the authoritative lane even when nobody is listening.
			// A socket whose handshake started just before this notification
			// may register just after it and must catch up to this version rather
			// than receive its stale handshake snapshot.
			tail, hasTail := v.animationLaneTail[rec.ID]
			bucket, hasListeners := v.sockets[rec.ID]
			now := time.Now().UnixMilli()
			// Unobserved transitions coalesce at now+lead. Pacing every invisible
			// fix-up would create a far-future backlog for the next reconnect.
			reserved, shouldBroadcast := reserveAnimationLane(
				now, rec.Version, tail, hasTail, hasListeners && len(bucket) > 0,
			)
			if !shouldBroadcast {
				v.server.logger.Warnln("Ignoring regressing socket version", logrus.Fields{
					"ID": rec.ID, "Version": rec.Version, "LaneVersion": tail.version,
				})
				continue
			}
			v.animationLaneTail[rec.ID] = reserved

			if hasListeners {
				// Someone's listening! Marshal once; every socket gets
				// byte-identical frames (and one shared serverPlayAt).
				versionData, timingData := marshalVersionFrames(rec.Version, now, reserved.serverPlayAt)
				for socket := range bucket {
					socket.SendMessage(versionData, timingData)
				}
			}
		case chat := <-v.notifyChat:
			bucket, ok := v.sockets[chat.gameID]
			if ok {
				for socket := range bucket {
					socket.SendChatNotification(chat.notification)
				}
			}
		case hb := <-v.heartbeat:
			gameHB, ok := v.lastHeartbeat[hb.gameID]
			if !ok {
				gameHB = make(map[boardgame.PlayerIndex]time.Time)
				v.lastHeartbeat[hb.gameID] = gameHB
			}
			gameHB[hb.playerIndex] = hb.ts
			// If this player was previously absent, clear and broadcast.
			if v.clearAbsentIfPresent(hb.gameID, hb.playerIndex) {
				v.broadcastPresenceChange(hb.gameID)
			}
		case mc := <-v.modeChanged:
			v.doBroadcastModeChanged(mc.gameID, mc.newMode)
		case gameID := <-v.presenceChanged:
			v.broadcastPresenceChange(gameID)
		case <-scanTicker.C:
			v.scanStaleHeartbeats()
		case <-v.doneChan:
			return
		}
	}
}

// scanStaleHeartbeats walks lastHeartbeat looking for entries older than
// absentThreshold. Promotes them into the absent set and broadcasts a
// presence-change notification per affected game. Also opportunistically
// evicts maps for finished games (long-running servers would otherwise
// accumulate one inner map per game ever played; see spec §11 + critic
// finding on map leaks).
//
// Called from the workLoop goroutine only; safe to read lastHeartbeat
// without locking.
func (v *versionNotifier) scanStaleHeartbeats() {
	cutoff := time.Now().Add(-absentThreshold)
	laneCutoffMS := time.Now().Add(-5 * time.Minute).UnixMilli()
	changedGames := make(map[string]bool)
	var gamesToEvict []string

	for gameID, gameHB := range v.lastHeartbeat {
		// Eviction probe: if the game is gone-or-finished, drop the map.
		// Cheap heuristic: the game has zero connected sockets AND
		// we last heard a heartbeat for it more than 5 minutes ago.
		// 5min > 30s absent threshold so a normal mid-game disconnect
		// doesn't trigger eviction.
		if _, hasSockets := v.sockets[gameID]; !hasSockets {
			anyRecent := false
			for _, ts := range gameHB {
				if time.Since(ts) < 5*time.Minute {
					anyRecent = true
					break
				}
			}
			if !anyRecent {
				gamesToEvict = append(gamesToEvict, gameID)
				continue
			}
		}

		for pi, ts := range gameHB {
			// Observers (Table views, spectators) are pi < 0 — they
			// don't occupy seats and must not appear in the absent set
			// ("Waiting for Player -1") or trip hostSkipTurn's gate.
			if pi < 0 {
				continue
			}
			if ts.Before(cutoff) {
				if v.markAbsent(gameID, pi) {
					changedGames[gameID] = true
				}
			}
		}
	}

	for _, gameID := range gamesToEvict {
		delete(v.lastHeartbeat, gameID)
		delete(v.animationLaneTail, gameID)
		v.absentMu.Lock()
		delete(v.absent, gameID)
		v.absentMu.Unlock()
	}
	// Lane entries also exist for notifications that arrived with no sockets,
	// and therefore may have no heartbeat map to drive the eviction above.
	// Retain them long enough for ordinary reconnects, then bound the map.
	for gameID, tail := range v.animationLaneTail {
		if _, hasSockets := v.sockets[gameID]; !hasSockets && tail.serverPlayAt < laneCutoffMS {
			delete(v.animationLaneTail, gameID)
		}
	}

	for gameID := range changedGames {
		v.broadcastPresenceChange(gameID)
	}
}

// markAbsent adds (gameID, pi) to the absent set; returns true if this is a
// transition (not already absent).
func (v *versionNotifier) markAbsent(gameID string, pi boardgame.PlayerIndex) bool {
	v.absentMu.Lock()
	defer v.absentMu.Unlock()
	set, ok := v.absent[gameID]
	if !ok {
		set = make(map[boardgame.PlayerIndex]bool)
		v.absent[gameID] = set
	}
	if set[pi] {
		return false
	}
	set[pi] = true
	return true
}

// clearAbsentIfPresent removes (gameID, pi) from the absent set; returns
// true if this is a transition (was previously absent).
func (v *versionNotifier) clearAbsentIfPresent(gameID string, pi boardgame.PlayerIndex) bool {
	v.absentMu.Lock()
	defer v.absentMu.Unlock()
	set, ok := v.absent[gameID]
	if !ok || !set[pi] {
		return false
	}
	delete(set, pi)
	if len(set) == 0 {
		delete(v.absent, gameID)
	}
	return true
}

// enqueuePresenceChange requests a presence-changed broadcast for gameID.
// Safe from any goroutine (e.g. joinSeatHandler): the actual broadcast
// happens in workLoop, which owns v.sockets. Non-blocking with a buffered
// channel; if the buffer is somehow full the notification is dropped —
// presence is re-derived on the next heartbeat scan, so a dropped nudge
// self-heals within absentScanInterval.
func (v *versionNotifier) enqueuePresenceChange(gameID string) {
	select {
	case v.presenceChanged <- gameID:
	default:
	}
}

// broadcastPresenceChange sends a "presence-changed" socket message to all
// sockets in the game. Client-side handler refetches gameInfo, which
// surfaces the updated Absent list. We use a distinct message type
// (rather than synthesizing a "version" with a sentinel) because the
// client-side version-targeting logic filters out versions < 0 — so a
// fake "version: -1" would never reach the state-refetch path.
// workLoop-ONLY: reads v.sockets. Handler goroutines must use
// enqueuePresenceChange instead.
func (v *versionNotifier) broadcastPresenceChange(gameID string) {
	v.broadcastSocketMessage(gameID, "presence-changed", map[string]interface{}{
		"gameID": gameID,
	})
}

// broadcastModeChanged enqueues a mode-changed notification for workLoop.
// Safe to call from any goroutine (e.g. HTTP handler goroutines). The actual
// broadcast happens inside workLoop which owns v.sockets.
func (v *versionNotifier) broadcastModeChanged(gameID, newMode string) {
	v.modeChanged <- modeChangedRecord{gameID: gameID, newMode: newMode}
}

// seedHeartbeat starts the absence clock for a player who has just claimed
// a seat, by recording a synthetic heartbeat at the claim instant. Without
// this, a player who claims a seat but never connects a socket has NO
// lastHeartbeat entry and is therefore never scanned into the absent set —
// i.e. the no-show player (the exact case host SkipTurn exists for) looks
// permanently present. Safe from any goroutine: routes through the same
// channel as real heartbeats, so workLoop stays the sole map owner.
func (v *versionNotifier) seedHeartbeat(gameID string, playerIndex boardgame.PlayerIndex) {
	v.heartbeat <- heartbeatRecord{gameID: gameID, playerIndex: playerIndex, ts: time.Now()}
}

// doBroadcastModeChanged is called from workLoop only. Sends a
// "mode-changed" socket message to every socket in the game. Client-side
// handler responds by reloading the page.
func (v *versionNotifier) doBroadcastModeChanged(gameID, newMode string) {
	v.broadcastSocketMessage(gameID, "mode-changed", map[string]interface{}{
		"newMode": newMode,
		"gameID":  gameID,
	})
}

// broadcastSocketMessage is the shared non-blocking broadcaster for the
// new socket message types added in P3-P5 (presence-changed, mode-changed).
// Non-blocking send via select+default — if a socket's send channel is
// full, drop this frame for that client rather than stalling the entire
// notifier goroutine on a slow socket. The version-changed path uses
// SendMessage which blocks; switching it to non-blocking is a separate
// hardening task (existing behavior preserved for now).
func (v *versionNotifier) broadcastSocketMessage(gameID, msgType string, data interface{}) {
	bucket, ok := v.sockets[gameID]
	if !ok {
		return
	}
	msg := socketMessage{Type: msgType, Data: data}
	payload, err := json.Marshal(msg)
	if err != nil {
		v.server.logger.Warnln("failed to marshal socket message", logrus.Fields{
			"type": msgType,
			"err":  err.Error(),
		})
		return
	}
	for socket := range bucket {
		select {
		case socket.send <- socketFrameBatch{payload}:
		default:
			// Slow client; drop this frame for them. They'll catch up
			// on the next state refetch or reconnect.
		}
	}
}

func (v *versionNotifier) registerSocket(s *socket) {
	//Should only be called by workLoop

	v.server.logger.Debugln("Socket registering", logrus.Fields{
		"ID": s.gameID,
	})

	bucket, ok := v.sockets[s.gameID]

	if !ok {
		bucket = make(map[*socket]bool)
		v.sockets[s.gameID] = bucket
	}

	bucket[s] = true

	// Catch-up uses whichever is newer: the handshake's storage snapshot or
	// the lane version already observed by workLoop. This closes the race where
	// a move commits between newSocket's snapshot and this registration.
	now := time.Now().UnixMilli()
	version := s.initialVersion
	tail, hasTail := v.animationLaneTail[s.gameID]
	if hasTail && tail.version >= version {
		version = tail.version
	} else {
		// Storage may be ahead of the notifier after startup. Make that
		// snapshot canonical so later registrations reuse the same target.
		var ok bool
		tail, ok = reserveAnimationLane(now, version, tail, hasTail, false)
		if ok {
			tail.pacesNext = false
			v.animationLaneTail[s.gameID] = tail
			hasTail = true
		}
	}
	playAt := now + animationLeadMS
	if hasTail {
		playAt = tail.serverPlayAt
	}
	versionData, timingData := marshalVersionFrames(version, now, playAt)
	s.SendMessage(versionData, timingData)
}

func (v *versionNotifier) unregisterSocket(s *socket) {
	//Should only be called by workloop

	v.server.logger.Debugln("Socket unregistering", logrus.Fields{
		"ID": s.gameID,
	})

	bucket, ok := v.sockets[s.gameID]

	if !ok {
		return
	}

	delete(bucket, s)

	// Drop the empty bucket so scanStaleHeartbeats' "game has zero
	// connected sockets" eviction probe can actually fire. Without this,
	// any game that ever had a socket keeps an empty bucket forever and
	// the lastHeartbeat/absent maps leak unboundedly.
	if len(bucket) == 0 {
		delete(v.sockets, s.gameID)
	}
}
