/*
Package stub is a library that helps generate stub code for new games
*/
package stub

import (
	"errors"
	"go/format"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
)

// Options is the default options struct. Name is the only required field; the
// zero-value of every other field is default.
type Options struct {
	//The name of the game
	Name string
	//DisplayName to output (skipped if "")
	DisplayName string
	//Description of game to output (skipped if "")
	Description       string
	MinNumPlayers     int
	MaxNumPlayers     int
	DefaultNumPlayers int
	//If true, won't save main_test.go
	SuppressTest  bool
	SuppressPhase bool
	//If true, won't add a CurrentPlayer to gameState
	SuppressCurrentPlayer bool
	//If true, won't render moves_normal.go, moves_setup.go, or moves.go if suppressphase is true
	SuppressMovesStubs bool
	//If true, instead of using reflection for delegate.Name() will just output {{.Name}}
	SuppressReflectedName               bool
	SuppressComponentsStubs             bool
	SuppressClientRenderGame            bool
	SuppressClientRenderPlayerInfo      bool
	EnableExampleDeck                   bool
	EnableExampleDynamicComponentValues bool
	EnableExampleEndState               bool
	EnableExampleComputedProperties     bool
	EnableExampleConstants              bool
	EnableExampleVariants               bool
	EnableExampleClient                 bool
	EnableExampleMoves                  bool
	//SuppressSeating turns off multiplayer seating. Seating is on by default:
	//without it the server has no way to put a real user into a player slot,
	//and a creator who accepts every default would get a game that compiles,
	//boots, serves, and silently seats nobody. Set this only for a game that
	//genuinely has no seats to fill (a solitaire or hotseat scaffold).
	SuppressSeating bool
	//EnableSeatPlayer and EnableInactivePlayer are what the templates read.
	//Validate derives them from SuppressSeating, so callers normally leave
	//them alone.
	EnableSeatPlayer     bool
	EnableInactivePlayer bool
}

// FileContents is the generated contents of the files to later write to the
// filesystem.
type FileContents map[string][]byte

// SuppressClient sets the options about clients to suppressed.
func (o *Options) SuppressClient() {
	o.SuppressClientRenderGame = true
	o.SuppressClientRenderPlayerInfo = true
}

// SuppressExtras sets all of the non-client extras that are on by default to
// off. That includes seating: seating needs a SetUp phase to seat into, and
// SuppressExtras turns phases off.
func (o *Options) SuppressExtras() {
	o.SuppressTest = true
	o.SuppressPhase = true
	o.SuppressCurrentPlayer = true
	o.SuppressMovesStubs = true
	o.SuppressComponentsStubs = true
	o.SuppressSeating = true
}

// EnableTutorials enables all of the off-by-default tutorial options. Seating
// is deliberately not among them: it is not tutorial content, it is the thing
// that makes the game multiplayer at all, so it is on by default and turned off
// with SuppressSeating.
func (o *Options) EnableTutorials() {
	o.EnableExampleDeck = true
	o.EnableExampleDynamicComponentValues = true
	o.EnableExampleEndState = true
	o.EnableExampleComputedProperties = true
	o.EnableExampleConstants = true
	o.EnableExampleVariants = true
	o.EnableExampleClient = true
	o.EnableExampleMoves = true
}

func nameLegal(gameName string) bool {
	matches, err := regexp.Match("^[a-zA-Z0-9]*$", []byte(gameName))

	if err != nil {
		log.Println("Regexmp illegal: " + err.Error())
		return false
	}

	return matches
}

// Validate verifies that Options is in a legal state. Makes sure Name exists
// and ensures it's lowerCase. Repeated calls are OK.
func (o *Options) Validate() error {

	o.Name = strings.TrimSpace(o.Name)
	o.Name = strings.ToLower(o.Name)

	if o.Name == "" {
		return errors.New("No name provided")
	}

	if !nameLegal(o.Name) {
		return errors.New("illegal characters in gamename. Must be a valid go package name")
	}

	if o.MinNumPlayers != 0 && o.MaxNumPlayers != 0 {
		if o.MaxNumPlayers < o.MinNumPlayers {
			return errors.New("Max num players less than min")
		}

		if o.DefaultNumPlayers != 0 {
			if o.DefaultNumPlayers < o.MinNumPlayers || o.DefaultNumPlayers > o.MaxNumPlayers {
				return errors.New("Default num players not within min/max range")
			}
		}
	}

	//We don't verify that the name is fully legal according to the boardgame
	//framework, because that test will fail given the test generated in
	//main_test.go.

	//Disallow illegal combinations
	if o.EnableExampleDynamicComponentValues {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleEndState {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleComputedProperties {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleConstants {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleClient {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleMoves {
		o.EnableExampleDeck = true
	}

	if o.EnableExampleMoves {
		o.SuppressPhase = false
	}

	if o.EnableExampleClient {
		o.SuppressClientRenderGame = false
		o.SuppressClientRenderPlayerInfo = false
	}

	if o.EnableExampleDeck {
		o.SuppressComponentsStubs = false
	}

	//Seating is on unless explicitly suppressed. SeatPlayer marks the players
	//it seats inactive, so the two behaviors always ship together.
	if !o.SuppressSeating {
		o.EnableSeatPlayer = true
		o.EnableInactivePlayer = true
	}

	if o.EnableInactivePlayer {
		o.EnableSeatPlayer = true
	}

	if o.EnableSeatPlayer {
		o.SuppressPhase = false
	}

	return nil
}

var spaceReducer *regexp.Regexp
var titleCaseReplacer *strings.Replacer

// Generate generates FileContents for the given set of options. A convenience
// wrapper around DefaultTemplateSet, templates.Generate(), and files.Format().
func Generate(opt *Options) (FileContents, error) {

	if err := opt.Validate(); err != nil {
		return nil, errors.New("Options didn't validate: " + err.Error())
	}

	templates, err := DefaultTemplateSet(opt)

	if err != nil {
		return nil, errors.New("Default Template Set errored: " + err.Error())
	}

	if templates == nil {
		return nil, errors.New("No templates returned")
	}

	files, err := templates.Generate(opt)

	if err != nil {
		return nil, errors.New("Couldn't generate file contents: " + err.Error())
	}

	if err := files.Format(); err != nil {
		return nil, errors.New("Couldn't go fmt generated file contents: " + err.Error())
	}

	return files, nil
}

// Format go formats all of the code om FileContents whose path ends in ".go",
// erroring if the code isn't valid. If an error is returned, then the contents
// of FileContents will not have been modified.
func (f FileContents) Format() error {

	newContent := make(map[string][]byte)

	for filename, rawSource := range f {
		if strings.ToLower(filepath.Ext(filename)) != ".go" {
			continue
		}

		transformedSource, err := format.Source(rawSource)

		if err != nil {
			return errors.New("Couldn't format go code for " + filename + ": " + err.Error())
		}

		newContent[filename] = transformedSource
	}

	for name, content := range newContent {
		f[name] = content
	}

	return nil
}

// Save saves the given FileContents to the filesystem, creating any implied
// directories. Dir is the prefix to join with each path in FileContents; "" is
// fine. Will error if overwite is not true and any of the files to create
// already exist.
func (f FileContents) Save(dir string, overwrite bool) error {
	return fileutil.WriteFilesAtomic(dir, f, overwrite, 0o644)
}
