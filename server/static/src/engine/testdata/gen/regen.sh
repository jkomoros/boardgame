#!/bin/sh
# Regenerates every generated file under server/static/src/engine from real
# Go. Run from anywhere inside the repository:
#
#     server/static/src/engine/testdata/gen/regen.sh
#
# The outputs are deterministic for a given Go toolchain and repository state
# (the generators seed their own scripting PRNGs with constants), so a rerun
# with no Go change leaves the tree clean.
#
#   cooked       -> ../../rand-cooked.ts         math/rand's rngCooked table
#   randvec      -> ../rand.vectors.json         math/rand and FNV-1 answers
#   sortvec      -> ../sort.vectors.json         sort.Sort's order among equal keys
#   defs         -> ../defs/*.def.json           GameDefinitions from live managers
#   idsvec       -> ../ids.vectors.json          crypto/sha1 and component ids
#   stackvec     -> ../stacks.vectors.json       stack operations, step by step
#                -> ../constraints.vectors.json  stack constraints and by-key access
#   sanitizevec  -> ../sanitize.vectors.json     per-viewer views of the vector game
#   goldenvec    -> ../golden/*.golden.json      digests and views of golden histories
#   rejectvec    -> ../rejections.vectors.json   proposals the Go engine refuses
#   facadevec    -> ../facade.vectors.json,      the Go-shaped facade (engine/go/):
#                   ../facade-stacks.vectors.json  language answers and stack calls
#   libvec       -> ../lib.vectors.json          range enums, grid graphs, dice, playing cards
#   stockvec     -> ../stock/*                   stockgame and seatgame played on the Go engine
#
# vecgame/auto_reader.go is produced by `boardgame-util codegen` run inside
# vecgame/; rerun that only if vecgame/main.go's state structs change.
# The same holds for stackvec/cgame/, stockgame/ and seatgame/.
set -eu
root=$(git rev-parse --show-toplevel)
cd "$root"
for gen in cooked randvec sortvec defs idsvec stackvec sanitizevec goldenvec rejectvec facadevec libvec stockvec; do
	echo "== $gen" >&2
	./scripts/go-local run "./server/static/src/engine/testdata/gen/$gen"
done
