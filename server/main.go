package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	games   map[string]*boardgame.Game
	factory GameFactory
	//We store the last error so that next time viewHandler is called we can
	//display it. Yes, this is a hack.
	lastErrorMessage string
}

type GameFactory func() *boardgame.Game

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

func NewServer(factory GameFactory) *Server {

	return &Server{
		games:   make(map[string]*boardgame.Game),
		factory: factory,
	}

}

//TODO: use go.rice here?
const (
	pathToLib = "$GOPATH/src/github.com/jkomoros/boardgame/server/"
)

func (s *Server) viewHandler(c *gin.Context) {
	//We serve the main app for every thing that we don't otherwise have a
	//handler for.
	c.HTML(http.StatusOK, "index.html", nil)

}

//gameAPISetup fetches the game configured in the URL and puts it in context.
func (s *Server) gameAPISetup(c *gin.Context) {

	id := c.Param("id")

	game := s.games[id]

	if game == nil {
		log.Println("Couldn't find game with id", id)
		return
	}

	c.Set("game", game)

}

func (s *Server) gameStatusHandler(c *gin.Context) {
	//This handler is designed to be a very simple status marker for the
	//current version of the specific game. It will be hit hard by all
	//clients, repeatedly, so it should be very fast.

	//TODO: use memcache for this handler

	obj, _ := c.Get("game")

	if obj == nil {
		c.JSON(http.StatusOK, gin.H{
			//TODO: handle this kind of rendering somewhere central
			"Status": "Failure",
			"Error":  "Not Found",
		})
		return
	}

	game := obj.(*boardgame.Game)

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Version": game.Version(),
	})
}

func (s *Server) newGameHandler(c *gin.Context) {
	game := s.factory()
	s.games[game.Id()] = game

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"GameId": game.Id(),
	})
}

func (s *Server) listGamesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Games":  s.listGames(),
	})
}

func (s *Server) listGames() []*boardgame.Game {

	result := make([]*boardgame.Game, len(s.games))

	i := 0
	for _, val := range s.games {
		result[i] = val
		i++
	}

	return result
}

func (s *Server) gameViewHandler(c *gin.Context) {

	obj, ok := c.Get("game")

	if !ok {
		c.JSON(http.StatusOK, gin.H{
			//TODO: handle this kind of rendering somewhere central
			"Status": "Failure",
			"Error":  "Not Found",
		})
		return
	}

	game := obj.(*boardgame.Game)

	args := gin.H{
		"Diagram": game.CurrentState().Diagram(),
		"Chest":   s.renderChest(game),
		"Forms":   s.generateForms(game),
		"Game":    game,
		"Error":   s.lastErrorMessage,
		"Status":  "Success",
	}

	s.lastErrorMessage = ""

	c.JSON(http.StatusOK, args)
}

func (s *Server) moveHandler(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		panic("This can only be called as a post.")
	}

	obj, _ := c.Get("game")

	if obj == nil {
		c.JSON(http.StatusOK, gin.H{
			//TODO: handle this kind of rendering somewhere central
			"Status": "Failure",
			"Error":  "Not Found",
		})
		return
	}

	game := obj.(*boardgame.Game)

	if err := s.makeMove(c, game); err != nil {
		s.lastErrorMessage = err.Error()
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Error":  s.lastErrorMessage,
	})
}

func (s *Server) makeMove(c *gin.Context, game *boardgame.Game) error {

	//This method is passed a context mainly just to get info from request.

	move := game.PlayerMoveByName(c.PostForm("MoveType"))

	//Is it  a fixup move?
	if move == nil {
		move = game.FixUpMoveByName(c.PostForm("MoveType"))
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

	if err := <-game.ProposeMove(move); err != nil {
		return errors.New(fmt.Sprint("Applying move failed", err))
	}
	//TODO: it would be nice if we could show which fixup moves we made, too,
	//somehow.

	return nil
}

func (s *Server) generateForms(game *boardgame.Game) []*MoveForm {

	var result []*MoveForm

	for _, move := range game.PlayerMoves() {

		move.DefaultsForState(game.CurrentState())

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

func (s *Server) renderChest(game *boardgame.Game) map[string][]interface{} {
	//Substantially copied from cli.renderChest().

	deck := make(map[string][]interface{})

	for _, name := range game.Chest().DeckNames() {

		components := game.Chest().Deck(name).Components()

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

	router.GET("/api/list/game", s.listGamesHandler)
	router.POST("/api/new/game", s.newGameHandler)

	gameAPIGroup := router.Group("/api/game/:id")
	gameAPIGroup.Use(s.gameAPISetup)
	//TODO: use the ID for something once we have mutiple games to decide between
	{
		gameAPIGroup.GET("view", s.gameViewHandler)
		gameAPIGroup.POST("move", s.moveHandler)
		gameAPIGroup.GET("status", s.gameStatusHandler)
	}

	router.Run(":8080")

}
