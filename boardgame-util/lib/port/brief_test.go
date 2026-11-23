package port

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	root, err := transpile.ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func loadExample(t *testing.T, game string) *Game {
	t.Helper()
	root := moduleRoot(t)
	definition, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(transpile.DefinitionsDir), game+".json"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := Load(root, GamePackage(game), definition, "00000000")
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// TestBriefDeterministic: a brief is byte-identical across two runs, each
// from a freshly loaded package (the proposal's acceptance for the brief
// writer), for memory's refused function and checkers' predicate
// constructor; and it carries every section of the template.
func TestBriefDeterministic(t *testing.T) {
	cases := []struct{ game, ref string }{
		{"memory", "memory.(*moveStartHideCardsTimer).Apply"},
		{"checkers", "checkers.(*gameDelegate).ConfigurePredicateConstructors"},
	}
	for _, c := range cases {
		var texts [][]byte
		var sums []string
		for run := 0; run < 2; run++ {
			g := loadExample(t, c.game)
			brief, err := g.WriteBrief(c.ref, DefaultExamples(g.Root))
			if err != nil {
				t.Fatal(err)
			}
			texts = append(texts, brief.Text)
			sums = append(sums, brief.SHA256)
		}
		if !bytes.Equal(texts[0], texts[1]) || sums[0] != sums[1] {
			t.Errorf("%s: the brief differs between two runs", c.ref)
		}
		text := string(texts[0])
		for _, section := range []string{"\nFUNCTION\n", "\nGO SOURCE\n", "\nWHY THE TRANSLATOR REFUSED IT\n", "\nTYPES IT TOUCHES\n",
			"\nFUNCTIONS IT CALLS (already translated; call them, do not rewrite them)\n", "\nTHE TYPESCRIPT API YOU MAY USE\n",
			"\nSEMANTICS\n", "\nEXAMPLES\n", "\nOUTPUT\n", "\nHOW TO CHECK\n"} {
			if !strings.Contains(text, section) {
				t.Errorf("%s: the brief has no section %q", c.ref, strings.TrimSpace(section))
			}
		}
		if strings.Contains(text, "\n  Example 1: "+c.game+".") {
			t.Errorf("%s: an example is from the game itself", c.ref)
		}
	}
}

// TestMemoryBriefCarriesTheValue: memory's refused function reads a package
// variable assigned in ConfigureMoves; the brief must carry what the value
// comes from and the moves it can name, or the agent cannot know it.
func TestMemoryBriefCarriesTheValue(t *testing.T) {
	g := loadExample(t, "memory")
	brief, err := g.WriteBrief("memory.(*moveStartHideCardsTimer).Apply", DefaultExamples(g.Root))
	if err != nil {
		t.Fatal(err)
	}
	text := string(brief.Text)
	if out := os.Getenv("PORT_BRIEF_OUT"); out != "" {
		_ = os.WriteFile(out, brief.Text, 0644)
	}
	for _, want := range []string{
		"package-var-read memory.hideCardMoveName, at moves.go:114",
		"var hideCardMoveName string",
		"hideCardMoveName = hideCardConfig.Name()",
		"hideCardConfig := ",
		"\"Hide Cards\": memory.moveHideCards",
		"game.HideCardsTimer: {\"Type\":\"timer\"",
		"const hideCardsDuration = 4 * time.Second",
		"(boardgame.Timer).Start: timer.start",
		"(*boardgame.Game).MoveByName: moveByName",
		"export function memory_moveStartHideCardsTimer_Apply(state: RulesState, m: MoveInstance): string | null ... }",
		"| apply | MoveData.Bodies.Apply |",
		"boardgame-util port check memory.(*moveStartHideCardsTimer).Apply --file server/static/src/engine/games/memory/memory_moveStartHideCardsTimer_Apply.ts",
		"start(durationMs: number, move: MoveInstance",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the brief lacks %q", want)
		}
	}
}
