package gamedef

import (
	"crypto/sha256"
	"encoding/hex"
	"go/format"
)

// SourceHashPrefix begins every SourceHash.
const SourceHashPrefix = "sha256:"

// SourceHash is the identity of one function's Go source, as FunctionEntry
// and the generated-code headers name it: "sha256:" followed by the
// lower-case hex SHA-256 of the source after gofmt (go/format.Source), so a
// change of layout alone does not change it.
//
// src is one function declaration (or function literal), from its first
// byte to its closing brace, doc comment excluded. If src does not parse as
// Go, it is hashed as given; callers pass source cut from a package that
// type-checked, so that does not arise in practice.
func SourceHash(src []byte) string {
	formatted, err := format.Source(src)
	if err != nil {
		formatted = src
	}
	sum := sha256.Sum256(formatted)
	return SourceHashPrefix + hex.EncodeToString(sum[:])
}
