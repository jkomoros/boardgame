package port

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// source is one game package's syntax and types, for the brief: verbatim
// source of functions, type and variable declarations, and what a
// function names.
type source struct {
	pkg   *packages.Package
	fset  *token.FileSet
	files []*ast.File
	// units are every function declaration and argument function literal
	// of the hand-written files, by FnRef (gamedef's numbering).
	units map[string]ast.Node
	// enclosing is the FnRef of the declaration a literal unit is in.
	enclosing map[string]string
	refs      map[*types.Func]string
	// gameState and playerState are the game's sub-state types.
	gameState, playerState *types.Named
	// texts caches file contents by name.
	texts map[string][]byte
}

func loadSource(importPath, dir string) (*source, error) {
	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir: dir,
	}
	loaded, err := packages.Load(config, importPath)
	if err != nil {
		return nil, fmt.Errorf("port: load %s: %w", importPath, err)
	}
	if len(loaded) != 1 || len(loaded[0].Errors) > 0 || loaded[0].Types == nil {
		return nil, fmt.Errorf("port: %s did not load as one type-checked package", importPath)
	}
	s := &source{pkg: loaded[0], fset: loaded[0].Fset, units: map[string]ast.Node{}, enclosing: map[string]string{}, refs: map[*types.Func]string{}, texts: map[string][]byte{}}
	for _, file := range s.pkg.Syntax {
		if !generatedFile(file) {
			s.files = append(s.files, file)
		}
	}
	for _, file := range s.files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			fn, ok := s.pkg.TypesInfo.Defs[fd.Name].(*types.Func)
			if !ok {
				continue
			}
			ref := fnRef(fn)
			s.refs[fn] = ref
			s.units[ref] = fd
			if fd.Body == nil {
				continue
			}
			count := 0
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if lit, ok := n.(*ast.FuncLit); ok {
					count++
					name := ref + ".func" + strconv.Itoa(count)
					s.units[name] = lit
					s.enclosing[name] = ref
					return false
				}
				return true
			})
		}
	}
	s.findStateTypes()
	return s, nil
}

func generatedFile(file *ast.File) bool {
	if ast.IsGenerated(file) {
		return true
	}
	for _, group := range file.Comments {
		if group.Pos() > file.Package {
			break
		}
		text := group.Text()
		if strings.Contains(text, "'boardgame-util codegen'") && strings.Contains(text, "DO NOT EDIT") {
			return true
		}
	}
	return false
}

// fnRef is gamedef's FnRef of a declared function.
func fnRef(fn *types.Func) string {
	sig := fn.Type().(*types.Signature)
	pkgName := fn.Pkg().Name()
	if recv := sig.Recv(); recv != nil {
		t := recv.Type()
		star := ""
		if pointer, ok := t.(*types.Pointer); ok {
			t = pointer.Elem()
			star = "*"
		}
		name := types.TypeString(t, func(*types.Package) string { return "" })
		if named, ok := types.Unalias(t).(*types.Named); ok {
			name = named.Obj().Name()
		}
		return pkgName + ".(" + star + name + ")." + fn.Name()
	}
	return pkgName + "." + fn.Name()
}

// findStateTypes: the delegate is NewDelegate's result; the sub-states are
// what its GameStateConstructor and PlayerStateConstructor return.
func (s *source) findStateTypes() {
	returned := func(fd *ast.FuncDecl) *types.Named {
		var result *types.Named
		if fd == nil || fd.Body == nil {
			return nil
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			if result != nil {
				return false
			}
			if _, ok := n.(*ast.FuncLit); ok {
				return false
			}
			ret, ok := n.(*ast.ReturnStmt)
			if !ok || len(ret.Results) == 0 {
				return true
			}
			if named := namedOf(s.pkg.TypesInfo.TypeOf(ret.Results[0])); named != nil && named.Obj().Pkg() == s.pkg.Types {
				result = named
			}
			return true
		})
		return result
	}
	var delegate *types.Named
	for _, file := range s.files {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name == "NewDelegate" {
				delegate = returned(fd)
			}
		}
	}
	if delegate == nil {
		return
	}
	method := func(name string) *ast.FuncDecl {
		obj, _, _ := types.LookupFieldOrMethod(types.NewPointer(delegate), true, s.pkg.Types, name)
		fn, ok := obj.(*types.Func)
		if !ok || fn.Pkg() != s.pkg.Types {
			return nil
		}
		fd, _ := s.units[s.refs[fn]].(*ast.FuncDecl)
		return fd
	}
	s.gameState = returned(method("GameStateConstructor"))
	s.playerState = returned(method("PlayerStateConstructor"))
}

func namedOf(t types.Type) *types.Named {
	if t == nil {
		return nil
	}
	if pointer, ok := types.Unalias(t).(*types.Pointer); ok {
		t = pointer.Elem()
	}
	named, _ := types.Unalias(t).(*types.Named)
	return named
}

// fileText is a file's bytes.
func (s *source) fileText(name string) []byte {
	if text, ok := s.texts[name]; ok {
		return text
	}
	text, err := os.ReadFile(name)
	if err != nil {
		text = nil
	}
	s.texts[name] = text
	return text
}

// nodeText is a node's source, verbatim.
func (s *source) nodeText(from, to token.Pos) string {
	start, end := s.fset.Position(from), s.fset.Position(to)
	text := s.fileText(start.Filename)
	if text == nil || start.Offset < 0 || end.Offset > len(text) || start.Offset > end.Offset {
		return ""
	}
	return string(text[start.Offset:end.Offset])
}

// position is pos relative to the package directory.
func (s *source) position(pos token.Pos) (string, int) {
	at := s.fset.Position(pos)
	name := at.Filename
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return name, at.Line
}

// lineText is the source line at pos, trimmed.
func (s *source) lineText(file string, line int) string {
	for _, f := range s.files {
		name := s.fset.Position(f.Pos()).Filename
		if strings.HasSuffix(name, "/"+file) || name == file {
			lines := strings.Split(string(s.fileText(name)), "\n")
			if line >= 1 && line <= len(lines) {
				return strings.TrimSpace(lines[line-1])
			}
		}
	}
	return ""
}

// unitText is a unit's source, from the function's first byte to its
// closing brace, doc comment excluded (what gamedef hashes).
func (s *source) unitText(ref string) string {
	node, ok := s.units[ref]
	if !ok {
		return ""
	}
	return s.nodeText(node.Pos(), node.End())
}

// signatureText is a declaration's Go signature: the source from `func` to
// the body's opening brace, exclusive.
func (s *source) signatureText(ref string) string {
	node, ok := s.units[ref]
	if !ok {
		return ""
	}
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Body == nil {
			return s.nodeText(n.Pos(), n.End())
		}
		return strings.TrimSpace(s.nodeText(n.Pos(), n.Body.Lbrace))
	case *ast.FuncLit:
		return strings.TrimSpace(s.nodeText(n.Pos(), n.Body.Lbrace))
	}
	return ""
}

// declText is the declaration of a named type or package variable: its
// spec's source, with the `type` or `var` keyword, and the doc comment of
// the spec or of an ungrouped declaration. For an object outside the
// game's package, the file is parsed from its position.
func (s *source) declText(obj types.Object) string {
	pos := obj.Pos()
	if !pos.IsValid() {
		return ""
	}
	at := s.fset.Position(pos)
	if at.Filename == "" {
		return ""
	}
	var file *ast.File
	fset := s.fset
	for _, f := range s.pkg.Syntax {
		if s.fset.Position(f.Pos()).Filename == at.Filename {
			file = f
		}
	}
	text := s.fileText(at.Filename)
	if file == nil {
		fset = token.NewFileSet()
		parsed, err := parser.ParseFile(fset, at.Filename, text, parser.ParseComments)
		if err != nil {
			return ""
		}
		file = parsed
	}
	keyword := "type"
	if _, ok := obj.(*types.Var); ok {
		keyword = "var"
	}
	if _, ok := obj.(*types.Const); ok {
		keyword = "const"
	}
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			var name *ast.Ident
			var names []*ast.Ident
			var doc *ast.CommentGroup
			switch sp := spec.(type) {
			case *ast.TypeSpec:
				name, doc = sp.Name, sp.Doc
			case *ast.ValueSpec:
				names, doc = sp.Names, sp.Doc
			}
			if name != nil {
				names = []*ast.Ident{name}
			}
			for _, n := range names {
				if fset.Position(n.Pos()).Line != at.Line || n.Name != obj.Name() {
					continue
				}
				if doc == nil && gd.Lparen == token.NoPos {
					doc = gd.Doc
				}
				from := spec.Pos()
				if gd.Lparen == token.NoPos {
					from = gd.Pos()
				}
				start := fset.Position(from).Offset
				end := fset.Position(spec.End()).Offset
				body := string(text[start:end])
				if gd.Lparen != token.NoPos {
					body = keyword + " " + body
				}
				if doc != nil {
					body = string(text[fset.Position(doc.Pos()).Offset:fset.Position(doc.End()).Offset]) + "\n" + body
				}
				return body
			}
		}
	}
	return ""
}

// named collects what a unit names: the game-declared and framework types
// it mentions (its signature, receiver, conversions, assertions, and the
// type of every value whose field or method it selects), the package
// variables and constants it reads, and the fields it selects on the
// game's sub-states, moves and component values.
type named struct {
	types  []*types.TypeName
	vars   []*types.Var
	consts []*types.Const
	// frameworkConsts are constants of other packages it reads.
	frameworkConsts []*types.Const
	selection       []fieldUse
}

// fieldUse is one field of a game struct the unit reads or writes.
type fieldUse struct {
	recv  *types.Named
	field string
}

func (s *source) collect(ref string) named {
	node := s.units[ref]
	var out named
	seenType := map[*types.TypeName]bool{}
	seenVar := map[*types.Var]bool{}
	seenConst := map[*types.Const]bool{}
	seenField := map[fieldUse]bool{}
	addType := func(t types.Type) {
		named := namedOf(t)
		if named == nil || named.Obj().Pkg() == nil {
			return
		}
		if obj := named.Obj(); !seenType[obj] {
			seenType[obj] = true
			out.types = append(out.types, obj)
		}
	}
	info := s.pkg.TypesInfo
	if fd, ok := node.(*ast.FuncDecl); ok && fd.Recv != nil {
		for _, field := range fd.Recv.List {
			addType(info.TypeOf(field.Type))
		}
	}
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			switch obj := info.Uses[x].(type) {
			case *types.TypeName:
				addType(obj.Type())
			case *types.Var:
				if obj.Parent() == s.pkg.Types.Scope() && !seenVar[obj] {
					seenVar[obj] = true
					out.vars = append(out.vars, obj)
				}
			case *types.Const:
				if obj.Pkg() == s.pkg.Types && obj.Parent() == s.pkg.Types.Scope() && !seenConst[obj] {
					seenConst[obj] = true
					out.consts = append(out.consts, obj)
				} else if obj.Pkg() != nil && obj.Pkg() != s.pkg.Types && !seenConst[obj] {
					seenConst[obj] = true
					out.frameworkConsts = append(out.frameworkConsts, obj)
				}
			}
		case *ast.SelectorExpr:
			sel := info.Selections[x]
			if sel == nil {
				return true
			}
			recv := namedOf(sel.Recv())
			if recv != nil && recv.Obj().Pkg() == s.pkg.Types {
				addType(recv)
				if sel.Kind() == types.FieldVal {
					use := fieldUse{recv: recv, field: x.Sel.Name}
					if !seenField[use] {
						seenField[use] = true
						out.selection = append(out.selection, use)
					}
				}
			} else if recv != nil {
				addType(recv)
			}
			if sel.Kind() == types.FieldVal {
				addType(sel.Obj().Type())
			}
		}
		return true
	})
	// Results of the function's calls whose types it holds in locals.
	ast.Inspect(node, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if v, ok := info.Defs[id].(*types.Var); ok && v != nil {
				addType(v.Type())
			}
		}
		return true
	})
	sort.SliceStable(out.types, func(i, j int) bool { return typeKey(out.types[i]) < typeKey(out.types[j]) })
	sort.SliceStable(out.vars, func(i, j int) bool { return out.vars[i].Name() < out.vars[j].Name() })
	sort.SliceStable(out.consts, func(i, j int) bool { return out.consts[i].Name() < out.consts[j].Name() })
	sort.SliceStable(out.frameworkConsts, func(i, j int) bool {
		a, b := out.frameworkConsts[i], out.frameworkConsts[j]
		return a.Pkg().Path()+"."+a.Name() < b.Pkg().Path()+"."+b.Name()
	})
	return out
}

// typeKey orders types: the game's own first, then by package path and name.
func typeKey(obj *types.TypeName) string {
	return obj.Pkg().Path() + "." + obj.Name()
}

// assignments are the statements of the package that assign v, each with
// the declarations, in the same function, of the identifiers its right-hand
// side uses (one level): how a package variable gets its value.
type assignment struct {
	in   string
	file string
	line int
	text string
	from []string
}

func (s *source) assignments(v *types.Var) []assignment {
	info := s.pkg.TypesInfo
	var out []assignment
	for ref, node := range s.units {
		if _, ok := node.(*ast.FuncDecl); !ok {
			continue
		}
		ast.Inspect(node, func(n ast.Node) bool {
			as, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, lhs := range as.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok || info.Uses[id] != v {
					continue
				}
				file, line := s.position(as.Pos())
				a := assignment{in: ref, file: file, line: line, text: s.nodeText(as.Pos(), as.End())}
				seen := map[types.Object]bool{}
				for _, rhs := range as.Rhs {
					ast.Inspect(rhs, func(m ast.Node) bool {
						rid, ok := m.(*ast.Ident)
						if !ok {
							return true
						}
						obj, ok := info.Uses[rid].(*types.Var)
						if !ok || seen[obj] || obj.Parent() == s.pkg.Types.Scope() {
							return true
						}
						seen[obj] = true
						if def := s.definingStmt(node, obj); def != "" {
							a.from = append(a.from, def)
						}
						return true
					})
				}
				out = append(out, a)
			}
			return true
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].file != out[j].file {
			return out[i].file < out[j].file
		}
		return out[i].line < out[j].line
	})
	return out
}

// definingStmt is the statement of fn that declares obj, verbatim.
func (s *source) definingStmt(fn ast.Node, obj types.Object) string {
	info := s.pkg.TypesInfo
	var found string
	ast.Inspect(fn, func(n ast.Node) bool {
		if found != "" {
			return false
		}
		switch x := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range x.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && info.Defs[id] == obj {
					found = s.nodeText(x.Pos(), x.End())
				}
			}
		case *ast.DeclStmt:
			ast.Inspect(x, func(m ast.Node) bool {
				if id, ok := m.(*ast.Ident); ok && info.Defs[id] == obj {
					found = s.nodeText(x.Pos(), x.End())
				}
				return true
			})
		}
		return true
	})
	return found
}

// goCalls are the qualified names (subset.QualifiedName's spelling) of the
// calls a unit makes outside its own package, sorted, without duplicates.
func (s *source) goCalls(ref string, qualified func(*types.Func) string) []string {
	info := s.pkg.TypesInfo
	seen := map[string]bool{}
	ast.Inspect(s.units[ref], func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		var id *ast.Ident
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			id = fun
		case *ast.SelectorExpr:
			id = fun.Sel
		}
		if id == nil {
			return true
		}
		fn, ok := info.Uses[id].(*types.Func)
		if !ok || fn.Pkg() == nil || fn.Pkg() == s.pkg.Types {
			return true
		}
		seen[qualified(fn)] = true
		return true
	})
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
