package stub

import (
	"strings"
	"testing"
)

/*
The generated scaffold is the only renderer the framework itself writes, so it
is the one place where "the default arrives without being asked for" is either
true or false for every new game.

Two batteries exist precisely so renderers stop rebuilding them:

  - server/static/src/styles/renderer-styles.ts, wired onto
    BoardgameBaseGameRenderer.styles. A renderer that writes
    `static override styles = css\`...\`` REPLACES it. Every hand-written
    example renderer composes instead, with
    `...(GameRenderer.styles ? [GameRenderer.styles] : [])`; the generator did
    not, so a freshly scaffolded game got zero of .horizontal/.vertical/
    .center/.flex/.active/.responding/.selected/.targetable/.disabled/
    .eliminated -- while renderer-styles.ts and boardgame-base-game-renderer.ts
    both asserted, as load-bearing fact, that it did.

  - server/static/src/components/boardgame-stat.ts, which reads the count AND
    the capacity off the stack itself. The generator emitted
    `.value=${...Hand.Indexes.length}`, which is the exact hand-rolled shape
    that component was built to delete: on a sized stack Indexes is padded with
    a -1 sentinel at every empty slot, so its length is the CAPACITY, not the
    count. Teaching that idiom from the scaffold is how it spreads.
*/

func generatedClient(t *testing.T, tutorial bool) (gameRenderer, playerInfo string) {
	t.Helper()
	opt := &Options{Name: "checkers"}
	if tutorial {
		opt.EnableTutorials()
	}
	contents, err := Generate(opt)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents["checkers/client/boardgame-render-game-checkers.ts"]),
		string(contents["checkers/client/boardgame-render-player-info-checkers.ts"])
}

// TestGeneratedGameRendererComposesSharedStyles is the acceptance test for the
// styles battery: a scaffolded game must inherit the shared vocabulary rather
// than silently blowing it away with its own `static styles`.
func TestGeneratedGameRendererComposesSharedStyles(t *testing.T) {
	for _, tutorial := range []bool{false, true} {
		name := "default"
		if tutorial {
			name = "tutorial"
		}
		t.Run(name, func(t *testing.T) {
			client, _ := generatedClient(t, tutorial)
			if !strings.Contains(client, "...(GameRenderer.styles ? [GameRenderer.styles] : [])") {
				t.Errorf("generated game renderer replaces the shared style vocabulary instead of composing it; it must spread GameRenderer.styles first:\n%s", client)
			}
			if strings.Contains(client, "static override styles = css`") {
				t.Error("generated game renderer assigns css`` directly to static styles, which REPLACES the shared vocabulary")
			}
		})
	}
}

// TestGeneratedPlayerInfoUsesStat is the acceptance test for the stat battery:
// the scaffold must ask the stack for its own count, never count Indexes.
func TestGeneratedPlayerInfoUsesStat(t *testing.T) {
	_, playerInfo := generatedClient(t, true)
	if strings.Contains(playerInfo, "Indexes.length") {
		t.Error("generated player info counts Indexes.length, which is the capacity of a sized stack, not its count; use <boardgame-stat .stack=${...}>")
	}
	if !strings.Contains(playerInfo, "<boardgame-stat") {
		t.Errorf("generated player info does not use <boardgame-stat>, the component built for exactly this:\n%s", playerInfo)
	}
	if !strings.Contains(playerInfo, ".stack=${this.playerState?.Hand ?? null}") {
		t.Errorf("generated player info must hand the stat the stack itself:\n%s", playerInfo)
	}
}

// TestGeneratedClientNeverCountsIndexes guards the whole generated client
// surface, not just the one line that had the bug.
func TestGeneratedClientNeverCountsIndexes(t *testing.T) {
	for _, tutorial := range []bool{false, true} {
		client, playerInfo := generatedClient(t, tutorial)
		for _, contents := range []string{client, playerInfo} {
			if strings.Contains(contents, "Indexes.length") {
				t.Errorf("generated client counts Indexes.length; a sized stack pads Indexes with -1 sentinels:\n%s", contents)
			}
		}
	}
}
