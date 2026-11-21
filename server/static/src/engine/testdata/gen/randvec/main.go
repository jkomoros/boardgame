// Command randvec emits ../../rand.vectors.json: known answers for Go's
// math/rand (v1) generator and hash/fnv's 64-bit FNV-1, produced by running
// the real Go implementations. Run from the repository root:
//
//	./scripts/go-local run ./server/static/src/engine/testdata/gen/randvec
package main

import (
	"encoding/json"
	"hash/fnv"
	"math"
	"math/rand"
	"os"
	"strconv"
)

type seedVec struct {
	Seed    string
	Uint64  []string
	Int63   []string
	Uint32  []uint32
	Int31   []int32
	Int31n  map[string][]int32
	Int63n  map[string][]string
	Intn    map[string][]string
	Float64 []string // strconv 'g' -1 round-trip form
	Float32 []string
	Perm    map[string][]int
	Shuffle map[string][]int
	// Mixed interleaves calls on one generator, the way a move's Apply does.
	Mixed []string
}

type fnvVec struct {
	Input string
	Sum64 string
	// Seed is int64(Sum64), which is what state.Rand() hands to NewSource.
	Seed string
	// Perm5 is rand.New(rand.NewSource(Seed)).Perm(5).
	Perm5 []int
}

func s(v int64) string  { return strconv.FormatInt(v, 10) }
func u(v uint64) string { return strconv.FormatUint(v, 10) }

func main() {
	seeds := []int64{
		0, 1, -1, 2, 42, 89482311, math.MaxInt32, math.MaxInt32 + 1, 2 * math.MaxInt32, -math.MaxInt32,
		math.MaxInt64, math.MinInt64, math.MinInt64 + 1, 1 << 32, -(1 << 32), 1<<40 + 12345,
		-7046029254386353131, 6364136223846793005, 1234567890123456789, -1234567890123456789,
	}
	// A spread of further seeds from a fixed LCG so the file is stable.
	x := uint64(88172645463325252)
	for i := 0; i < 40; i++ {
		x = x*6364136223846793005 + 1442695040888963407
		seeds = append(seeds, int64(x))
	}

	int31ns := []int32{1, 2, 3, 5, 7, 10, 52, 100, 1 << 16, 1<<16 + 1, 1 << 30, 1<<30 + 1, math.MaxInt32 - 1, math.MaxInt32}
	int63ns := []int64{1, 2, 3, 10, 1000, 1 << 31, 1<<31 + 1, 1 << 32, 1<<40 + 3, 1 << 52, 1<<52 - 1, 3 * (1 << 50)}
	intns := []int{1, 2, 6, 52, math.MaxInt32, 1 << 31, 1<<31 + 7, 1<<45 + 1}
	permSizes := []int{0, 1, 2, 3, 4, 10, 20, 52, 120, 607, 1000}

	var out struct {
		Seeds []seedVec
		FNV   []fnvVec
	}

	for idx, seed := range seeds {
		full := idx < 6 || idx%10 == 0
		v := seedVec{Seed: s(seed), Int31n: map[string][]int32{}, Int63n: map[string][]string{}, Intn: map[string][]string{}, Perm: map[string][]int{}, Shuffle: map[string][]int{}}
		r := rand.New(rand.NewSource(seed))
		n64 := 20
		if full {
			n64 = 1300 // past two full turns of the 607-word register
		}
		for i := 0; i < n64; i++ {
			v.Uint64 = append(v.Uint64, u(r.Uint64()))
		}
		r = rand.New(rand.NewSource(seed))
		for i := 0; i < 20; i++ {
			v.Int63 = append(v.Int63, s(r.Int63()))
		}
		r = rand.New(rand.NewSource(seed))
		for i := 0; i < 20; i++ {
			v.Uint32 = append(v.Uint32, r.Uint32())
		}
		r = rand.New(rand.NewSource(seed))
		for i := 0; i < 20; i++ {
			v.Int31 = append(v.Int31, r.Int31())
		}
		for _, n := range int31ns {
			r = rand.New(rand.NewSource(seed))
			var vals []int32
			for i := 0; i < 12; i++ {
				vals = append(vals, r.Int31n(n))
			}
			v.Int31n[strconv.Itoa(int(n))] = vals
		}
		for _, n := range int63ns {
			r = rand.New(rand.NewSource(seed))
			var vals []string
			for i := 0; i < 12; i++ {
				vals = append(vals, s(r.Int63n(n)))
			}
			v.Int63n[s(n)] = vals
		}
		for _, n := range intns {
			r = rand.New(rand.NewSource(seed))
			var vals []string
			for i := 0; i < 12; i++ {
				vals = append(vals, strconv.Itoa(r.Intn(n)))
			}
			v.Intn[strconv.Itoa(n)] = vals
		}
		r = rand.New(rand.NewSource(seed))
		for i := 0; i < 10; i++ {
			v.Float64 = append(v.Float64, strconv.FormatFloat(r.Float64(), 'g', -1, 64))
		}
		r = rand.New(rand.NewSource(seed))
		for i := 0; i < 10; i++ {
			v.Float32 = append(v.Float32, strconv.FormatFloat(float64(r.Float32()), 'g', -1, 64))
		}
		for _, n := range permSizes {
			if n > 120 && !full {
				continue
			}
			r = rand.New(rand.NewSource(seed))
			v.Perm[strconv.Itoa(n)] = r.Perm(n)
			r = rand.New(rand.NewSource(seed))
			arr := make([]int, n)
			for i := range arr {
				arr[i] = i
			}
			r.Shuffle(n, func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
			v.Shuffle[strconv.Itoa(n)] = arr
		}
		r = rand.New(rand.NewSource(seed))
		for i := 0; i < 6; i++ {
			p, _ := json.Marshal(r.Perm(5 + i))
			v.Mixed = append(v.Mixed, "perm:"+string(p))
			v.Mixed = append(v.Mixed, "intn6:"+strconv.Itoa(r.Intn(6)))
			v.Mixed = append(v.Mixed, "float64:"+strconv.FormatFloat(r.Float64(), 'g', -1, 64))
			v.Mixed = append(v.Mixed, "int63:"+s(r.Int63()))
		}
		out.Seeds = append(out.Seeds, v)
	}

	inputs := []string{
		"", "a", "ab", "insecurestarterdefault0", "insecurestarterdefault1", "insecurestarterdefault17",
		"6CB9690B3B55FE3D25F10986D87C27210", "6CB9690B3B55FE3D25F10986D87C272189",
		"714104226CB9A317ED004924E4A56BBE12", "FAKESALTFORTESTING",
		"héllo wörld", "日本語のテキスト", "emoji 🎲🃏 pair", "\x00\x01\x7f",
		"The quick brown fox jumps over the lazy dog",
	}
	for i := 0; i < 64; i++ {
		inputs = append(inputs, "GAMEID0123456789SALTSALTSALTSALT"+strconv.Itoa(i))
	}
	for _, in := range inputs {
		h := fnv.New64()
		h.Write([]byte(in))
		sum := h.Sum64()
		out.FNV = append(out.FNV, fnvVec{Input: in, Sum64: u(sum), Seed: s(int64(sum)), Perm5: rand.New(rand.NewSource(int64(sum))).Perm(5)})
	}

	blob, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("server/static/src/engine/testdata/rand.vectors.json", append(blob, '\n'), 0644); err != nil {
		panic(err)
	}
}
