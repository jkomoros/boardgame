package gamedef_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance/corpus"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/examples/werewolf"
	"github.com/jkomoros/boardgame/server/static/src/engine/testdata/gen/vecgame"
	memstore "github.com/jkomoros/boardgame/storage/memory"
	"github.com/sirupsen/logrus"
)

// repoRoot is the repository root, relative to this package.
const repoRoot = "../../.."

type game struct {
	name        string
	importPath  string
	newDelegate func() boardgame.GameDelegate
}

// games are the seven example games, as boardgame-util gamedef export
// lists them.
var games = []game{
	{"memory", "github.com/jkomoros/boardgame/examples/memory", func() boardgame.GameDelegate { return memory.NewDelegate() }},
	{"blackjack", "github.com/jkomoros/boardgame/examples/blackjack", func() boardgame.GameDelegate { return blackjack.NewDelegate() }},
	{"pig", "github.com/jkomoros/boardgame/examples/pig", func() boardgame.GameDelegate { return pig.NewDelegate() }},
	{"tictactoe", "github.com/jkomoros/boardgame/examples/tictactoe", func() boardgame.GameDelegate { return tictactoe.NewDelegate() }},
	{"werewolf", "github.com/jkomoros/boardgame/examples/werewolf", func() boardgame.GameDelegate { return werewolf.NewDelegate() }},
	{"debuganimations", "github.com/jkomoros/boardgame/examples/debuganimations", func() boardgame.GameDelegate { return debuganimations.NewDelegate() }},
	{"checkers", "github.com/jkomoros/boardgame/examples/checkers", func() boardgame.GameDelegate { return checkers.NewDelegate() }},
}

func gameNamed(t *testing.T, name string) game {
	t.Helper()
	for _, g := range games {
		if g.name == name {
			return g
		}
	}
	t.Fatalf("no game %s", name)
	return game{}
}

func newManager(t *testing.T, g game) *boardgame.GameManager {
	t.Helper()
	manager, err := boardgame.NewGameManager(g.newDelegate(), memstore.NewStorageManager())
	if err != nil {
		t.Fatalf("%s: %v", g.name, err)
	}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
	return manager
}

// static is the static half, or nil while it is not implemented (as
// boardgame-util gamedef export tolerates).
func static(t *testing.T, g game) *gamedef.StaticInfo {
	t.Helper()
	info, err := gamedef.Static(g.importPath)
	if err != nil {
		if strings.Contains(err.Error(), "not implemented") {
			return nil
		}
		t.Fatalf("%s: static: %v", g.name, err)
	}
	return info
}

func export(t *testing.T, g game) *gamedef.GameData {
	t.Helper()
	data, err := gamedef.Export(newManager(t, g), static(t, g))
	if err != nil {
		t.Fatalf("%s: export: %v", g.name, err)
	}
	return data
}

func marshal(t *testing.T, data *gamedef.GameData) []byte {
	t.Helper()
	blob, err := gamedef.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return blob
}

// TestTestdataDefsCurrent keeps server/static/src/engine/testdata/defs/*.def.json
// equal, byte for byte, to the projection gen/defs writes.
func TestTestdataDefsCurrent(t *testing.T) {
	files, err := vecgame.DefFiles()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(repoRoot, "server", "static", "src", "engine", "testdata", "defs")
	onDisk, err := filepath.Glob(filepath.Join(dir, "*.def.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(onDisk) != len(files) {
		t.Errorf("%d def files checked in, %d generated", len(onDisk), len(files))
	}
	for _, file := range files {
		got, err := os.ReadFile(filepath.Join(dir, file.Name+".def.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, file.Blob) {
			t.Errorf("%s.def.json is stale; run ./scripts/go-local run ./server/static/src/engine/testdata/gen/defs", file.Name)
		}
	}
}

// TestDefinitionsCurrent is gamedef export --check.
func TestDefinitionsCurrent(t *testing.T) {
	dir := filepath.Join(repoRoot, "conformance", "definitions")
	for _, g := range games {
		got, err := os.ReadFile(filepath.Join(dir, g.name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, marshal(t, export(t, g))) {
			t.Errorf("conformance/definitions/%s.json is stale; run boardgame-util gamedef export", g.name)
		}
	}
	onDisk, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	if len(onDisk) != len(games) {
		t.Errorf("%d definitions checked in; want %d", len(onDisk), len(games))
	}
}

func TestExportDeterministic(t *testing.T) {
	for _, g := range games {
		first := marshal(t, export(t, g))
		second := marshal(t, export(t, g))
		if !bytes.Equal(first, second) {
			t.Errorf("%s: two exports differ", g.name)
		}
		var decoded gamedef.GameData
		if err := json.Unmarshal(first, &decoded); err != nil {
			t.Errorf("%s: does not round-trip: %v", g.name, err)
		}
	}
}

// TestNonNullCollections checks the schema's nullability rule.
func TestNonNullCollections(t *testing.T) {
	for _, g := range games {
		var doc map[string]any
		if err := json.Unmarshal(marshal(t, export(t, g)), &doc); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"ComponentEnums", "Initial", "Traits", "Computed", "LegalTemplates", "Predicates", "Moves", "SetUp", "End", "Agents", "Functions", "Definition"} {
			if doc[key] == nil {
				t.Errorf("%s: %s is null", g.name, key)
			}
		}
		for _, raw := range doc["Moves"].([]any) {
			move := raw.(map[string]any)
			for _, key := range []string{"Fields", "Plan", "Bodies"} {
				if move[key] == nil {
					t.Errorf("%s %s: %s is null", g.name, move["Name"], key)
				}
			}
			if move["Plan"].(map[string]any)["Specs"] == nil {
				t.Errorf("%s %s: Plan.Specs is null", g.name, move["Name"])
			}
			if stock, ok := move["Stock"].(map[string]any); ok && stock["Config"] == nil {
				t.Errorf("%s %s: Stock.Config is null", g.name, move["Name"])
			}
		}
	}
}

// header is the part of a conformance document this test compares.
type header struct {
	Game struct {
		Name              string `json:"name"`
		DisplayName       string `json:"displayName"`
		MinNumPlayers     int    `json:"minNumPlayers"`
		MaxNumPlayers     int    `json:"maxNumPlayers"`
		DefaultNumPlayers int    `json:"defaultNumPlayers"`
	} `json:"game"`
	Engine struct {
		PhaseEnum string `json:"phaseEnum"`
		GroupEnum string `json:"groupEnum"`
	} `json:"engine"`
	Moves                      []conformance.MoveType                 `json:"moves"`
	MoveInputSchema            []boardgame.MoveInputSchemaMove        `json:"moveInputSchema"`
	MoveChoiceProjectionSchema []json.RawMessage                      `json:"moveChoiceProjectionSchema"`
	ComputedProperties         []boardgame.ComputedPropertyDescriptor `json:"computedProperties"`
	VariantConfig              json.RawMessage                        `json:"variantConfig"`
	Chest                      struct {
		Decks map[string][]struct {
			Index  int
			Values json.RawMessage
		}
		Enums map[string]struct {
			Values       map[string]string
			DefaultValue int
			Parents      map[string]int
			Dimensions   []int
		}
		Constants      json.RawMessage
		LegalTemplates map[string]string
	} `json:"chest"`
}

// corpusHeader reads the header of the game's shortest checked-in document.
func corpusHeader(t *testing.T, name string) (*header, string) {
	t.Helper()
	blob, err := os.ReadFile(filepath.Join(repoRoot, corpus.Dir, corpus.IndexFile))
	if err != nil {
		t.Fatal(err)
	}
	var index []corpus.IndexEntry
	if err := json.Unmarshal(blob, &index); err != nil {
		t.Fatal(err)
	}
	file, versions := "", 0
	for _, entry := range index {
		if entry.Game == name && (file == "" || entry.Versions < versions) {
			file, versions = entry.File, entry.Versions
		}
	}
	if file == "" {
		t.Fatalf("no conformance document for %s", name)
	}
	compressed, err := os.ReadFile(filepath.Join(repoRoot, corpus.Dir, filepath.FromSlash(file)))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := corpus.Decompress(compressed)
	if err != nil {
		t.Fatal(err)
	}
	var h header
	if err := json.Unmarshal(raw, &h); err != nil {
		t.Fatal(err)
	}
	return &h, file
}

func canonical(t *testing.T, v any) string {
	t.Helper()
	blob, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(blob))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// propertyTypeName is the header's spelling ("TypePlayerIndex") of a
// GameData field or computed type ("playerIndex").
func propertyTypeName(t *testing.T, typeName string) string {
	t.Helper()
	for pt := boardgame.PropertyType(0); pt < 64; pt++ {
		if gamedef.TypeName(pt) == typeName && typeName != "" {
			return pt.String()
		}
	}
	t.Fatalf("no property type %q", typeName)
	return ""
}

// allFields is a move's fields including what its base supplies.
func allFields(move gamedef.MoveData) map[string]gamedef.FieldData {
	result := map[string]gamedef.FieldData{}
	for name, field := range move.Fields {
		result[name] = field
	}
	if move.Base != gamedef.BaseDefault {
		result["TargetPlayerIndex"] = gamedef.FieldData{Type: "playerIndex", Input: string(boardgame.MoveInputContextOwned), Codec: string(boardgame.MoveInputCodecPlayerIndex)}
	}
	return result
}

// TestAgreesWithConformanceHeader checks every field a conformance document
// header already carries against the export, for the six corpus games.
func TestAgreesWithConformanceHeader(t *testing.T) {
	for _, g := range games {
		if g.name == "checkers" {
			continue
		}
		t.Run(g.name, func(t *testing.T) {
			h, file := corpusHeader(t, g.name)
			data := export(t, g)
			t.Logf("compared with %s", file)

			if h.Game.Name != data.Name || h.Game.DisplayName != data.DisplayName || h.Game.MinNumPlayers != data.MinNumPlayers ||
				h.Game.MaxNumPlayers != data.MaxNumPlayers || h.Game.DefaultNumPlayers != data.DefaultNumPlayers {
				t.Errorf("game: header %+v; export %s %s %d %d %d", h.Game, data.Name, data.DisplayName, data.MinNumPlayers, data.MaxNumPlayers, data.DefaultNumPlayers)
			}
			if h.Engine.PhaseEnum != data.PhaseEnum || h.Engine.GroupEnum != data.GroupEnum {
				t.Errorf("enums: header %+v; export %q %q", h.Engine, data.PhaseEnum, data.GroupEnum)
			}

			// Move table, in registration order.
			if len(h.Moves) != len(data.Moves) {
				t.Fatalf("%d moves in the header, %d exported", len(h.Moves), len(data.Moves))
			}
			fields := 0
			for i, want := range h.Moves {
				got := data.Moves[i]
				if want.Name != got.Name || want.HelpText != got.HelpText || want.IsFixUp != got.IsFixUp || want.IsGatheringStart != got.GatheringStart {
					t.Errorf("move %d: header %q %q fixup=%v gathering=%v; export %q %q fixup=%v gathering=%v", i, want.Name, want.HelpText, want.IsFixUp, want.IsGatheringStart, got.Name, got.HelpText, got.IsFixUp, got.GatheringStart)
				}
				keys := []string{got.Name}
				if got.HiddenAnimationKey != "" {
					keys = append(keys, got.HiddenAnimationKey)
				}
				if !reflect.DeepEqual(want.AnimationKeys, keys) {
					t.Errorf("%s: animation keys %v; export %v", got.Name, want.AnimationKeys, keys)
				}
				if want.NamePublic != (got.NameSanitization == "") {
					t.Errorf("%s: namePublic %v; export NameSanitization %q", got.Name, want.NamePublic, got.NameSanitization)
				}
				exported := allFields(got)
				if len(want.Fields) != len(exported) {
					t.Errorf("%s: %d fields in the header, %d exported", got.Name, len(want.Fields), len(exported))
				}
				for _, field := range want.Fields {
					mine, ok := exported[field.Name]
					if !ok {
						t.Errorf("%s: field %s not exported", got.Name, field.Name)
						continue
					}
					if propertyTypeName(t, mine.Type) != field.Type || mine.Enum != field.EnumName {
						t.Errorf("%s.%s: header %s %q; export %s %q", got.Name, field.Name, field.Type, field.EnumName, mine.Type, mine.Enum)
					}
					fields++
				}
			}

			// Input schema: every non-fix-up move, every field.
			byName := map[string]gamedef.MoveData{}
			proposable := 0
			for _, move := range data.Moves {
				byName[move.Name] = move
				if !move.IsFixUp {
					proposable++
				}
			}
			if len(h.MoveInputSchema) != proposable {
				t.Errorf("%d moves in the input schema; %d non-fix-up moves exported", len(h.MoveInputSchema), proposable)
			}
			inputs := 0
			for _, schema := range h.MoveInputSchema {
				move, ok := byName[schema.Name]
				if !ok || move.IsFixUp {
					t.Errorf("input schema move %s is not an exported non-fix-up move", schema.Name)
					continue
				}
				exported := allFields(move)
				if len(schema.Fields) != len(exported) {
					t.Errorf("%s: %d input fields, %d exported", schema.Name, len(schema.Fields), len(exported))
				}
				for _, field := range schema.Fields {
					mine := exported[field.Name]
					if field.WireType != mine.Type || field.Disposition != mine.Input || field.Codec != mine.Codec || field.EnumName != mine.Enum {
						t.Errorf("%s.%s: schema %+v; export %+v", schema.Name, field.Name, field, mine)
					}
					if len(field.EnumValues) > 0 {
						def, ok := data.Definition.Enums[mine.Enum]
						if !ok {
							def, ok = data.ComponentEnums[mine.Enum]
						}
						var values []string
						for _, v := range def.Values {
							values = append(values, v)
						}
						sort.Strings(values)
						if !ok || !reflect.DeepEqual(values, field.EnumValues) {
							t.Errorf("%s.%s: enum values %v; exported enum %v", schema.Name, field.Name, field.EnumValues, values)
						}
					}
					inputs++
				}
			}

			// Choice projections.
			withChoices := 0
			for _, move := range data.Moves {
				if move.Choices != nil {
					withChoices++
				}
			}
			if len(h.MoveChoiceProjectionSchema) != withChoices {
				t.Errorf("%d choice projections in the header; %d moves export Choices", len(h.MoveChoiceProjectionSchema), withChoices)
			}
			for _, raw := range h.MoveChoiceProjectionSchema {
				var entry struct {
					MoveName string `json:"moveName"`
				}
				if err := json.Unmarshal(raw, &entry); err != nil {
					t.Fatal(err)
				}
				if got := byName[entry.MoveName].Choices; canonical(t, json.RawMessage(got)) != canonical(t, raw) {
					t.Errorf("%s: choices %s; export %s", entry.MoveName, raw, got)
				}
			}

			// Computed descriptors.
			if len(h.ComputedProperties) != len(data.Computed) {
				t.Errorf("%d computed properties in the header, %d exported", len(h.ComputedProperties), len(data.Computed))
			}
			for _, want := range h.ComputedProperties {
				found := false
				for _, got := range data.Computed {
					if got.Name == want.Name && got.Scope == string(want.Scope) && got.Type == gamedef.TypeName(want.Type) {
						found = true
					}
				}
				if !found {
					t.Errorf("computed %+v not exported (%+v)", want, data.Computed)
				}
			}

			// Variants.
			if canonical(t, h.VariantConfig) != canonical(t, data.Variants) {
				t.Errorf("variants: header %s; export %s", h.VariantConfig, data.Variants)
			}

			// Chest: decks, enums, constants, templates.
			if len(h.Chest.Decks) != len(data.Definition.Decks) {
				t.Errorf("%d decks in the header, %d exported", len(h.Chest.Decks), len(data.Definition.Decks))
			}
			components := 0
			for name, deck := range h.Chest.Decks {
				def := data.Definition.Decks[name]
				if len(deck) != len(def.Components) || def.Size != len(deck) {
					t.Errorf("deck %s: %d components, export %d (size %d)", name, len(deck), len(def.Components), def.Size)
					continue
				}
				for i, c := range deck {
					if c.Index != i || canonical(t, c.Values) != canonical(t, def.Components[i]) {
						t.Errorf("deck %s component %d: header %s; export %s", name, i, c.Values, def.Components[i])
					}
					components++
				}
			}
			if len(h.Chest.Enums) != len(data.Definition.Enums) {
				t.Errorf("%d enums in the header, %d exported", len(h.Chest.Enums), len(data.Definition.Enums))
			}
			for name, want := range h.Chest.Enums {
				got := data.Definition.Enums[name]
				if !reflect.DeepEqual(want.Values, got.Values) || want.DefaultValue != got.DefaultValue ||
					!reflect.DeepEqual(want.Parents, got.Parents) || !reflect.DeepEqual(want.Dimensions, got.Range) {
					t.Errorf("enum %s: header %+v; export %+v", name, want, got)
				}
			}
			if canonical(t, h.Chest.Constants) != canonical(t, data.Constants) {
				t.Errorf("constants: header %s; export %v", h.Chest.Constants, data.Constants)
			}
			for key, text := range data.LegalTemplates {
				if h.Chest.LegalTemplates[key] != text {
					t.Errorf("template %s: header %q; export %q", key, h.Chest.LegalTemplates[key], text)
				}
			}
			for key := range h.Chest.LegalTemplates {
				if _, ok := data.LegalTemplates[key]; !ok && !strings.HasPrefix(key, "legal.") {
					t.Errorf("template %s is in the merged table but neither exported nor a framework default", key)
				}
			}
			t.Logf("%d moves, %d fields, %d input fields, %d choice projections, %d computed, %d components, %d enums, %d game templates",
				len(data.Moves), fields, inputs, withChoices, len(data.Computed), components, len(data.Definition.Enums), len(data.LegalTemplates))
		})
	}
}

func TestInitial(t *testing.T) {
	ttt := export(t, gameNamed(t, "tictactoe"))
	if len(ttt.Initial.Players) != ttt.MaxNumPlayers {
		t.Fatalf("tictactoe: %d initial players; max %d", len(ttt.Initial.Players), ttt.MaxNumPlayers)
	}
	for i, want := range []string{"X", "O"} {
		got := ttt.Initial.Players[i]
		if got["TokenValue"] != want || got["TokensToPlaceThisTurn"] != 1 || len(got) != 2 {
			t.Errorf("tictactoe player %d: %v", i, got)
		}
	}
	if len(ttt.Initial.Game) != 0 || len(ttt.Initial.Dynamic) != 0 {
		t.Errorf("tictactoe: game %v, dynamic %v", ttt.Initial.Game, ttt.Initial.Dynamic)
	}

	wolf := export(t, gameNamed(t, "werewolf"))
	if len(wolf.Initial.Players) != wolf.MaxNumPlayers {
		t.Fatalf("werewolf: %d initial players; max %d", len(wolf.Initial.Players), wolf.MaxNumPlayers)
	}
	for i, got := range wolf.Initial.Players {
		if got["DayVote"] != -1 || got["NightVote"] != -1 || len(got) != 2 {
			t.Errorf("werewolf player %d: %v", i, got)
		}
	}

	piggy := export(t, gameNamed(t, "pig"))
	if !reflect.DeepEqual(piggy.Initial.Game, map[string]any{"TargetScore": 100}) {
		t.Errorf("pig game: %v", piggy.Initial.Game)
	}
	if dice := piggy.Initial.Dynamic["dice"]; len(dice) != 1 || dice[0]["Value"] != 1 {
		t.Errorf("pig dynamic: %v", piggy.Initial.Dynamic)
	}
}

// TestDefaultNumPlayersAsGoUsesIt: memory and debuganimations declare
// DefaultNumPlayeres (misspelled), so base.GameDelegate's default applies.
func TestDefaultNumPlayersAsGoUsesIt(t *testing.T) {
	baseDefault := (&base.GameDelegate{}).DefaultNumPlayers()
	for _, name := range []string{"memory", "debuganimations"} {
		g := gameNamed(t, name)
		delegate := g.newDelegate()
		if _, ok := reflect.TypeOf(delegate).MethodByName("DefaultNumPlayeres"); !ok {
			t.Fatalf("%s no longer declares DefaultNumPlayeres; this test is moot", name)
		}
		data := export(t, g)
		if data.DefaultNumPlayers != baseDefault || data.DefaultNumPlayers != delegate.DefaultNumPlayers() {
			t.Errorf("%s: DefaultNumPlayers %d; base default %d", name, data.DefaultNumPlayers, baseDefault)
		}
	}
}

func TestCheckers(t *testing.T) {
	data := export(t, gameNamed(t, "checkers"))
	def := data.Definition

	spaces, ok := def.Enums["spaces"]
	if !ok || !reflect.DeepEqual(spaces.Range, []int{8, 8}) || len(spaces.Values) != 64 {
		t.Errorf("spaces enum: %+v", spaces)
	}
	if got := def.Player.Props["CapturedTokens"].Constraints; !reflect.DeepEqual(got, []string{"same(Color)"}) {
		t.Errorf("CapturedTokens constraints: %v", got)
	}
	if got := def.Game.Props["UnusedTokens"].Constraints; got != nil {
		t.Errorf("UnusedTokens constraints: %v", got)
	}
	// The board: a sized stack with one slot per key of the spaces enum,
	// sized by the BOARD_SIZE constant.
	board := def.Game.Props["Spaces"]
	if board.Type != "sizedStack" || board.Size != len(spaces.Values) || data.Constants["BOARD_SIZE"] != board.Size {
		t.Errorf("Spaces: %+v; constants %v", board, data.Constants)
	}

	tokens := def.Decks["Tokens"]
	dynamic, ok := def.DynamicComponentValues["Tokens"]
	if !ok || dynamic.Props["Crowned"].Type != "bool" {
		t.Errorf("Tokens dynamic values: %+v", dynamic)
	}
	if got := data.Initial.Dynamic["Tokens"]; len(got) != tokens.Size || tokens.Size == 0 {
		t.Errorf("Tokens initial dynamic values: %d for %d components", len(got), tokens.Size)
	} else {
		for i, values := range got {
			if len(values) != 0 {
				t.Errorf("Tokens component %d initial values %v", i, values)
			}
		}
	}
	if props := tokens.ComponentProps; props["Color"].Type != "enum" || props["Color"].Enum == "" {
		t.Errorf("Tokens component props: %+v", props)
	}

	moved := false
	for _, move := range data.Moves {
		if move.Name != "Move Token" {
			continue
		}
		moved = true
		for _, field := range []string{"TokenIndexToMove", "SpaceIndex"} {
			if got := move.Fields[field]; got.Type != "enum" || got.Enum != "spaces" || got.Codec != "integer" {
				t.Errorf("Move Token %s: %+v", field, got)
			}
		}
		if !move.Plan.Declared || len(move.Plan.Specs) == 0 {
			t.Errorf("Move Token plan: %+v", move.Plan)
		}
	}
	if !moved {
		t.Error("no Move Token move")
	}

	ttt := export(t, gameNamed(t, "tictactoe"))
	if got := ttt.Definition.Player.Props["UnusedTokens"].Constraints; !reflect.DeepEqual(got, []string{"same(Value)"}) {
		t.Errorf("tictactoe UnusedTokens constraints: %v", got)
	}
}

func TestDefinitionAdditions(t *testing.T) {
	jack := export(t, gameNamed(t, "blackjack"))
	if _, ok := jack.ComponentEnums["Suit"]; !ok {
		t.Errorf("blackjack component enums: %v", jack.ComponentEnums)
	}
	if props := jack.Definition.Decks["cards"].ComponentProps; props["Suit"].Enum != "Suit" || props["Rank"].Enum != "Rank" {
		t.Errorf("blackjack component props: %v", props)
	}
	if jack.Definition.Decks["cards"].GenericValues == nil {
		t.Error("blackjack has no generic values")
	}
	if !jack.Traits.Game.RoundRobin || jack.Traits.Game.DrawDiscardPair == nil || *jack.Traits.Game.DrawDiscardPair != (gamedef.DrawDiscardPair{Draw: "DrawStack", Discard: "DiscardStack"}) ||
		!jack.Traits.Player.Seat || !jack.Traits.Player.Elimination || jack.Traits.Player.GameScore != gamedef.GameScoreFromScore {
		t.Errorf("blackjack traits: %+v", jack.Traits)
	}

	mem := export(t, gameNamed(t, "memory"))
	if mem.Definition.Decks["cards"].GenericValues == nil {
		t.Error("memory has no generic values")
	}
	anim := export(t, gameNamed(t, "debuganimations"))
	if parents := anim.Definition.Enums["phase"].Parents; len(parents) == 0 {
		t.Error("debuganimations phase tree enum has no parents")
	}
	if tokens := anim.Definition.Decks["tokens"]; tokens.ComponentProps != nil {
		t.Errorf("debuganimations tokens have component props %v", tokens.ComponentProps)
	}
}

func TestMoves(t *testing.T) {
	jack := export(t, gameNamed(t, "blackjack"))
	progressions := 0
	for _, move := range jack.Moves {
		if move.Progression != nil {
			progressions++
		}
		switch move.Name {
		case "Begin Game":
			if move.Stock == nil || move.Stock.Type != "StartPhase" || move.Stock.Config["github.com/jkomoros/boardgame/moves.StartPhaseEnum"] != "phase" ||
				move.Stock.Config["github.com/jkomoros/boardgame/moves.StartPhase"] != int64(1) || !move.EndsPhaseProgression || move.GoType != "moves.StartPhase" {
				t.Errorf("Begin Game: %+v %+v", move, move.Stock)
			}
		case "Current Player Hit":
			if move.Base != gamedef.BaseCurrentPlayer || move.Stock == nil || move.Stock.Type != "DrawToPlayer" || move.GoType != "blackjack.moveCurrentPlayerHit" || !move.Plan.HasCustom || !reflect.DeepEqual(move.LegalPhases, []int{2}) {
				t.Errorf("Current Player Hit: %+v", move)
			}
		case "Seat Player":
			if !move.SeatsPlayer || move.Base != gamedef.BaseDefault || move.Fields["TargetPlayerIndex"].Type != "playerIndex" {
				t.Errorf("Seat Player: %+v", move)
			}
		}
	}
	if progressions != 15 {
		t.Errorf("blackjack: %d moves with a progression; want 15 (as moves.DescribeProgression's tests count)", progressions)
	}

	ttt := export(t, gameNamed(t, "tictactoe"))
	for _, move := range ttt.Moves {
		if move.Name == "Place Token" {
			if move.NameSanitization != "self:visible" || move.HiddenAnimationKey != "Hidden Action" || move.Fields["Slot"].Sanitize != "self:visible" || move.Base != gamedef.BaseCurrentPlayer {
				t.Errorf("Place Token: %+v", move)
			}
		}
	}
	wolf := export(t, gameNamed(t, "werewolf"))
	for _, move := range wolf.Moves {
		if move.Name == "Cast Vote" && move.Base != gamedef.BaseAnyPlayer {
			t.Errorf("Cast Vote base: %s", move.Base)
		}
	}
}

// fakeStatic is a StaticInfo agreeing with data's run-time half.
func fakeStatic(data *gamedef.GameData) *gamedef.StaticInfo {
	info := &gamedef.StaticInfo{
		Package:    data.Source.Package,
		Moves:      map[string]gamedef.MoveStatic{},
		Computed:   map[string]gamedef.FnRef{},
		Predicates: map[string]gamedef.FnRef{},
		Traits:     data.Traits,
		SetUp:      gamedef.SetUpData{Distribute: "x.(*gameDelegate).DistributeComponentToStarterStack"},
		Functions:  []gamedef.FunctionEntry{{Func: "x.b"}, {Func: "x.a"}},
	}
	for _, move := range data.Moves {
		static := gamedef.MoveStatic{GoType: move.GoType}
		if move.Stock != nil {
			static.StockType = move.Stock.Type
		}
		if !strings.HasPrefix(move.GoType, "moves.") {
			static.Bodies.Apply = fmt.Sprintf("x.(*%s).Apply", strings.TrimPrefix(move.GoType, "x."))
		}
		info.Moves[move.GoType] = static
	}
	for _, c := range data.Computed {
		info.Computed[c.Name] = "x.computed" + c.Name
	}
	return info
}

func TestStaticJoin(t *testing.T) {
	g := gameNamed(t, "blackjack")
	runtime, err := gamedef.Export(newManager(t, g), nil)
	if err != nil {
		t.Fatal(err)
	}
	info := fakeStatic(runtime)
	joined, err := gamedef.Export(newManager(t, g), info)
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	if joined.SetUp.Distribute == "" || len(joined.Functions) != 2 || joined.Functions[0].Func != "x.a" {
		t.Errorf("static parts not joined: %+v %+v", joined.SetUp, joined.Functions)
	}
	for _, move := range joined.Moves {
		if !strings.HasPrefix(move.GoType, "moves.") && move.Bodies.Apply == "" {
			t.Errorf("%s: bodies not joined", move.Name)
		}
	}
	if joined.Computed[0].Compute == "" {
		t.Error("computed FnRef not joined")
	}

	// A move type the static half does not know.
	missing := fakeStatic(runtime)
	delete(missing.Moves, "blackjack.moveCurrentPlayerHit")
	if _, err := gamedef.Export(newManager(t, g), missing); err == nil {
		t.Error("export succeeded without a static entry for a move")
	}
	// A disagreeing stock ancestor.
	disagree := fakeStatic(runtime)
	entry := disagree.Moves["blackjack.moveCurrentPlayerHit"]
	entry.StockType = "CurrentPlayer"
	disagree.Moves["blackjack.moveCurrentPlayerHit"] = entry
	if _, err := gamedef.Export(newManager(t, g), disagree); err == nil {
		t.Error("export succeeded with a disagreeing stock ancestor")
	}
	// Disagreeing traits.
	traits := fakeStatic(runtime)
	traits.Traits.Game.DrawDiscardPair = &gamedef.DrawDiscardPair{Draw: "A", Discard: "B"}
	if _, err := gamedef.Export(newManager(t, g), traits); err == nil {
		t.Error("export succeeded with a disagreeing DrawDiscardPair")
	}
	roundRobin := fakeStatic(runtime)
	roundRobin.Traits.Game.RoundRobin = false
	if _, err := gamedef.Export(newManager(t, g), roundRobin); err == nil {
		t.Error("export succeeded with a disagreeing RoundRobin")
	}
	// A declared GameScore replaces the behavior's.
	score := fakeStatic(runtime)
	score.Traits.Player.GameScore = "x.(*playerState).GameScore"
	joinedScore, err := gamedef.Export(newManager(t, g), score)
	if err != nil || joinedScore.Traits.Player.GameScore != "x.(*playerState).GameScore" {
		t.Errorf("declared GameScore: %v %v", err, joinedScore)
	}
}

func TestWithoutAdditionsLeavesOriginal(t *testing.T) {
	data := export(t, gameNamed(t, "checkers"))
	before := canonical(t, data.Definition)
	projected := data.Definition.WithoutAdditions()
	if canonical(t, data.Definition) != before {
		t.Error("WithoutAdditions modified its receiver")
	}
	blob := canonical(t, projected)
	for _, addition := range []string{"GenericValues", "ComponentProps", "Parents", "Range", "Constraints"} {
		if strings.Contains(blob, strconv.Quote(addition)) {
			t.Errorf("projection still has %s", addition)
		}
	}
}
