package server

import (
	"encoding/json"
	"github.com/jkomoros/boardgame"
	"html/template"
	"log"
	"net/http"
	"os"
	"reflect"
	"unicode"
)

type templateArgs map[string]interface{}

type Server struct {
	game *boardgame.Game
}

type MoveForm struct {
	Name        string
	Description string
	Fields      []*MoveFormField
}

type MoveFormFieldType int

const (
	FieldUnknown MoveFormFieldType = iota
	FieldInt
	FieldBool
)

type MoveFormField struct {
	Name string
	Type MoveFormFieldType
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
	args["Deck"] = s.renderDeck()
	args["Forms"] = s.generateForms()

	s.renderTemplate(w, "main", args)

}

func (s *Server) generateForms() []*MoveForm {

	var result []*MoveForm

	for _, move := range s.game.Moves() {
		moveItem := &MoveForm{
			Name:        move.Name(),
			Description: move.Description(),
			Fields:      formFields(move),
		}
		result = append(result, moveItem)
	}

	return result
}

func moveFieldNameShouldBeIncluded(name string) bool {
	//TODO: this is recreated a number of places, which implies it should be
	//in the base library.

	if len(name) < 1 {
		return false
	}

	firstChar := []rune(name)[0]

	if firstChar != unicode.ToUpper(firstChar) {
		//It was not upper case, thus private, thus should not be included.
		return false
	}

	return true
}

func formFields(move boardgame.Move) []*MoveFormField {

	var result []*MoveFormField

	s := reflect.ValueOf(move).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fieldName := typeOfT.Field(i).Name
		if !moveFieldNameShouldBeIncluded(fieldName) {
			continue
		}

		var fieldType MoveFormFieldType

		switch f.Type().Name() {
		case "int":
			fieldType = FieldInt
		case "bool":
			fieldType = FieldBool
		default:
			fieldType = FieldUnknown
		}

		result = append(result, &MoveFormField{
			Name: fieldName,
			Type: fieldType,
		})
	}

	return result
}

func (s *Server) renderDeck() string {
	//Substantially copied from cli.renderChest().

	deck := make(map[string][]interface{})

	for _, name := range s.game.Chest().DeckNames() {

		components := s.game.Chest().Deck(name).Components()

		values := make([]interface{}, len(components))

		for i, component := range components {
			values[i] = struct {
				Index  int
				Values interface{}
			}{
				i,
				component.Values,
			}
		}

		deck[name] = values
	}

	json, err := json.MarshalIndent(deck, "", "  ")

	if err != nil {
		panic(err)
	}

	return string(json)
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
