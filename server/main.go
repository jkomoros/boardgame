package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"net/http"
	"os"
	"strconv"
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
	Name         string
	Type         MoveFormFieldType
	DefaultValue interface{}
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
		"StateWrapper": string(boardgame.Serialize(s.game.StateWrapper.JSON())),
		"Diagram":      s.game.StateWrapper.State.Diagram(),
		"Chest":        s.renderChest(),
		"Forms":        s.generateForms(),
		"Game":         s.game,
		"Error":        errorMessage,
	}

	c.HTML(http.StatusOK, "main.tmpl", args)

}

func (s *Server) makeMove(c *gin.Context) error {

	//This method is passed a context mainly just to get info from request.

	move := s.game.PlayerMoveByName(c.PostForm("MoveType"))

	//Is it  a fixup move?
	if move == nil {
		move = s.game.FixUpMoveByName(c.PostForm("MoveType"))
	}

	if move == nil {
		return errors.New("Invalid MoveType")
	}

	//TODO: should we use gin's Binding to do this instead?

	for _, field := range formFields(move) {

		rawVal := c.PostForm(field.Name)

		switch field.Type {
		case FieldInt:
			if rawVal == "" {
				return errors.New(fmt.Sprint("An int field had no value", field.Name))
			}
			num, err := strconv.Atoi(rawVal)
			if err != nil {
				return errors.New(fmt.Sprint("Couldn't set field", field.Name, err))
			}
			move.SetProp(field.Name, num)
		case FieldBool:
			if rawVal == "" {
				move.SetProp(field.Name, false)
				continue
			}
			num, err := strconv.Atoi(rawVal)
			if err != nil {
				return errors.New(fmt.Sprint("Couldn't set field", field.Name, err))
			}
			if num == 1 {
				move.SetProp(field.Name, true)
			} else {
				move.SetProp(field.Name, false)
			}
		case FieldUnknown:
			return errors.New(fmt.Sprint("Field", field.Name, "was an unknown value type"))
		}
	}

	if err := <-s.game.ProposeMove(move); err != nil {
		return errors.New(fmt.Sprint("Applying move failed", err))
	}
	//TODO: it would be nice if we could show which fixup moves we made, too,
	//somehow.

	return nil
}

func (s *Server) generateForms() []*MoveForm {

	var result []*MoveForm

	for _, move := range s.game.PlayerMoves() {

		move.DefaultsForState(s.game.StateWrapper.State)

		moveItem := &MoveForm{
			Name:        move.Name(),
			Description: move.Description(),
			Fields:      formFields(move),
		}
		result = append(result, moveItem)
	}

	return result
}

func formFields(move boardgame.Move) []*MoveFormField {

	var result []*MoveFormField

	for _, fieldName := range move.Props() {

		val := move.Prop(fieldName)

		var fieldType MoveFormFieldType

		switch val.(type) {
		default:
			fieldType = FieldUnknown
		case int:
			fieldType = FieldInt
		case bool:
			fieldType = FieldBool
		}

		result = append(result, &MoveFormField{
			Name:         fieldName,
			Type:         fieldType,
			DefaultValue: val,
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

	router.Static("/static", os.ExpandEnv(pathToLib)+"static/")

	router.GET("/", s.viewHandler)
	router.POST("/", s.viewHandler)

	router.Run(":8080")

}
