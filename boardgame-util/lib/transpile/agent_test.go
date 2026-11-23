package transpile

import (
	"bytes"
	"strings"
	"testing"
)

const testBrief = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// TestGenerateKeepsAgentBlocks: an agent-written block (port accept's
// header) survives regeneration verbatim while its function's Go source
// hash is unchanged, takes the translator's place in the bodies switch,
// brings its imports, and is dropped when the hash moves on.
func TestGenerateKeepsAgentBlocks(t *testing.T) {
	p := loadFixture(t)
	base, err := p.Generate([]byte("null"), nil, "aaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	const ref = "emitgame.(*moveAll).Apply"
	var hash string
	for _, fn := range base.Functions {
		if fn.Func == ref {
			hash = fn.SourceHash
		}
	}
	if hash == "" {
		t.Fatal("no " + ref + " in the fixture")
	}
	code := "export function emitgame_moveAll_Apply(state: RulesState, m: MoveInstance): string | null {\n" +
		"\t// An agent's text: 'first' in a comment and a string imports nothing.\n" +
		"\tconst s = 'moveTo';\n" +
		"\tvoid s;\n" +
		"\treturn idiv(lenOf(null), 1) === 0 ? null : 'unreachable';\n" +
		"}\n"
	existing := WithAgentBlock(base.Files["bodies.gen.ts"], ref, "test-model", testBrief, hash, "bbbbbbbb", code)
	out, err := p.Generate([]byte("null"), existing, "cccccccc")
	if err != nil {
		t.Fatal(err)
	}
	text := string(out.Files["bodies.gen.ts"])
	wantHeader := AgentHeader("test-model", testBrief, ref, hash, "bbbbbbbb")
	if !strings.Contains(text, wantHeader+"\n"+code) {
		t.Fatalf("the agent block was not kept verbatim:\n%s", text)
	}
	if strings.Contains(text, "go:"+ref+" sha256:") && strings.Count(text, "go:"+ref+" ") != 1 {
		t.Errorf("both the translator's and the agent's block of %s were written", ref)
	}
	if !strings.Contains(text, "case '"+ref+"':\n\t\t\treturn emitgame_moveAll_Apply;") {
		t.Errorf("the agent block is not in the bodies switch")
	}
	if !strings.Contains(text, "import { idiv } from '../../go/num.ts';") && !strings.Contains(text, "idiv, ") && !strings.Contains(text, ", idiv") {
		t.Errorf("the agent block's idiv was not imported")
	}
	if b, ok := out.Agent[ref]; !ok || b.Agent != "test-model" || b.Brief != testBrief {
		t.Errorf("Output.Agent = %+v", out.Agent)
	}
	blocks := Blocks(out.Files["bodies.gen.ts"])
	if blocks[ref].Producer != "agent:test-model" || blocks["emitgame.folded"].Producer != "transpile@1" {
		t.Errorf("Blocks producers: %q, %q", blocks[ref].Producer, blocks["emitgame.folded"].Producer)
	}

	// Stable: regenerating from its own output changes nothing.
	again, err := p.Generate([]byte("null"), out.Files["bodies.gen.ts"], "dddddddd")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again.Files["bodies.gen.ts"], out.Files["bodies.gen.ts"]) {
		t.Errorf("a second Generate changed the file with an agent block")
	}

	// Adopting: without the block, the translator's own text is back,
	// byte for byte (the other blocks kept their first: commits).
	adopted, err := p.Generate([]byte("null"), WithoutBlock(out.Files["bodies.gen.ts"], ref), "aaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(adopted.Files["bodies.gen.ts"], base.Files["bodies.gen.ts"]) {
		t.Errorf("removing the agent block did not give back the translator's file")
	}

	// A stale hash: the Go changed since the agent wrote it.
	stale := bytes.Replace(out.Files["bodies.gen.ts"], []byte("go:"+ref+" "+hash), []byte("go:"+ref+" sha256:"+strings.Repeat("0", 64)), 1)
	dropped, err := p.Generate([]byte("null"), stale, "aaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(dropped.Files["bodies.gen.ts"]), "@generated agent:") || len(dropped.Agent) != 0 {
		t.Errorf("an agent block with a stale source hash was kept")
	}
}

func TestAgentImports(t *testing.T) {
	got := agentImports("export function f(state: RulesState): LegalVerdict {\n\t// nonNil in a comment\n\tconst x = 'first';\n\treturn x.length > 0 ? passVerdict() : failT(`player`);\n}\n")
	want := map[string][]string{
		"legal/types.ts":      {"failT", "passVerdict"},
		"type:legal/types.ts": {"LegalVerdict"},
		"type:rules/state.ts": {"RulesState"},
	}
	if len(got) != len(want) {
		t.Fatalf("agentImports = %v, want %v", got, want)
	}
	for module, names := range want {
		if strings.Join(got[module], ",") != strings.Join(names, ",") {
			t.Errorf("agentImports[%s] = %v, want %v", module, got[module], names)
		}
	}
}
