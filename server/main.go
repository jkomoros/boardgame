package server

import (
	"github.com/jkomoros/boardgame"
	"html/template"
	"log"
	"net/http"
	"os"
)

type templateArgs map[string]interface{}

type Server struct {
	game *boardgame.Game
}

func NewServer(game *boardgame.Game) *Server {
	return &Server{
		game: game,
	}
}

const (
	pathToLib = "$GOPATH/src/github.com/jkomoros/boardgame/server/"
)

func (s *Server) viewHandler(w http.ResponseWriter, r *http.Request) {

	args := make(templateArgs)

	args["State"] = string(boardgame.Serialize(s.game.State.JSON()))
	args["Diagram"] = s.game.State.Payload.Diagram()

	s.renderTemplate(w, "main", args)

}

func (s *Server) renderTemplate(w http.ResponseWriter, tmpl string, args templateArgs) {
	//TODO: this seems brittle!
	t, err := template.ParseFiles(os.ExpandEnv(pathToLib) + "templates/" + tmpl + ".tmpl")

	if err != nil {
		panic("Couldn't find template " + tmpl + " " + err.Error())
	}

	if args == nil {
		args = make(templateArgs)
	}

	args["Game"] = s.game

	t.Execute(w, args)
}

//Start is where you start the server, and it never returns until it's time to shut down.
func (s *Server) Start() {
	http.HandleFunc("/", s.viewHandler)
	log.Println("Open localhost:8080 in your browser.")
	http.ListenAndServe(":8080", nil)
}
