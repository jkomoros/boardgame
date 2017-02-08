package server

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"net/http"
	"os"
	"reflect"
	"unicode"
)

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

func (s *Server) viewHandler(c *gin.Context) {

	var errorMessage string

	if c.Request.Method == http.MethodPost {

		if err := s.makeMove(c); err != nil {
			errorMessage = err.Error()
		}

	}

	args := gin.H{
		"State":   string(boardgame.Serialize(s.game.State.JSON())),
		"Diagram": s.game.State.Payload.Diagram(),
		"Chest":   s.renderChest(),
		"Forms":   s.generateForms(),
		"Game":    s.game,
		"Error":   errorMessage,
	}

	c.HTML(http.StatusOK, "main.tmpl", args)

}

func (s *Server) makeMove(c *gin.Context) error {

	//This method is passed a context mainly just to get info from request.

	move := s.game.MoveByName(c.PostForm("MoveType"))

	if move == nil {
		return errors.New("Invalid MoveType")
	}

	//TODO: actually make the move.

	return errors.New("This functionality is not yet implemented")
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

func (s *Server) renderChest() string {
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

//Start is where you start the server, and it never returns until it's time to shut down.
func (s *Server) Start() {

	router := gin.Default()

	router.LoadHTMLGlob(os.ExpandEnv(pathToLib) + "templates/*")

	router.GET("/", s.viewHandler)
	router.POST("/", s.viewHandler)

	router.Run(":8080")

}
