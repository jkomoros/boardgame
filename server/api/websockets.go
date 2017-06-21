package api

import (
	"github.com/gorilla/websocket"
)

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
	//TODO: do work here
}

func (s *socket) writePump() {
	//TODO: do work here
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
		case <-v.notifyVersion:
			//Send message
		case <-v.doneChan:
			break
		}
	}
}

func (v *versionNotifier) registerSocket(s *socket) {
	//Should only be called by workLoop

	bucket, ok := v.sockets[s.gameId]

	if !ok {
		bucket = make(map[*socket]bool)
		v.sockets[s.gameId] = bucket
	}

	bucket[s] = true
}

func (v *versionNotifier) unregisterSocket(s *socket) {
	//Should only be called by workloop

	bucket, ok := v.sockets[s.gameId]

	if !ok {
		return
	}

	delete(bucket, s)
}
