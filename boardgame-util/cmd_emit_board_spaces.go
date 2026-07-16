package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type emitBoardSpaces struct {
	baseSubCommand
	Check bool
}

type authoredBoardSpace struct {
	Key   string
	Label string
	Order int
}

type generatedBoardSpacesFile struct {
	path     string
	contents []byte
}

func (e *emitBoardSpaces) Name() string { return "emit-board-spaces" }
func (e *emitBoardSpaces) Description() string {
	return "Generates literal TypeScript space keys from authored SVG boards"
}
func (e *emitBoardSpaces) HelpText() string {
	return e.Name() + ` scans configured game client SVG files for data-board-space,
data-board-label, and optional data-board-order. Each authored board produces a
literal-key TypeScript contract (_board_spaces.ts for board.svg), so SVG key
renames become strict compile failures instead of runtime-empty interaction.`
}
func (e *emitBoardSpaces) WritOptions() []*writ.Option {
	return []*writ.Option{{Names: []string{"check"}, Description: "Verify generated board-space contracts are current without writing files.", Decoder: writ.NewFlagDecoder(&e.Check), Flag: true}}
}
func (e *emitBoardSpaces) Run(_ writ.Path, _ []string) {
	packages, err := e.Base().GetConfig(false).Dev.AllGamePackages()
	if err != nil {
		e.Base().errAndQuit("Not all game packages were valid: " + err.Error())
	}
	if err := emitBoardSpacesForPackages(packages, e.Check); err != nil {
		e.Base().errAndQuit("Couldn't emit board spaces: " + err.Error())
	}
	if e.Check {
		fmt.Println("Generated board-space contracts are current")
	} else {
		fmt.Println("Successfully generated board-space contracts")
	}
}

func emitBoardSpacesForPackages(packages []*gamepkg.Pkg, check bool) error {
	var generated []generatedBoardSpacesFile
	for _, pkg := range packages {
		client := pkg.ClientFolder()
		if client == "" {
			continue
		}
		err := filepath.WalkDir(client, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || strings.ToLower(filepath.Ext(path)) != ".svg" {
				return nil
			}
			spaces, err := parseAuthoredBoardSpaces(path)
			if err != nil {
				return err
			}
			if len(spaces) == 0 {
				return nil
			}
			name := safeBoardContractName(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
			output := "_" + name + "_spaces.ts"
			if name == "board" {
				output = "_board_spaces.ts"
			}
			generated = append(generated, generatedBoardSpacesFile{
				path: filepath.Join(filepath.Dir(path), output), contents: generateBoardSpacesTypeScript(path, spaces),
			})
			return nil
		})
		if err != nil {
			return fmt.Errorf("%s: %w", pkg.Import(), err)
		}
	}
	sort.Slice(generated, func(i, j int) bool { return generated[i].path < generated[j].path })
	for _, file := range generated {
		current, err := os.ReadFile(file.path)
		if check {
			if err != nil || !bytes.Equal(current, file.contents) {
				return staleGeneratedClientContracts("generated board-space contract is stale: " + file.path)
			}
			continue
		}
		if err := os.WriteFile(file.path, file.contents, 0o644); err != nil {
			return err
		}
		fmt.Printf("  Generated %s (%d bytes)\n", file.path, len(file.contents))
	}
	return nil
}

func parseAuthoredBoardSpaces(path string) ([]authoredBoardSpace, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := xml.NewDecoder(io.LimitReader(file, 2*1024*1024+1))
	var result []authoredBoardSpace
	keys := map[string]bool{}
	orders := map[int]bool{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid SVG XML: %w", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := map[string]string{}
		for _, attr := range start.Attr {
			attrs[attr.Name.Local] = attr.Value
		}
		key, found := attrs["data-board-space"]
		if !found {
			continue
		}
		if key == "" || keys[key] {
			return nil, fmt.Errorf("empty or duplicate data-board-space %q", key)
		}
		label := strings.TrimSpace(attrs["data-board-label"])
		if label == "" {
			return nil, fmt.Errorf("space %q requires data-board-label for generated contracts", key)
		}
		order := len(result)
		if raw, exists := attrs["data-board-order"]; exists {
			order, err = strconv.Atoi(raw)
			if err != nil || order < 0 {
				return nil, fmt.Errorf("space %q has invalid data-board-order %q", key, raw)
			}
		}
		if orders[order] {
			return nil, fmt.Errorf("duplicate data-board-order %d", order)
		}
		keys[key], orders[order] = true, true
		result = append(result, authoredBoardSpace{Key: key, Label: label, Order: order})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result, nil
}

func generateBoardSpacesTypeScript(source string, spaces []authoredBoardSpace) []byte {
	var out strings.Builder
	out.WriteString("/* Auto-generated by boardgame-util from " + filepath.Base(source) + ". DO NOT EDIT. */\n\n")
	out.WriteString("export const BoardSpaceKeys = [")
	for index, space := range spaces {
		if index > 0 {
			out.WriteString(", ")
		}
		if number, err := strconv.Atoi(space.Key); err == nil && strconv.Itoa(number) == space.Key {
			out.WriteString(space.Key)
		} else {
			encoded, _ := json.Marshal(space.Key)
			out.Write(encoded)
		}
	}
	out.WriteString("] as const;\n\nexport type BoardSpaceKey = (typeof BoardSpaceKeys)[number];\n\n")
	out.WriteString("export const BoardSpaceLabels = {\n")
	for _, space := range spaces {
		key, _ := json.Marshal(space.Key)
		label, _ := json.Marshal(space.Label)
		fmt.Fprintf(&out, "  %s: %s,\n", key, label)
	}
	out.WriteString("} as const satisfies Readonly<Record<string, string>>;\n")
	return []byte(out.String())
}

func safeBoardContractName(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return unicode.ToLower(r)
		}
		return '_'
	}, value)
}
