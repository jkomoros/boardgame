package stub

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

// If true, will save out the files generated. Useful for generating new golden
// output when output is changed. Flip to true, run `go test`, verify the diff
// looks right, and then flip this back to false before committing.
const generateNewGolden = false

// The go tool will ignore everything rooted in 'testdata'
const testDir = "testdata"

func TestBasicGenerate(t *testing.T) {

	opt := &Options{
		Name: "checkers",
	}

	tmpls, err := DefaultTemplateSet(opt)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(tmpls)).DoesNotEqual(0)

	contents, err := tmpls.Generate(opt)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(contents)).DoesNotEqual(0)

	assert.For(t).ThatActual(contents["checkers/main.go"]).IsNotNil()
	assert.For(t).ThatActual(contents["checkers/client/boardgame-render-game-checkers.ts"]).IsNotNil()
	assert.For(t).ThatActual(contents["checkers/client/boardgame-render-player-info-checkers.ts"]).IsNotNil()
	if _, exists := contents["checkers/client/boardgame-render-game-checkers.js"]; exists {
		t.Fatal("legacy JavaScript renderer should not be generated")
	}
}

func TestExampleClientUsesWorkingTypedDefaults(t *testing.T) {
	opt := &Options{Name: "checkers"}
	opt.EnableTutorials()
	contents, err := Generate(opt)
	if err != nil {
		t.Fatal(err)
	}
	client := string(contents["checkers/client/boardgame-render-game-checkers.ts"])
	playerInfo := string(contents["checkers/client/boardgame-render-player-info-checkers.ts"])
	for _, required := range []string{
		"@registerGameRenderer",
		"cardView<GameState['DrawStack']>",
		"<boardgame-game-surface heading=\"Checkers\">",
		"<boardgame-component-zone",
		"<boardgame-game-outcome",
		"<boardgame-player-grid>",
		"<boardgame-player-panel",
		".active=${index === this.currentPlayerIndex}",
		"label=\"Draw pile\"",
		".componentView=${this.cards}",
		"component.Values.Value",
		".componentView=${this.cards.withProperties({ rotated: true })}",
		"<boardgame-action-button .action=${this.move(MoveNames.DrawCard)}>",
		"<boardgame-action-bar slot=\"actions\" label=\"Turn actions\">",
		"<boardgame-turn-status",
		".turn=${this.turnStatus}",
		"player.Computed?.RoundScore ?? 0",
	} {
		if !strings.Contains(client, required) {
			t.Errorf("generated client is missing %q", required)
		}
	}
	if !strings.Contains(playerInfo, "@registerPlayerInfoRenderer") {
		t.Error("generated player-info renderer must use exact generated registration")
	}
	if strings.Contains(client, "import '../../src/components/") {
		t.Error("generated renderer should register supported primitives through the public client facade")
	}
	if !strings.Contains(playerInfo, `<boardgame-stat label="Cards" .stack=${this.playerState?.Hand ?? null}>`) {
		t.Error("generated player info does not use the typed boardgame-stat stack API")
	}
	for _, legacy := range []string{"boardgame-deck-defaults", "{{item.", "component-rotated", "componentAttrs", "propose-move", "data-arg-", "this.proposeMove"} {
		if strings.Contains(client, legacy) {
			t.Errorf("generated client contains legacy authoring syntax %q", legacy)
		}
	}
}

func TestGolden(t *testing.T) {

	minimalOptions := &Options{
		//ensure we validate name
		Name: " Checkers",
	}

	minimalOptions.SuppressClient()
	minimalOptions.SuppressExtras()

	tutorialOptions := &Options{
		Name: "checkers",
		//Ensure that if the display name is not "" we output it.
		DisplayName: "CHECKERS!!!",
	}

	tutorialOptions.EnableTutorials()

	tests := map[string]*Options{
		"default": {
			Name:              "checkers",
			Description:       "A classic game for two players where you advance across the board, capturing the other player's pawns",
			MinNumPlayers:     2,
			MaxNumPlayers:     4,
			DefaultNumPlayers: 2,
		},
		"minimal":  minimalOptions,
		"tutorial": tutorialOptions,
	}

	if generateNewGolden {
		fmt.Println("Saving new golden. Before committing, flip generateNewGolden back to false.")
	}

	for name, opt := range tests {
		compareGolden(t, name, opt)
	}

}

func compareGolden(t *testing.T, name string, opt *Options) {

	contents, err := Generate(opt)

	assert.For(t, name).ThatActual(err).IsNil()

	dir := filepath.Join(testDir, name)

	if generateNewGolden {

		//Save out contents as new golden files to compare against
		contents.Save(dir, true)

		gameDir := filepath.Join(dir, opt.Name)

		cmd := exec.Command("go", "generate")
		cmd.Dir = gameDir
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			fmt.Println("Couldn't generate: " + err.Error())
			return
		}

		//Generated golden; now verify that the generated pass tests. We do
		//this now so that general tests will be fast; we verify that future
		//tests output the same thing, and then verify that the thing they
		//equal was valid when generated.
		cmd = exec.Command("go", "test")
		cmd.Dir = filepath.Join(dir, opt.Name)
		buf := &bytes.Buffer{}
		cmd.Stderr = buf

		if err := cmd.Run(); err != nil {
			fmt.Println("New package didn't pass test: " + name + ": " + err.Error())
			fmt.Println(buf.String())
			t.FailNow()
			return
		}

		return
	}

	golden, err := fileContentsFromDir(dir)

	assert.For(t, name).ThatActual(err).IsNil()

	assert.For(t, name).ThatActual(contents).Equals(golden).ThenDiffOnFail()

}

// TestTutorialStubBuilds is the build tripwire: it regenerates the tutorial
// stub into a scratch package, runs `go generate` over it and builds it, so
// that a change to the underlying libraries that leaves the stub uncompilable
// fails a test rather than waiting for the next person to run the stub.
//
// IT USED TO LIVE INSIDE compareGolden, AND THAT MADE THE TUTORIAL GOLDEN
// DEAD. The tripwire ended with `fmt.Println("Couldn't generate: ...")` and a
// bare `return` -- a `return` out of compareGolden, before the golden
// comparison at the bottom of it. And the generate could not do anything BUT
// fail: the scratch dir came from os.MkdirTemp(""), which is outside any Go
// module, so `go generate` there died on "go.mod file not found" on every run
// on every machine. Measured, in an isolated copy: appending a comment to
// testdata/tutorial/checkers/main.go left `go test` green, while the same
// change to testdata/default/checkers/main.go turned it red. The tutorial
// case is the one that exercises EnableTutorials(), i.e. the largest template
// surface in the package, and it had never compared anything.
//
// Two changes keep that from coming back. The comparison is no longer
// downstream of the build at all -- compareGolden does nothing but compare --
// and the scratch package is created INSIDE this package's directory, so it
// inherits the repo's go.mod and `go generate` can actually run.
func TestTutorialStubBuilds(t *testing.T) {

	//The generated package carries `//go:generate boardgame-util codegen`, so
	//the tripwire needs that binary. Skipping is loud (`go test -v` prints the
	//reason) and, unlike the old behaviour, it cannot take a golden
	//comparison down with it.
	if _, err := exec.LookPath("boardgame-util"); err != nil {
		t.Skip("boardgame-util is not on PATH, so `go generate` cannot run; " +
			"build and install it to exercise this tripwire")
	}

	opt := &Options{
		Name:        "checkers",
		DisplayName: "CHECKERS!!!",
	}
	opt.EnableTutorials()

	contents, err := Generate(opt)

	if err != nil {
		t.Fatal("couldn't generate: " + err.Error())
	}

	//Inside the package dir, NOT os.MkdirTemp(""): the generated package has
	//to sit within a module for `go generate` and `go build` to resolve
	//anything at all. The go tool ignores directories it cannot see as
	//packages until they are built, and this one is removed either way.
	tempDir, err := os.MkdirTemp(".", "TEMP_test_pkg_")

	if err != nil {
		t.Fatal("Couldn't create temp dir: " + err.Error())
	}

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatal("couldn't clean up temp testing dir: " + err.Error())
		}
	}()

	if err := contents.Save(tempDir, false); err != nil {
		t.Fatal("couldn't save contents: " + err.Error())
	}

	gameDir := filepath.Join(tempDir, opt.Name)

	cmd := exec.Command("go", "generate")
	cmd.Dir = gameDir
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal("`go generate` failed on the generated stub: " + err.Error() + ": " + string(out))
	}

	cmd = exec.Command("go", "build")
	cmd.Dir = gameDir
	out, err = cmd.CombinedOutput()

	if err != nil {
		t.Fatal("Didn't build (likely underlying library changed) " + err.Error() + ": " + string(out))
	}
}

// fileContentsFromDir loads up filecontents from the given path so they can be
// compared to the golden.
func fileContentsFromDir(path string) (FileContents, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New(path + " doesnt' exist")
	}

	result := make(FileContents)

	if err := recursiveListFilesForFileContents(path, "", result); err != nil {
		return nil, errors.New("couldn't list files: " + err.Error())
	}

	return result, nil

}

// basePath is actual dir to list recursively; prefix is the prefix to affix to
// dir contenst to put in contents.
func recursiveListFilesForFileContents(basePath, prefix string, contents FileContents) error {

	infos, err := os.ReadDir(basePath)

	if err != nil {
		return errors.New("Couldn't list path: " + err.Error())
	}

	for _, info := range infos {
		if info.IsDir() {
			if err := recursiveListFilesForFileContents(filepath.Join(basePath, info.Name()), filepath.Join(prefix, info.Name()), contents); err != nil {
				return err
			}
			continue
		}
		//info represents a file.

		//Skip auto-generated files
		if strings.HasPrefix(info.Name(), "auto_") && strings.HasSuffix(info.Name(), ".go") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(basePath, info.Name()))

		if err != nil {
			return errors.New("couldn't read " + filepath.Join(basePath, info.Name()) + ": " + err.Error())
		}

		contents[filepath.Join(prefix, info.Name())] = content
	}

	return nil

}
