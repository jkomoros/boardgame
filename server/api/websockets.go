package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jkomoros/boardgame"
	"log"
	"strconv"
	"time"
)

const (
	maxMessageSize = 512
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
)

const debugSockets = true

type gameVersionChanged struct {
	Id      string
	Version int
}

type versionNotifier struct {
	sockets       map[string]map[*socket]bool
	register      chan *socket
	unregister    chan *socket
	notifyVersion chan gameVersionChanged
	doneChan      chan bool
}

type socket struct {
	gameId   string
	notifier *versionNotifier
	conn     *websocket.Conn
	send     chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) socketHandler(c *gin.Context) {
	game := s.getGame(c)

	renderer := NewRenderer(c)

	if game == nil {
		renderer.Error("No such game")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		renderer.Error("Couldn't upgrade socket: " + err.Error())
		return
	}

	socket := newSocket(game.Id(), conn, s.notifier)
	s.notifier.register <- socket

}

func newSocket(gameId string, conn *websocket.Conn, notifier *versionNotifier) *socket {
	result := &socket{
		notifier: notifier,
		conn:     conn,
		send:     make(chan []byte, 256),
		gameId:   gameId,
	}
	go result.readPump()
	go result.writePump()

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
				log.Printf("error: %v", err)
			}
			break
		}
		log.Println("Unexpectedly got a message: ", message)
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

func newVersionNotifier() *versionNotifier {
	result := &versionNotifier{
		sockets:       make(map[string]map[*socket]bool),
		register:      make(chan *socket),
		unregister:    make(chan *socket),
		notifyVersion: make(chan gameVersionChanged),
		doneChan:      make(chan bool),
	}
	go result.workLoop()
	return result
}

func (v *versionNotifier) gameChanged(game *boardgame.Game) {
	v.notifyVersion <- gameVersionChanged{
		Id:      game.Id(),
		Version: game.Version(),
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
			debugLog("Sending message for " + rec.Id + " " + strconv.Itoa(rec.Version))
			//Send message
			bucket, ok := v.sockets[rec.Id]
			if ok {
				//Someone's listening!
				for socket := range bucket {
					socket.send <- []byte(strconv.Itoa(rec.Version))
				}
			}
		case <-v.doneChan:
			break
		}
	}
}

func debugLog(message string) {
	if !debugSockets {
		return
	}

	log.Println(message)
}

func (v *versionNotifier) registerSocket(s *socket) {
	//Should only be called by workLoop

	debugLog("Socket registering")

	bucket, ok := v.sockets[s.gameId]

	if !ok {
		bucket = make(map[*socket]bool)
		v.sockets[s.gameId] = bucket
	}

	bucket[s] = true
}

func (v *versionNotifier) unregisterSocket(s *socket) {
	//Should only be called by workloop

	debugLog("Socket unregistering")

	bucket, ok := v.sockets[s.gameId]

	if !ok {
		return
	}

	delete(bucket, s)
}
