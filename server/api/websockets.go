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

// animationLeadMS is the default delay between when the server broadcasts
// a state change and when clients should start the resulting animation.
// 250ms gives typical LAN/Wi-Fi enough head-time to converge without
// feeling laggy (spec §8.4). Per-game override via the (future)
// CompanionAnimationLeadDelegate hook.
const animationLeadMS = 250

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

type versionNotifier struct {
	sockets       map[string]map[*socket]bool
	register      chan *socket
	unregister    chan *socket
	notifyVersion chan gameVersionChanged
	notifyChat    chan chatBroadcast
	heartbeat     chan heartbeatRecord
	doneChan      chan bool
	server        *Server

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
	send     chan []byte

	// playerIndex is the seat this socket is bound to, or
	// ObserverPlayerIndex for unseated viewers (Table view connections,
	// spectators on a shared game). Populated at handshake time from
	// effectivePlayerIndex(c). Heartbeats from this socket are recorded
	// against (gameID, playerIndex).
	playerIndex boardgame.PlayerIndex
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

}

func newSocket(game *boardgame.Game, conn *websocket.Conn, notifier *versionNotifier, playerIndex boardgame.PlayerIndex) *socket {
	result := &socket{
		notifier:    notifier,
		conn:        conn,
		send:        make(chan []byte, 256),
		gameID:      game.ID(),
		playerIndex: playerIndex,
	}
	go result.readPump()
	go result.writePump()

	//As soon as the socke tis opened, send the current version. That way if
	//the connection broke right when the version changed, we'll still catch up.
	result.SendMessage(gameVersionChanged{
		ID:      game.ID(),
		Version: game.Version(),
	})

	return result
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
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			s.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}

}

func (s *socket) SendMessage(message gameVersionChanged) {
	// Send JSON-framed message. Clients feature-detect the format.
	msg := socketMessage{Type: "version", Data: message.Version}
	data, err := json.Marshal(msg)
	if err != nil {
		// Fallback to raw version number if JSON fails
		s.send <- []byte(strconv.Itoa(message.Version))
		return
	}
	s.send <- data

	// Sibling "version-timing" message carries the cross-screen animation
	// sync timestamps (spec §8.4). Sent immediately after "version" so a
	// new client can correlate the two by version number. Old clients
	// ignore the new type — backward-compatible.
	now := time.Now().UnixMilli()
	timing := socketMessage{
		Type: "version-timing",
		Data: map[string]interface{}{
			"version":      message.Version,
			"serverSentAt": now,
			"serverPlayAt": now + animationLeadMS,
		},
	}
	if timingData, terr := json.Marshal(timing); terr == nil {
		s.send <- timingData
	}
}

func (s *socket) SendChatNotification(notification chatNotification) {
	msg := socketMessage{Type: "chat", Data: notification}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.send <- data
}

func newVersionNotifier(s *Server) *versionNotifier {
	result := &versionNotifier{
		sockets:       make(map[string]map[*socket]bool),
		register:      make(chan *socket),
		unregister:    make(chan *socket),
		notifyVersion: make(chan gameVersionChanged),
		notifyChat:    make(chan chatBroadcast),
		heartbeat:     make(chan heartbeatRecord, 64), // small buffer absorbs bursty heartbeats
		doneChan:      make(chan bool),
		server:        s,
		lastHeartbeat: make(map[string]map[boardgame.PlayerIndex]time.Time),
		absent:        make(map[string]map[boardgame.PlayerIndex]bool),
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
			//Send message
			bucket, ok := v.sockets[rec.ID]
			if ok {
				//Someone's listening!
				for socket := range bucket {
					socket.SendMessage(rec)
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
		case <-scanTicker.C:
			v.scanStaleHeartbeats()
		case <-v.doneChan:
			return
		}
	}
}

// scanStaleHeartbeats walks lastHeartbeat looking for entries older than
// absentThreshold. Promotes them into the absent set and broadcasts a
// presence-change notification per affected game. Called from the workLoop
// goroutine only; safe to read lastHeartbeat without locking.
func (v *versionNotifier) scanStaleHeartbeats() {
	cutoff := time.Now().Add(-absentThreshold)
	changedGames := make(map[string]bool)
	for gameID, gameHB := range v.lastHeartbeat {
		for pi, ts := range gameHB {
			if ts.Before(cutoff) {
				if v.markAbsent(gameID, pi) {
					changedGames[gameID] = true
				}
			}
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

// broadcastPresenceChange wakes clients up with a version-style notify
// (re-using the existing socket message channel to keep client wiring
// simple). The client refetches state, which surfaces the new Absent list.
// V1 uses a synthetic version of -1 in the broadcast envelope so clients
// can distinguish from real state changes if they care; for now the
// existing client just treats any "version" message as "go fetch state".
func (v *versionNotifier) broadcastPresenceChange(gameID string) {
	bucket, ok := v.sockets[gameID]
	if !ok {
		return
	}
	// Re-broadcast the latest game version (we don't track it here so use
	// a sentinel; clients refetch state and get the actual version).
	notif := gameVersionChanged{ID: gameID, Version: -1}
	for socket := range bucket {
		socket.SendMessage(notif)
	}
}

// broadcastModeChanged sends a "mode-changed" socket message to every
// socket currently connected to gameID. Client-side handler responds by
// reloading the page (boardgame-game-state-manager.ts); on reload the
// surface=table / surface=hand cookies are cleared by the server's
// response to switchToSolo, and the loader picks the solo renderer.
func (v *versionNotifier) broadcastModeChanged(gameID, newMode string) {
	bucket, ok := v.sockets[gameID]
	if !ok {
		return
	}
	msg := socketMessage{
		Type: "mode-changed",
		Data: map[string]interface{}{
			"newMode": newMode,
		},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	for socket := range bucket {
		socket.send <- data
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
}
