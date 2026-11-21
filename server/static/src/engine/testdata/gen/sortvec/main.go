// Command sortvec emits ../../sort.vectors.json: the exact permutation Go's
// sort.Sort produces for inputs with many equal keys. sort.Sort is
// deterministic but not stable, so which of two equal elements ends up first
// is a property of Go's pdqsort that Stack.SortComponents has to reproduce.
// Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/sortvec
package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"sort"
)

type vec struct {
	Label string
	// Keys are the sort keys, in input order.
	Keys []int
	// Order[i] is the input position of the element that ends at position i.
	Order []int
}

type byKey struct {
	keys []int
	tags []int
}

func (b *byKey) Len() int           { return len(b.keys) }
func (b *byKey) Less(i, j int) bool { return b.keys[i] < b.keys[j] }
func (b *byKey) Swap(i, j int) {
	b.keys[i], b.keys[j] = b.keys[j], b.keys[i]
	b.tags[i], b.tags[j] = b.tags[j], b.tags[i]
}

func main() {
	rnd := rand.New(rand.NewSource(424242))
	var out []vec
	add := func(label string, keys []int) {
		data := &byKey{keys: append([]int(nil), keys...), tags: make([]int, len(keys))}
		for i := range data.tags {
			data.tags[i] = i
		}
		sort.Sort(data)
		out = append(out, vec{Label: label, Keys: append([]int{}, keys...), Order: data.tags})
	}
	sizes := []int{0, 1, 2, 3, 5, 8, 11, 12, 13, 14, 16, 20, 24, 33, 49, 50, 51, 64, 100, 127, 128, 129, 200, 500, 1000, 3000}
	for _, n := range sizes {
		for _, distinct := range []int{1, 2, 3, 7, n/4 + 1, n + 1} {
			keys := make([]int, n)
			for i := range keys {
				keys[i] = rnd.Intn(distinct)
			}
			add("random", keys)
		}
		asc := make([]int, n)
		desc := make([]int, n)
		saw := make([]int, n)
		pipe := make([]int, n)
		almost := make([]int, n)
		for i := 0; i < n; i++ {
			asc[i] = i / 3
			desc[i] = (n - i) / 3
			saw[i] = i % 5
			if i < n/2 {
				pipe[i] = i / 2
			} else {
				pipe[i] = (n - i) / 2
			}
			almost[i] = i / 2
		}
		for k := 0; k < 3 && n > 3; k++ {
			a, b := rnd.Intn(n), rnd.Intn(n)
			almost[a], almost[b] = almost[b], almost[a]
		}
		add("ascending with ties", asc)
		add("descending with ties", desc)
		add("sawtooth", saw)
		add("organ pipe", pipe)
		add("nearly sorted", almost)
	}
	blob, err := json.Marshal(struct{ Vectors []vec }{out})
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("server/static/src/engine/testdata/sort.vectors.json", append(blob, '\n'), 0644); err != nil {
		panic(err)
	}
}
