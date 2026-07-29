package stub

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestDefaultOptionsGenerateSeating pins that a creator who accepts every
// default gets a game the server can actually seat users into.
//
// Seating used to be reachable only by answering yes to "Generate extra
// tutorial content that demonstrate how to wire up more things?" -- a prompt
// that defaults to no and whose wording gives no hint it controls seating. The
// result was a scaffold that compiles, boots, serves, and silently never seats
// anybody, with nothing saying so.
func TestDefaultOptionsGenerateSeating(t *testing.T) {

	contents, err := Generate(&Options{Name: "checkers"})
	if err != nil {
		t.Fatalf("generating: %v", err)
	}

	playerState := string(contents["checkers/player_state.go"])
	for _, required := range []string{"behaviors.Seat", "behaviors.InactivePlayer"} {
		if !strings.Contains(playerState, required) {
			t.Errorf("the default stub's playerState should embed %s; without it "+
				"the server has no notion of a seat to put a user in", required)
		}
	}

	main := string(contents["checkers/main.go"])
	if !strings.Contains(main, "moves.SeatPlayer") {
		t.Error("the default stub should configure moves.SeatPlayer; without it " +
			"the server cannot seat players and NumSeatedActivePlayers is always 0")
	}
}

// TestSuppressSeatingIsTheOnlyWayOut pins that seating is opt-out rather than
// opt-in, and that the opt-out still works for the games that genuinely have no
// seats to fill.
func TestSuppressSeatingIsTheOnlyWayOut(t *testing.T) {

	for description, opt := range map[string]*Options{
		"explicit SuppressSeating": {Name: "checkers", SuppressSeating: true},
		"SuppressExtras implies it": func() *Options {
			opt := &Options{Name: "checkers"}
			opt.SuppressExtras()
			return opt
		}(),
	} {
		contents, err := Generate(opt)
		if err != nil {
			t.Fatalf("%s: generating: %v", description, err)
		}

		if strings.Contains(string(contents["checkers/main.go"]), "moves.SeatPlayer") {
			t.Errorf("%s: should leave moves.SeatPlayer out of the stub", description)
		}
		if strings.Contains(string(contents["checkers/player_state.go"]), "behaviors.Seat") {
			t.Errorf("%s: should leave behaviors.Seat out of the playerState", description)
		}
	}
}

// TestSeatPlayerIsScopedToThePhaseThatActivatesPlayers is the second half of
// the scaffolding bug: even with seating enabled, the generated game seated
// late joiners into a permanent freeze.
//
// moves.SeatPlayer marks every player it seats inactive, and the only thing
// that undoes that is Optional(ActivateInactivePlayer), which lives inside
// moves.DefaultRoundSetup. The stub puts DefaultRoundSetup inside
// AddOrderedForPhase(phaseSetUp, ...) while registering SeatPlayer with a bare
// moves.Add -- legal in every phase. The generated game runs phaseSetUp exactly
// once and then starts phaseNormal forever, so anyone the server seated after
// that was inactivated and never activated again: a silent permanent spectator.
//
// The fix scopes SeatPlayer to the same phase that activates. That phase is
// where seating is supposed to happen anyway: the game starts in phaseSetUp
// (a tree enum's default value is its first real member, not the root) and
// WaitForEnoughPlayers holds it there until enough players are seated.
func TestSeatPlayerIsScopedToThePhaseThatActivatesPlayers(t *testing.T) {

	tutorialOptions := &Options{Name: "checkers"}
	tutorialOptions.EnableTutorials()

	for name, opt := range map[string]*Options{
		"default":  {Name: "checkers"},
		"tutorial": tutorialOptions,
	} {
		contents, err := Generate(opt)
		if err != nil {
			t.Fatalf("%s: generating: %v", name, err)
		}

		group, ok := seatPlayerRegistrationGroup(t, string(contents["checkers/main.go"]))
		if !ok {
			t.Errorf("%s: could not find a moves.Add* group registering moves.SeatPlayer", name)
			continue
		}

		if group != "AddForPhase(phaseSetUp)" {
			t.Errorf("%s: moves.SeatPlayer is registered via %s, so the server can "+
				"seat a player in a phase where no ActivateInactivePlayer will ever "+
				"run; that player is inactivated on seating and stays a permanent "+
				"spectator. Expected AddForPhase(phaseSetUp)", name, group)
		}
	}
}

// seatPlayerRegistrationGroup parses generated Go and reports which moves.Add*
// call registers moves.SeatPlayer, as e.g. "Add" or "AddForPhase(phaseSetUp)".
// Parsing rather than grepping matters here: the whole bug is about which
// grouping call encloses the registration, which a substring search cannot see.
func seatPlayerRegistrationGroup(t *testing.T, source string) (string, bool) {
	t.Helper()

	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, "main.go", source, 0)
	if err != nil {
		t.Fatalf("parsing generated main.go: %v", err)
	}

	// Walk down from every moves.Add* call, so the nearest enclosing one wins.
	var found string
	var describe func(node ast.Node, enclosing string)
	describe = func(node ast.Node, enclosing string) {
		ast.Inspect(node, func(n ast.Node) bool {
			if selector, ok := n.(*ast.SelectorExpr); ok {
				if selectorName(selector) == "moves.SeatPlayer" && found == "" {
					found = enclosing
				}
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := selectorName(call.Fun)
			if !strings.HasPrefix(name, "moves.Add") {
				return true
			}
			label := strings.TrimPrefix(name, "moves.")
			args := call.Args
			if strings.HasSuffix(label, "ForPhase") && len(args) > 0 {
				label += "(" + exprString(args[0]) + ")"
				args = args[1:]
			}
			for _, arg := range args {
				describe(arg, label)
			}
			return false
		})
	}

	for _, decl := range parsed.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "ConfigureMoves" {
			continue
		}
		describe(funcDecl, "")
	}

	return found, found != ""
}

func selectorName(expr ast.Expr) string {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name + "." + selector.Sel.Name
}

func exprString(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return selectorName(typed)
	}
	return ""
}
