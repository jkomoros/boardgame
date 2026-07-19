package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

const legalImportPath = "github.com/jkomoros/boardgame/legal"

type legalCallPath struct {
	predicate string
	path      string
	file      string
	line      int
	column    int
}

type legalPathArg struct {
	index  int
	prefix string
}

var literalLegalPathArgs = map[string][]legalPathArg{
	"PropAtLeast":                      {{0, ""}},
	"PropCompare":                      {{0, ""}},
	"PropEquals":                       {{0, ""}},
	"PropNotEquals":                    {{0, ""}},
	"StackCount":                       {{0, ""}},
	"StackEmpty":                       {{0, ""}},
	"StackNotEmpty":                    {{0, ""}},
	"ComponentPresentAt":               {{0, ""}, {1, ""}},
	"ComponentAbsentAt":                {{0, ""}, {1, ""}},
	"ComponentPresentAtKey":            {{0, ""}, {1, ""}},
	"MayMoveTo":                        {{0, ""}, {1, ""}, {2, ""}},
	"MayMoveToSlot":                    {{0, ""}, {1, ""}, {2, ""}, {3, ""}},
	"MayMoveToSameSlot":                {{0, ""}, {1, ""}, {2, ""}},
	"MayMoveAllTo":                     {{0, ""}, {1, ""}},
	"MaySwapComponents":                {{0, ""}, {1, ""}, {2, ""}},
	"MaySwapComponentsByKey":           {{0, ""}, {1, ""}, {2, ""}},
	"RevealableCardAt":                 {{0, ""}, {1, ""}, {2, ""}},
	"ComponentPropEqualsCurrentPlayer": {{0, ""}, {1, ""}, {2, "player."}},
	"ProposerIsPlayerFromMove":         {{0, "move."}},
	"PlayerBool":                       {{0, "player."}},
	"PlayerBoolIs":                     {{0, "player."}},
	"StackConstraints":                 {{0, "game."}, {1, "game."}},
}

var legalPredicateNames = map[string]string{
	"PropAtLeast":                      "propAtLeast",
	"PropCompare":                      "propCompare",
	"PropEquals":                       "propEquals",
	"PropNotEquals":                    "propNotEquals",
	"StackCount":                       "stackCount",
	"StackEmpty":                       "stackEmpty",
	"StackNotEmpty":                    "stackNotEmpty",
	"ComponentPresentAt":               "componentPresentAt",
	"ComponentAbsentAt":                "componentAbsentAt",
	"ComponentPresentAtKey":            "componentPresentAtKey",
	"MayMoveTo":                        "mayMoveTo",
	"MayMoveToSlot":                    "mayMoveToSlot",
	"MayMoveToSameSlot":                "mayMoveToSlot",
	"MayMoveAllTo":                     "mayMoveAllTo",
	"MaySwapComponents":                "maySwapComponents",
	"MaySwapComponentsByKey":           "maySwapComponentsByKey",
	"RevealableCardAt":                 "revealableCardAt",
	"ComponentPropEqualsCurrentPlayer": "componentPropEqualsCurrentPlayer",
	"ProposerIsPlayerFromMove":         "proposerIsPlayerFromMove",
	"PlayerBool":                       "playerBool",
	"PlayerBoolIs":                     "playerBool",
	"PlayerBoolAt":                     "playerBoolAt",
	"PlayerHasSubmitted":               "playerBoolAt",
	"PlayerHasNotSubmitted":            "playerBoolAt",
	"PlayerIsActive":                   "playerBoolAt",
	"PlayerIsInactive":                 "playerBoolAt",
	"PlayerSeatIsFilled":               "playerBoolAt",
	"PlayerSeatIsClosed":               "playerBoolAt",
	"PlayerIsAdmin":                    "playerBoolAt",
	"StackConstraints":                 "stackConstraints",
	"InPhase":                          "inPhase",
	"ProposerIsCurrentPlayer":          "proposerIsCurrentPlayer",
	"AllActivePlayers":                 "allActivePlayers",
}

var behaviorProps = map[string]string{
	"PlayerHasSubmitted":    "PlayerSubmitted",
	"PlayerHasNotSubmitted": "PlayerSubmitted",
	"PlayerIsActive":        "PlayerInactive",
	"PlayerIsInactive":      "PlayerInactive",
	"PlayerSeatIsFilled":    "SeatFilled",
	"PlayerSeatIsClosed":    "SeatClosed",
	"PlayerIsAdmin":         "IsAdmin",
}

func sourceLegalCalls(packageDir string) ([]legalCallPath, error) {
	fset := token.NewFileSet()
	loaded, err := packages.Load(&packages.Config{
		Dir: packageDir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Fset:       fset,
		Tests:      false,
		BuildFlags: []string{"-mod=readonly"},
	}, ".")
	if err != nil {
		return nil, fmt.Errorf("load package for legal-call analysis: %w", err)
	}
	if len(loaded) != 1 || len(loaded[0].Errors) != 0 {
		var details []string
		for _, pkg := range loaded {
			for _, problem := range pkg.Errors {
				details = append(details, problem.Error())
			}
		}
		return nil, fmt.Errorf("load package for legal-call analysis returned %d packages with errors: %v", len(loaded), details)
	}
	pkg := loaded[0]
	var result []legalCallPath
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			function := calledLegalFunction(pkg.TypesInfo, call.Fun)
			if function == nil {
				return true
			}
			name := function.Name()
			predicate, known := legalPredicateNames[name]
			if !known {
				return true
			}
			for _, descriptor := range literalLegalPathArgs[name] {
				if descriptor.index >= len(call.Args) {
					continue
				}
				if value, ok := stringLiteral(call.Args[descriptor.index]); ok {
					result = append(result, positionedLegalCall(fset, predicate, descriptor.prefix+value, call.Args[descriptor.index].Pos()))
				}
			}
			if name == "PlayerBoolAt" && len(call.Args) >= 2 {
				if prefix, ok := playerSelectorPrefix(pkg.TypesInfo, call.Args[0]); ok {
					if prop, ok := stringLiteral(call.Args[1]); ok {
						result = append(result, positionedLegalCall(fset, predicate, prefix+prop, call.Args[1].Pos()))
					}
				}
			}
			if prop, ok := behaviorProps[name]; ok && len(call.Args) >= 1 {
				if prefix, ok := playerSelectorPrefix(pkg.TypesInfo, call.Args[0]); ok {
					result = append(result, positionedLegalCall(fset, predicate, prefix+prop, call.Fun.Pos()))
				}
			}
			switch name {
			case "InPhase":
				result = append(result, positionedLegalCall(fset, predicate, "game.Phase", call.Fun.Pos()))
			case "ProposerIsCurrentPlayer":
				result = append(result,
					positionedLegalCall(fset, predicate, "move.TargetPlayerIndex", call.Fun.Pos()),
					positionedLegalCall(fset, predicate, "game.CurrentPlayer", call.Fun.Pos()),
				)
			case "AllActivePlayers":
				if len(call.Args) == 1 {
					for _, quantified := range allActivePlayerPaths(pkg.TypesInfo, call.Args[0]) {
						result = append(result, positionedLegalCall(fset, predicate, quantified.path, quantified.pos))
					}
				}
			}
			return true
		})
	}
	return result, nil
}

type positionedPath struct {
	path string
	pos  token.Pos
}

func allActivePlayerPaths(info *types.Info, expression ast.Expr) []positionedPath {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return nil
	}
	function := calledLegalFunction(info, call.Fun)
	if function == nil {
		return nil
	}
	switch function.Name() {
	case "Any":
		var result []positionedPath
		for _, argument := range call.Args {
			result = append(result, allActivePlayerPaths(info, argument)...)
		}
		return result
	case "PlayerBool", "PlayerBoolIs":
		if len(call.Args) >= 1 {
			if prop, ok := stringLiteral(call.Args[0]); ok {
				return []positionedPath{{path: "players[*]." + prop, pos: call.Args[0].Pos()}}
			}
		}
	case "PropAtLeast", "PropCompare":
		if len(call.Args) >= 1 {
			if path, ok := stringLiteral(call.Args[0]); ok && strings.HasPrefix(path, "player.") {
				return []positionedPath{{path: "players[*]." + strings.TrimPrefix(path, "player."), pos: call.Args[0].Pos()}}
			}
		}
	}
	return nil
}

func calledLegalFunction(info *types.Info, expression ast.Expr) *types.Func {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	function, ok := info.Uses[selector.Sel].(*types.Func)
	if !ok || function.Pkg() == nil || function.Pkg().Path() != legalImportPath {
		return nil
	}
	return function
}

func playerSelectorPrefix(info *types.Info, expression ast.Expr) (string, bool) {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return "", false
	}
	function := calledLegalFunction(info, call.Fun)
	if function == nil {
		return "", false
	}
	switch function.Name() {
	case "CurrentPlayer":
		return "player.", true
	case "Proposer":
		return "proposer.", true
	case "PlayerFromMove":
		if len(call.Args) == 1 {
			if field, ok := stringLiteral(call.Args[0]); ok {
				return "players[move." + field + "].", true
			}
		}
	}
	return "", false
}

func stringLiteral(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func positionedLegalCall(fset *token.FileSet, predicate, path string, pos token.Pos) legalCallPath {
	position := fset.Position(pos)
	return legalCallPath{predicate: predicate, path: path, file: position.Filename, line: position.Line, column: position.Column}
}
