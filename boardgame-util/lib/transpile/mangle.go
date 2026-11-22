package transpile

import (
	"fmt"
	"strconv"
	"strings"
)

// Mangle is the TypeScript name of the function a FnRef names. It is
// reversible (Unmangle), and every name it returns is a valid TypeScript
// identifier:
//
//	pkg.fn                  pkg_fn
//	pkg.(*T).M              pkg_T_M                 (a pointer receiver)
//	pkg.(T).M               pkg_T$v_M               (a value receiver)
//	<any of the above>.funcN  <its mangling>$N      (the Nth function literal)
//
// Parts are joined by a single underscore; an underscore inside a Go
// identifier is doubled. So the parts are recovered by splitting at every
// underscore that is not one of a doubled pair: two parts are a function,
// three a method. "$" never occurs in a Go identifier.
func Mangle(ref string) (string, error) {
	base, literal, err := splitLiteral(ref)
	if err != nil {
		return "", err
	}
	dot := strings.Index(base, ".")
	if dot <= 0 {
		return "", fmt.Errorf("transpile: FnRef %q has no package", ref)
	}
	pkg, rest := base[:dot], base[dot+1:]
	var parts []string
	if strings.HasPrefix(rest, "(") {
		closing := strings.Index(rest, ").")
		if closing < 0 {
			return "", fmt.Errorf("transpile: FnRef %q has an unterminated receiver", ref)
		}
		recv, method := rest[1:closing], rest[closing+2:]
		pointer := strings.HasPrefix(recv, "*")
		recv = strings.TrimPrefix(recv, "*")
		if !identifier(recv) || !identifier(method) {
			return "", fmt.Errorf("transpile: FnRef %q is not pkg.(*T).M", ref)
		}
		recvPart := escapeUnderscores(recv)
		if !pointer {
			recvPart += "$v"
		}
		parts = []string{escapeUnderscores(pkg), recvPart, escapeUnderscores(method)}
	} else {
		if !identifier(rest) {
			return "", fmt.Errorf("transpile: FnRef %q is not pkg.fn", ref)
		}
		parts = []string{escapeUnderscores(pkg), escapeUnderscores(rest)}
	}
	if !identifier(pkg) {
		return "", fmt.Errorf("transpile: FnRef %q has a bad package name", ref)
	}
	name := strings.Join(parts, "_")
	if literal > 0 {
		name += "$" + strconv.Itoa(literal)
	}
	return name, nil
}

// Unmangle is the FnRef a Mangle result names.
func Unmangle(name string) (string, error) {
	literal := 0
	// "$v" is always followed by "_" and the method, so a trailing "$" and
	// digits can only be a literal's number.
	if i := strings.LastIndex(name, "$"); i >= 0 {
		if n, err := strconv.Atoi(name[i+1:]); err == nil && n > 0 && strconv.Itoa(n) == name[i+1:] {
			literal = n
			name = name[:i]
		}
	}
	var parts []string
	var current strings.Builder
	for i := 0; i < len(name); i++ {
		if name[i] == '_' {
			if i+1 < len(name) && name[i+1] == '_' {
				current.WriteByte('_')
				i++
				continue
			}
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(name[i])
	}
	parts = append(parts, current.String())
	var ref string
	switch len(parts) {
	case 2:
		ref = parts[0] + "." + parts[1]
	case 3:
		recv := parts[1]
		if strings.HasSuffix(recv, "$v") {
			recv = strings.TrimSuffix(recv, "$v")
		} else {
			recv = "*" + recv
		}
		ref = parts[0] + ".(" + recv + ")." + parts[2]
	default:
		return "", fmt.Errorf("transpile: %q is not a mangled FnRef", name)
	}
	if literal > 0 {
		ref += ".func" + strconv.Itoa(literal)
	}
	return ref, nil
}

// splitLiteral separates a trailing ".funcN" (N > 0) from a FnRef.
func splitLiteral(ref string) (string, int, error) {
	i := strings.LastIndex(ref, ".func")
	if i < 0 {
		return ref, 0, nil
	}
	n, err := strconv.Atoi(ref[i+len(".func"):])
	if err != nil || n <= 0 {
		return ref, 0, nil
	}
	return ref[:i], n, nil
}

func escapeUnderscores(s string) string {
	return strings.ReplaceAll(s, "_", "__")
}

// identifier reports whether s is an ASCII Go identifier (the examples'
// names are all ASCII; anything else is not mangled).
//
// An identifier that starts or ends with an underscore is refused: next to
// a separator its doubled underscores would be ambiguous.
func identifier(s string) bool {
	if s == "" || strings.HasPrefix(s, "_") || strings.HasSuffix(s, "_") {
		return false
	}
	for i, r := range s {
		switch {
		case r == '_', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}
