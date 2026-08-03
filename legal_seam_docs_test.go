package boardgame

import (
	"os"
	"strings"
	"testing"
)

/*
TestSeamAllowlistDocsNameEveryAllowedBase is the prose half of
TestSeamAllowlistErrorsNameEveryAllowedBase (legal_plan_test.go).

That test closed the drift in the three BOOT ERRORS that report the
declarative-legality seam: each had hand-copied legalSupportedMovesBaseTypes as
a five-name literal against an allowlist of eight, so each told a creator that
the two reusable verbs they were most likely to embed -- MoveComponentToSlot and
DrawToPlayer -- could not opt in. They now render
legalSupportedMovesBaseTypesProse().

The documentation kept the literals, and they had drifted exactly the same way.
When this test was written, two of the four documented enumerations were stale
at five names:

  - moves/doc.go's "# Declarative Legality" opener named Default, CurrentPlayer,
    FixUp, FixUpMulti, and StartPhase -- and then, thirty lines further down in
    THE SAME DOC COMMENT, correctly named all eight. The package doc
    contradicted itself.
  - TUTORIAL.md's "The composition seam is ..." bullet named the same stale
    five. The tutorial is where a creator looks first, so it was the most
    expensive of the four to have wrong.

A boot error and a doc comment fail differently -- the error is read by someone
already stuck, the doc by someone deciding whether to try -- but they drift
identically, because both are a literal that nothing recomputes. So both get
the same mechanism.

# What this pins, and what it does not

For every anchor phrase below, this finds the enclosing paragraph and requires
that it names EVERY entry in legalSupportedMovesBaseTypes. It does not assert
that a paragraph names ONLY those: legal/doc.go's seam paragraph deliberately
contrasts the allowlist against the opaque types in the same breath
("DealCountComponents, FinishTurn, RoundRobin, and the rest -- is opaque"), and
a blanket must-not-contain would forbid that useful sentence. The direction that
actually drifted -- and the direction that misleads a creator into believing a
supported base is unsupported -- is omission, and omission is what this catches.

Every anchor must be found. An anchor that matches nothing is a hard failure
rather than a silent skip, so rewording a seam paragraph out from under this
test turns it red instead of quietly making it vacuous.
*/
func TestSeamAllowlistDocsNameEveryAllowedBase(t *testing.T) {

	// Each anchor is a phrase that introduces (or immediately follows) a
	// documented enumeration of the seam allowlist. Anchors are matched
	// case-insensitively so a sentence-initial capital does not decide whether
	// the invariant is enforced.
	anchors := map[string][]string{
		"moves/doc.go": {
			// The "# Declarative Legality" section opener, which lists the
			// supporting bases before saying what they support.
			"support an additional, optional way",
			// The parenthetical near the end of the same section.
			"the supported seam is",
		},
		"legal/doc.go": {
			"the composition seam is",
		},
		"TUTORIAL.md": {
			"the composition seam is",
		},
	}

	for path, phrases := range anchors {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("could not read %v, which this test exists to check: %v", path, err)
		}
		text := string(contents)

		for _, phrase := range phrases {
			paragraphs := paragraphsContaining(text, phrase)
			if len(paragraphs) == 0 {
				t.Errorf("%v no longer contains the anchor phrase %q. Either the seam enumeration moved (point this test at its new anchor) or it was deleted (drop the anchor). Leaving it unmatched would make this test pass without checking anything, which is the failure mode it was written to prevent.", path, phrase)
				continue
			}
			for _, paragraph := range paragraphs {
				for name := range legalSupportedMovesBaseTypes {
					if paragraphNamesBase(paragraph, name) {
						continue
					}
					t.Errorf("%v's seam enumeration anchored on %q does not name moves.%s, which IS on the seam allowlist (legalSupportedMovesBaseTypes in legal_plan.go). A creator reading this is told that embedding it rules out declarative legality, which is false. Paragraph:\n%v", path, phrase, name, paragraph)
				}
			}
		}
	}
}

// paragraphNamesBase reports whether a documentation paragraph names the given
// framework move base type in one of the two forms the docs use: a godoc link
// ("[DrawToPlayer]") or a package-qualified reference ("moves.DrawToPlayer",
// which also covers TUTORIAL.md's backticked `moves.DrawToPlayer`). Deliberately
// NOT a bare-word match: "Default" and "FixUp" appear in ordinary prose all over
// these files, and a bare-word match would let a paragraph pass on an incidental
// mention rather than on an actual enumeration entry.
func paragraphNamesBase(paragraph, name string) bool {
	return strings.Contains(paragraph, "["+name+"]") ||
		strings.Contains(paragraph, "moves."+name)
}

// paragraphsContaining returns every block of text that contains phrase
// (matched case-insensitively). A block is a blank-line-delimited paragraph,
// except that a paragraph which is a markdown bullet list is narrowed to the
// individual bullet -- TUTORIAL.md's seam enumeration is one bullet in a long
// list of unrelated ones, and reporting the whole list would both bury the
// finding and let a neighboring bullet's incidental "moves.DrawToPlayer"
// satisfy the check for a bullet that never mentions it.
func paragraphsContaining(text, phrase string) []string {

	var result []string
	for _, paragraph := range strings.Split(text, "\n\n") {
		if !blockContains(paragraph, phrase) {
			continue
		}
		for _, block := range markdownBullets(paragraph) {
			if blockContains(block, phrase) {
				result = append(result, block)
			}
		}
	}
	return result
}

// blockContains reports whether block contains phrase, ignoring case and
// ignoring how the text happens to be wrapped. Wrap-insensitivity matters: an
// anchor phrase that a future reflow splits across two lines would otherwise
// stop matching, and the test would report a missing anchor rather than the
// invariant it is actually there to check.
func blockContains(block, phrase string) bool {
	normalize := func(s string) string {
		return strings.Join(strings.Fields(strings.ToLower(s)), " ")
	}
	return strings.Contains(normalize(block), normalize(phrase))
}

// markdownBullets splits a paragraph into its top-level markdown bullets, or
// returns the paragraph unchanged if it is not a bullet list. Continuation
// lines (a wrapped bullet) stay with the bullet they belong to.
func markdownBullets(paragraph string) []string {
	lines := strings.Split(paragraph, "\n")

	var bullets []string
	for _, line := range lines {
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			bullets = append(bullets, line)
			continue
		}
		if len(bullets) == 0 {
			//Not a bullet list (or has a preamble); don't try to be clever.
			return []string{paragraph}
		}
		bullets[len(bullets)-1] += "\n" + line
	}

	if len(bullets) == 0 {
		return []string{paragraph}
	}
	return bullets
}
