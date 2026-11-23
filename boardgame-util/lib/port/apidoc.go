package port

import (
	"strings"
)

// tsDecl is one top-level exported declaration of a TypeScript module, as
// the brief shows it: its doc comment and its text (a function's
// signature without its body; an interface, type or constant whole).
type tsDecl struct {
	Name string
	// Kind is "function", "const", "interface" or "type".
	Kind string
	Doc  string
	Text string
}

// tsModule is one TypeScript module's leading comment and exported
// declarations, in source order.
type tsModule struct {
	Header string
	Decls  []tsDecl
}

// parseTSModule reads the exported declarations of a TypeScript source the
// way the engine writes them: each starts at column 0 with `export`, the
// comment lines right above it are its doc, a function's signature ends at
// the first line that ends with `{` with its parentheses balanced, and any
// other declaration ends where its brackets balance at a `;` or `}`.
// Nothing here is a TypeScript parser; it is enough for the engine's own
// modules, whose layout the sandbox and lint keep regular.
func parseTSModule(text string) tsModule {
	lines := strings.Split(text, "\n")
	codes := codeLines(text)
	var mod tsModule
	i := 0
	// The leading comment, up to the first blank line after it or the
	// first non-comment line.
	var header []string
	for ; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if strings.HasPrefix(t, "//") {
			header = append(header, lines[i])
			continue
		}
		break
	}
	mod.Header = strings.Join(header, "\n")
	var doc []string
	inBlock := false
	for ; i < len(lines); i++ {
		line := lines[i]
		t := strings.TrimSpace(line)
		if inBlock {
			doc = append(doc, line)
			if strings.Contains(t, "*/") {
				inBlock = false
			}
			continue
		}
		if strings.HasPrefix(line, "/**") || strings.HasPrefix(line, "/*") {
			doc = []string{line}
			if !strings.Contains(t[2:], "*/") {
				inBlock = true
			}
			continue
		}
		if strings.HasPrefix(line, "//") {
			if len(doc) > 0 && !strings.HasPrefix(strings.TrimSpace(doc[len(doc)-1]), "//") {
				doc = nil
			}
			doc = append(doc, line)
			continue
		}
		if !strings.HasPrefix(line, "export ") {
			doc = nil
			continue
		}
		decl := tsDecl{Doc: strings.Join(doc, "\n")}
		doc = nil
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		decl.Kind = fields[1]
		name := fields[2]
		if decl.Kind == "declare" || decl.Kind == "default" || decl.Kind == "{" || decl.Kind == "*" {
			continue
		}
		if decl.Kind == "async" {
			continue
		}
		cut := strings.IndexAny(name, "(<:=;{ ")
		if cut >= 0 {
			name = name[:cut]
		}
		decl.Name = name
		var body, bodyCode []string
		depth := 0
		for ; i < len(lines); i++ {
			l := lines[i]
			body = append(body, l)
			bodyCode = append(bodyCode, codes[i])
			depth += bracketDelta(codes[i])
			trimmed := strings.TrimSpace(codes[i])
			if decl.Kind == "function" {
				if strings.HasSuffix(trimmed, "{") && parenBalance(bodyCode) == 0 {
					last := len(body) - 1
					cut := strings.LastIndex(codes[i], "{")
					body[last] = strings.TrimRight(body[last][:cut], " \t")
					// Skip the body.
					for depth > 0 && i+1 < len(lines) {
						i++
						depth += bracketDelta(codes[i])
					}
					break
				}
				continue
			}
			if depth <= 0 && (strings.HasSuffix(trimmed, ";") || decl.Kind == "interface" && strings.HasSuffix(trimmed, "}")) {
				break
			}
			if depth <= 0 && trimmed == "" {
				break
			}
		}
		decl.Text = strings.Join(body, "\n")
		mod.Decls = append(mod.Decls, decl)
	}
	return mod
}

// codeLines is text split into lines with every comment and string or
// template literal blanked, so that only code is left to count brackets in.
func codeLines(text string) []string {
	out := []byte(text)
	const (
		code = iota
		line
		block
		str
	)
	state := code
	var quote byte
	for i := 0; i < len(out); i++ {
		c := out[i]
		switch state {
		case code:
			switch {
			case c == '/' && i+1 < len(out) && out[i+1] == '/':
				state = line
				out[i] = ' '
			case c == '/' && i+1 < len(out) && out[i+1] == '*':
				state = block
				out[i] = ' '
			case c == '\'' || c == '"' || c == '`':
				state, quote = str, c
				out[i] = ' '
			}
		case line:
			if c == '\n' {
				state = code
			} else {
				out[i] = ' '
			}
		case block:
			if c == '*' && i+1 < len(out) && out[i+1] == '/' {
				out[i], out[i+1] = ' ', ' '
				i++
				state = code
			} else if c != '\n' {
				out[i] = ' '
			}
		case str:
			if c == '\\' && i+1 < len(out) {
				out[i] = ' '
				if out[i+1] != '\n' {
					out[i+1] = ' '
				}
				i++
				continue
			}
			if c == quote || c == '\n' && quote != '`' {
				state = code
			}
			if c != '\n' {
				out[i] = ' '
			}
		}
	}
	return strings.Split(string(out), "\n")
}

// bracketDelta counts the braces, brackets and parentheses a code line opens.
func bracketDelta(line string) int {
	depth := 0
	for _, r := range line {
		switch r {
		case '{', '[', '(':
			depth++
		case '}', ']', ')':
			depth--
		}
	}
	return depth
}

func parenBalance(lines []string) int {
	depth := 0
	for _, line := range lines {
		for _, r := range line {
			switch r {
			case '(':
				depth++
			case ')':
				depth--
			}
		}
	}
	return depth
}

// find is the declaration named name, or false.
func (m tsModule) find(name string) (tsDecl, bool) {
	for _, d := range m.Decls {
		if d.Name == name {
			return d, true
		}
	}
	return tsDecl{}, false
}

// writeDecl writes one declaration, indented, doc comment first.
func writeDecl(b *strings.Builder, d tsDecl, indent string) {
	if d.Doc != "" {
		b.WriteString(indentLines(d.Doc, indent))
		b.WriteByte('\n')
	}
	b.WriteString(indentLines(d.Text, indent))
	b.WriteByte('\n')
}

// indentLines prefixes every non-empty line with indent.
func indentLines(text, indent string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) != "" {
			lines[i] = indent + l
		} else {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}
