package transpile

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

// Version is the producer the generated headers name: transpile@<Version>.
// Bump it when the emitter changes what it writes for the same Go.
const Version = 1

// Package is one loaded game package: its static report (the function
// manifest, roles, verdicts) and the syntax and types the emitter reads.
type Package struct {
	Static *gamedef.StaticReport

	// AssumeGapsClosed translates the functions refused only for engine
	// gaps (bucket engineGap) as if the gap table were empty. It is for
	// measuring what the translator would do once the engine has them;
	// output written with it is never checked in.
	AssumeGapsClosed bool

	pkg   *packages.Package
	tpkg  *types.Package
	info  *types.Info
	fset  *token.FileSet
	files []*ast.File

	delegate    *types.Named
	gameState   *types.Named
	playerState *types.Named

	// units are every function declaration and known function literal of the
	// package's hand-written files, by FnRef.
	units map[string]ast.Node
	// refs is the FnRef of each declared function.
	refs map[*types.Func]string
	// erased are the functions the static half erases (concreteStates).
	erased map[*types.Func]bool
	// roles is each manifest function's role.
	roles map[string]string
}

// Function is the translator's answer for one manifest function.
type Function struct {
	Func string
	Role string
	// Name is the mangled TypeScript name.
	Name       string
	SourceHash string
	// Code is the TypeScript function declaration (no header), set when
	// Refusals is empty.
	Code string
	// Refusals are the checker's refusals for a function the manifest
	// refuses, or the emitter's own when it could not translate a function
	// the checker accepts (a disagreement: a bug in one of them).
	Refusals []Refusal
	// EmitterRefused is true when the refusals are the emitter's, not the
	// checker's.
	EmitterRefused bool
	// Imports are the facade names Code uses, by module path relative to the
	// engine directory ("go/api.ts").
	Imports map[string][]string
}

// Load loads the game package pattern names (an import path or a directory
// pattern), resolved from dir ("" for the current directory).
func Load(pattern, dir string) (*Package, error) {
	static, err := gamedef.LoadStatic(pattern, dir)
	if err != nil {
		return nil, err
	}
	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir: dir,
	}
	loaded, err := packages.Load(config, pattern)
	if err != nil {
		return nil, fmt.Errorf("transpile: load %s: %w", pattern, err)
	}
	if len(loaded) != 1 || len(loaded[0].Errors) > 0 || loaded[0].Types == nil {
		return nil, fmt.Errorf("transpile: %s did not load as one type-checked package", pattern)
	}
	pkg := loaded[0]
	p := &Package{
		Static: static,
		pkg:    pkg,
		tpkg:   pkg.Types,
		info:   pkg.TypesInfo,
		fset:   pkg.Fset,
		units:  make(map[string]ast.Node),
		refs:   make(map[*types.Func]string),
		erased: make(map[*types.Func]bool),
		roles:  make(map[string]string),
	}
	for _, file := range pkg.Syntax {
		if !generatedFile(file) {
			p.files = append(p.files, file)
		}
	}
	p.collectUnits()
	p.findStateTypes()
	byRef := make(map[string]*types.Func)
	for fn, ref := range p.refs {
		byRef[ref] = fn
	}
	for _, erased := range static.Erased {
		if fn := byRef[erased.Func]; fn != nil {
			p.erased[fn] = true
		}
	}
	for _, entry := range static.Functions {
		p.roles[entry.Func] = entry.Role
	}
	return p, nil
}

// Name is the game package's name.
func (p *Package) Name() string { return p.tpkg.Name() }

// Dir is the package directory.
func (p *Package) Dir() string { return p.Static.Dir }

// Translate translates every function of the manifest, in callee-first
// order (StaticReport.Order).
func (p *Package) Translate() []Function {
	byFunc := make(map[string]gamedef.FunctionEntry)
	for _, entry := range p.Static.Functions {
		byFunc[entry.Func] = entry
	}
	var result []Function
	for _, ref := range p.Static.Order {
		entry := byFunc[ref]
		fn := Function{Func: ref, Role: entry.Role, SourceHash: entry.SourceHash}
		name, err := Mangle(ref)
		if err != nil {
			fn.Refusals = []Refusal{{Func: ref, Reason: ReasonUnknownType, Detail: "emitter: " + err.Error()}}
			fn.EmitterRefused = true
			result = append(result, fn)
			continue
		}
		fn.Name = name
		gapOnly := p.AssumeGapsClosed && p.Static.Buckets[ref] == gamedef.BucketEngineGap
		if entry.Verdict != gamedef.VerdictSubset && !gapOnly {
			fn.Refusals = entry.Refusals
			result = append(result, fn)
			continue
		}
		code, imports, refusal := p.emitFunction(ref, entry.Role)
		if refusal != nil {
			fn.Refusals = []Refusal{*refusal}
			fn.EmitterRefused = true
		} else {
			fn.Code = code
			fn.Imports = imports
		}
		result = append(result, fn)
	}
	return result
}

// TryRefused runs the emitter on a function the checker refuses, for the
// report: whether the emitter alone could have written it (most often a
// stale engine-gap entry). Its output is never written.
//
// A function the checker refuses can hold anything, so an emitter failure
// that is not a refusal is reported as one here instead of propagating.
func (p *Package) TryRefused(ref string) (ok bool, refused *Refusal) {
	defer func() {
		if r := recover(); r != nil {
			ok, refused = false, &Refusal{Func: ref, Reason: ReasonUnknownType, Detail: fmt.Sprintf("emitter: %v", r)}
		}
	}()
	role := p.roles[ref]
	_, _, refusal := p.emitFunction(ref, role)
	return refusal == nil, refusal
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

// collectUnits finds every declaration and every function literal passed
// as an argument, numbered funcN among the declaration's top-level literals
// in source order (gamedef's numbering).
func (p *Package) collectUnits() {
	for _, file := range p.files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			fn, ok := p.info.Defs[fd.Name].(*types.Func)
			if !ok {
				continue
			}
			ref := fnRef(fn)
			p.refs[fn] = ref
			p.units[ref] = fd
			if fd.Body == nil {
				continue
			}
			count := 0
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if lit, ok := n.(*ast.FuncLit); ok {
					count++
					p.units[ref+".func"+strconv.Itoa(count)] = lit
					return false
				}
				return true
			})
		}
	}
}

func (p *Package) findStateTypes() {
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
			t := p.info.TypeOf(ret.Results[0])
			if pointer, ok := t.(*types.Pointer); ok {
				t = pointer.Elem()
			}
			if named, ok := types.Unalias(t).(*types.Named); ok && named.Obj().Pkg() == p.tpkg {
				result = named
			}
			return true
		})
		return result
	}
	for _, file := range p.files {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name == "NewDelegate" {
				p.delegate = returned(fd)
			}
		}
	}
	if p.delegate == nil {
		return
	}
	method := func(name string) *ast.FuncDecl {
		obj, _, _ := types.LookupFieldOrMethod(types.NewPointer(p.delegate), true, p.tpkg, name)
		fn, ok := obj.(*types.Func)
		if !ok || fn.Pkg() != p.tpkg {
			return nil
		}
		ref := p.refs[fn]
		fd, _ := p.units[ref].(*ast.FuncDecl)
		return fd
	}
	p.gameState = returned(method("GameStateConstructor"))
	p.playerState = returned(method("PlayerStateConstructor"))
}

// delegateMethod is the game's own declaration of a GameDelegate method.
func (p *Package) delegateMethod(name string) *types.Func {
	if p.delegate == nil {
		return nil
	}
	obj, _, _ := types.LookupFieldOrMethod(types.NewPointer(p.delegate), true, p.tpkg, name)
	fn, ok := obj.(*types.Func)
	if !ok || fn.Pkg() != p.tpkg {
		return nil
	}
	if _, known := p.refs[fn]; !known {
		return nil
	}
	return fn
}

// position is a refusal position relative to the package directory.
func (p *Package) position(pos token.Pos) (string, int, int) {
	if !pos.IsValid() {
		return "", 0, 0
	}
	at := p.fset.Position(pos)
	file := at.Filename
	if rel, err := filepath.Rel(p.Static.Dir, file); err == nil && !strings.HasPrefix(rel, "..") {
		file = filepath.ToSlash(rel)
	} else {
		file = filepath.Base(file)
	}
	return file, at.Line, at.Column
}

func (p *Package) callee(call *ast.CallExpr) *types.Func {
	fn, _ := typeutil.Callee(p.info, call).(*types.Func)
	return fn
}

// sortedImports turns a set of names per module into sorted lists.
func sortedImports(set map[string]map[string]bool) map[string][]string {
	out := make(map[string][]string, len(set))
	for module, names := range set {
		list := make([]string, 0, len(names))
		for name := range names {
			list = append(list, name)
		}
		sort.Strings(list)
		out[module] = list
	}
	return out
}
