package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
)

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

type versionNotifier struct {
	sockets       map[string]map[*socket]bool
	register      chan *socket
	unregister    chan *socket
	notifyVersion chan gameVersionChanged
	doneChan      chan bool
	server        *Server
}

type socket struct {
	gameID   string
	notifier *versionNotifier
	conn     *websocket.Conn
	send     chan []byte
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

	socket := newSocket(game, conn, s.notifier)
	s.notifier.register <- socket

}

func newSocket(game *boardgame.Game, conn *websocket.Conn, notifier *versionNotifier) *socket {
	result := &socket{
		notifier: notifier,
		conn:     conn,
		send:     make(chan []byte, 256),
		gameID:   game.ID(),
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
	s.send <- []byte(strconv.Itoa(message.Version))
}

func newVersionNotifier(s *Server) *versionNotifier {
	result := &versionNotifier{
		sockets:       make(map[string]map[*socket]bool),
		register:      make(chan *socket),
		unregister:    make(chan *socket),
		notifyVersion: make(chan gameVersionChanged),
		doneChan:      make(chan bool),
		server:        s,
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

func (v *versionNotifier) done() {
	close(v.doneChan)
}

func (v *versionNotifier) workLoop() {
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
		case <-v.doneChan:
			break
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
