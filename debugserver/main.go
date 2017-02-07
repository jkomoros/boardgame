package debugserver

import (
	"fmt"
	"github.com/jkomoros/boardgame"
	"log"
	"net/http"
)

type Server struct {
	game *boardgame.Game
}

func NewServer(game *boardgame.Game) *Server {
	return &Server{
		game: game,
	}
}

func (s *Server) viewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

//Start is where you start the server, and it never returns until it's time to shut down.
func (s *Server) Start() {
	http.HandleFunc("/", s.viewHandler)
	log.Println("Open localhost:8080 in your browser.")
	http.ListenAndServe(":8080", nil)
}
