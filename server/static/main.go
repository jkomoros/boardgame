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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Server struct {
	fs http.FileSystem
}

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

func (s *Server) staticHandler(c *gin.Context) {
	request := c.Request
	url := request.URL.String()

	if strings.HasSuffix(url, "/") {
		c.HTML(http.StatusOK, "index.html", nil)
		return
	}

	file, _ := s.fs.Open(url)

	if file != nil {

		contents, _ := ioutil.ReadAll(file)

		c.Status(http.StatusOK)

		c.Writer.Write(contents)
		return

	}

	c.HTML(http.StatusOK, "index.html", nil)
}

//ShadowedFS is a simple FileSystem that tries the first FS and if that fails falls back on the Secondary.
type shadowedFS struct {
	Primary   http.FileSystem
	Secondary http.FileSystem
	Redirects map[string]string
}

func (s *shadowedFS) Open(name string) (http.File, error) {

	for from, to := range s.Redirects {
		if name == from {
			log.Println("Found redirect for", name, "to", to)
			return s.Open(to)
		}
	}

	if file, err := s.Primary.Open(name); err == nil {
		log.Println("Serving", name, "from primary")
		return file, nil
	}
	log.Println("Attempting to serve", name, "from secondary")
	return s.Secondary.Open(name)
}

func newShadowedFS(primary http.FileSystem, secondary http.FileSystem) *shadowedFS {
	return &shadowedFS{
		Primary:   primary,
		Secondary: secondary,
		Redirects: make(map[string]string),
	}
}

//AddRedirect adds a redirect so whenever from is fetched, we'll actually
//return the result for to. Take care to not create loops!
func (s *shadowedFS) AddRedirect(from string, to string) {
	s.Redirects[from] = to
}

//Start is where you start the server, and it never returns until it's time to shut down.
func (s *Server) Start() {

	router := gin.Default()

	expandedPathToLib := os.ExpandEnv(pathToLib)

	router.NoRoute(s.staticHandler)

	router.LoadHTMLFiles(expandedPathToLib + "webapp/index.html")

	fs := newShadowedFS(http.Dir("webapp"), http.Dir(expandedPathToLib+"webapp"))

	fs.AddRedirect("/config-src/boardgame-config.html", "/config-src/boardgame-config-dev.html")

	s.fs = fs

	router.Run(":8080")

}
