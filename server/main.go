package server

import (
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
	//We store the last error so that next time viewHandler is called we can
	//display it. Yes, this is a hack.
	lastErrorMessage string
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

//TODO: use go.rice here?
const (
	pathToLib = "$GOPATH/src/github.com/jkomoros/boardgame/server/"
)

func (s *Server) viewHandler(c *gin.Context) {

	c.HTML(http.StatusOK, "index.html", nil)

}

func (s *Server) gameStatusHandler(c *gin.Context) {
	//This handler is designed to be a very simple status marker for the
	//current version of the specific game. It will be hit hard by all
	//clients, repeatedly, so it should be very fast.

	//TODO: use memcache for this handler

	c.JSON(http.StatusOK, gin.H{
		"Version": s.game.StateWrapper.Version,
	})
}

func (s *Server) gameViewHandler(c *gin.Context) {

	args := gin.H{
		"Diagram": s.game.StateWrapper.State.Diagram(),
		"Chest":   s.renderChest(),
		"Forms":   s.generateForms(),
		"Game":    s.game,
		"Error":   s.lastErrorMessage,
	}

	s.lastErrorMessage = ""

	c.JSON(http.StatusOK, args)
}

func (s *Server) moveHandler(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		panic("This can only be called as a post.")
	}
	if err := s.makeMove(c); err != nil {
		s.lastErrorMessage = err.Error()
	}
	c.Redirect(http.StatusFound, "/")
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

func (s *Server) renderChest() map[string][]interface{} {
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

	return deck
}

//Start is where you start the server, and it never returns until it's time to shut down.
func (s *Server) Start() {

	router := gin.Default()

	expandedPathToLib := os.ExpandEnv(pathToLib)

	router.NoRoute(s.viewHandler)

	router.LoadHTMLFiles(expandedPathToLib + "webapp/index.html")

	router.Static("/bower_components", expandedPathToLib+"webapp/bower_components")
	router.Static("/src", expandedPathToLib+"webapp/src")
	router.Static("/game-src", expandedPathToLib+"webapp/game-src")

	router.GET("/api/game/view", s.gameViewHandler)
	router.POST("/api/game/move", s.moveHandler)
	router.GET("/api/game/status", s.gameStatusHandler)

	router.Run(":8080")

}
