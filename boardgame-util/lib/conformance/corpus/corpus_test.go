package corpus

import (
	"bytes"
	"strings"
	"testing"
)

const repoRoot = "../../../.."

// TestCorpusIsCurrent regenerates every document on the real engine and
// compares it with the checked-in corpus, so the corpus cannot rot. If the
// engine or an example game changed on purpose, regenerate with:
//
//	./scripts/go-local run ./boardgame-util conformance export
func TestCorpusIsCurrent(t *testing.T) {
	problems, err := Check(repoRoot)
	if err != nil {
		t.Fatalf("generate corpus: %v", err)
	}
	if len(problems) > 0 {
		t.Fatalf("the conformance corpus is stale; regenerate with\n\t./scripts/go-local run ./boardgame-util conformance export\n%s", strings.Join(problems, "\n"))
	}
}

// TestExportIsDeterministic exports the same runs twice and requires
// identical bytes, compressed and uncompressed.
func TestExportIsDeterministic(t *testing.T) {
	entries, err := Entries(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, entry := range entries {
		// One golden-driven and every scenario-driven game is enough here;
		// TestCorpusIsCurrent already pins every document against disk.
		if !strings.Contains(entry.File, "scenario-") && entry.File != "memory/golden-fec3529ffca41c00.json.gz" {
			continue
		}
		var runs [2][]byte
		for i := range runs {
			doc, err := entry.Build(repoRoot)
			if err != nil {
				t.Fatalf("%s: %v", entry.File, err)
			}
			if runs[i], err = doc.Marshal(); err != nil {
				t.Fatalf("%s: %v", entry.File, err)
			}
		}
		if !bytes.Equal(runs[0], runs[1]) {
			t.Errorf("%s: two exports differ%s", entry.File, firstDifference(runs[0], runs[1]))
		}
		first, err := Compress(runs[0])
		if err != nil {
			t.Fatal(err)
		}
		second, err := Compress(runs[1])
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, second) {
			t.Errorf("%s: compression is not deterministic", entry.File)
		}
		round, err := Decompress(first)
		if err != nil || !bytes.Equal(round, runs[0]) {
			t.Errorf("%s: compression does not round-trip: %v", entry.File, err)
		}
		checked++
	}
	if checked < 6 {
		t.Fatalf("only %d documents were checked", checked)
	}
}
