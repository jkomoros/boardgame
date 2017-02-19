/*

static server is a development server for hosting your static client-side
files for the boardgame app. When you deploy, you just upload the bundled
output and set the ErrorPage to return index.html, and no server is necessary.

static server does a bit of magic during development. It presents a consistent
view of the world, but it actually shadows your local /webapp folder on top of
the package default /webapp folder. So if there's a hit in your /webapp, it
returns that. Otherwise, it defaults to the package /webapp.

The other magic it does is /static/config-src/boardgame-config.html is actually
fetched from /static/config-src/boardgame-config-dev.html, so you can have
different endpoints configured in production and in dev.

*/
package static

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

type Server struct{}

/*
NewServer returns a new server. Get it to run by calling Start().

Use it like so:

	func main() {
		static.NewServer().Start()
	}

*/
func NewServer() *Server {
	return &Server{}
}

//TODO: figure out a more dynamic way to figure out where the other resources are.
const (
	pathToLib = "$GOPATH/src/github.com/jkomoros/boardgame/server/static/"
)

func (s *Server) viewHandler(c *gin.Context) {
	//We serve the main app for every thing that we don't otherwise have a
	//handler for.
	c.HTML(http.StatusOK, "index.html", nil)

}

//Start is where you start the server, and it never returns until it's time to shut down.
func (s *Server) Start() {

	router := gin.Default()

	expandedPathToLib := os.ExpandEnv(pathToLib)

	router.NoRoute(s.viewHandler)

	router.LoadHTMLFiles(expandedPathToLib + "webapp/index.html")

	static := router.Group("/static")
	{
		static.Static("/bower_components", expandedPathToLib+"webapp/bower_components")
		static.Static("/src", expandedPathToLib+"webapp/src")
		static.StaticFile("/config-src/boardgame-config.html", expandedPathToLib+"webapp/config-src/boardgame-config-dev.html")
		static.Static("/game-src", expandedPathToLib+"webapp/game-src")
	}
	router.Run(":8080")

}
