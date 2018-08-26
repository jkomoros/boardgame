package config

import (
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestUnmarshalGameNode(t *testing.T) {

	tests := []struct {
		description string
		in          string
		expected    *GameNode
	}{
		{
			"No nesting",
			`
				[
					"checkers",
					"blackjack"
				]
			`,
			&GameNode{
				Leafs: []string{
					"checkers",
					"blackjack",
				},
			},
		},
		{
			"One level nesting",
			`
				{
					"github.com/jkomoros":[
						"checkers",
						"blackjack"
					]
				}
			`,
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
				},
			},
		},
		{
			"two layer nesting",
			`
				{
					"github.com/jkomoros":{
						"boardgame": [
							"checkers",
							"blackjack"
						],
						"other-repo": [
							"pass"
						]
					}
				}
			`,
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Mids: map[string]*GameNode{
							"boardgame": {
								Leafs: []string{
									"checkers",
									"blackjack",
								},
							},
							"other-repo": {
								Leafs: []string{
									"pass",
								},
							},
						},
					},
				},
			},
		},
		{
			"mixed mid and leaf",
			`
				{
					"github.com/jkomoros":{
						"boardgame": {
							"checkers": [
								""
							],
							"subdir": [
								"blackjack",
								"memory"
							]
						}
					}
				}
			`,
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Mids: map[string]*GameNode{
							"boardgame": {
								Mids: map[string]*GameNode{
									"checkers": {
										Leafs: []string{
											"",
										},
									},
									"subdir": {
										Leafs: []string{
											"blackjack",
											"memory",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		var gameNode *GameNode
		err := json.Unmarshal([]byte(test.in), &gameNode)
		if test.expected == nil {
			assert.For(t, i, test.description).ThatActual(err).IsNotNil()
		} else {
			assert.For(t, i, test.description).ThatActual(err).IsNil()
		}
		assert.For(t, i, test.description).ThatActual(gameNode).Equals(test.expected).ThenDiffOnFail()
	}

}

func TestGameNodeExtend(t *testing.T) {
	tests := []struct {
		description string
		base        *GameNode
		other       *GameNode
		expected    *GameNode
	}{
		{
			"No other",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"blackjack",
							"checkers",
						},
					},
				},
			},
			nil,
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"blackjack",
							"checkers",
						},
					},
				},
			},
		},
		{
			"No base",
			nil,
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"blackjack",
							"checkers",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"blackjack",
							"checkers",
						},
					},
				},
			},
		},
		{
			"Simple no nesting one duplicate",
			&GameNode{
				Leafs: []string{
					"blackjack",
					"checkers",
				},
			},
			&GameNode{
				Leafs: []string{
					"checkers",
					"pass",
				},
			},
			&GameNode{
				Leafs: []string{
					"blackjack",
					"checkers",
					"pass",
				},
			},
		},
		{
			"One level nesting no overlap",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"blackjack",
							"checkers",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/bob": {
						Leafs: []string{
							"pass",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com": {
						Mids: map[string]*GameNode{
							"jkomoros": {
								Leafs: []string{
									"blackjack",
									"checkers",
								},
							},
							"bob": {
								Leafs: []string{
									"pass",
								},
							},
						},
					},
				},
			},
		},
		{
			"One level nesting partial overlap",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
					"github.com/a": {
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"checkers",
							"memory",
						},
					},
					"github.com/b": {
						Leafs: []string{
							"pass",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com": {
						Mids: map[string]*GameNode{
							"jkomoros": {
								Leafs: []string{
									"blackjack",
									"checkers",
									"memory",
								},
							},
							"a": {
								Leafs: []string{
									"blackjack",
									"checkers",
								},
							},
							"b": {
								Leafs: []string{
									"pass",
								},
							},
						},
					},
				},
			},
		},
		{
			"overlap at second level",
			&GameNode{
				Mids: map[string]*GameNode{
					"one": {
						Mids: map[string]*GameNode{
							"github.com/jkomoros": {
								Leafs: []string{
									"checkers",
									"blackjack",
								},
							},
							"github.com/a": {
								Leafs: []string{
									"checkers",
									"blackjack",
								},
							},
						},
					},
					"two": {
						Leafs: []string{
							"a",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"one": {
						Mids: map[string]*GameNode{
							"github.com/jkomoros": {
								Leafs: []string{
									"checkers",
									"memory",
								},
							},
							"github.com/b": {
								Leafs: []string{
									"pass",
								},
							},
						},
					},
					"three": {
						Leafs: []string{
							"b",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"one/github.com": {
						Mids: map[string]*GameNode{
							"jkomoros": {
								Leafs: []string{
									"blackjack",
									"checkers",
									"memory",
								},
							},
							"a": {
								Leafs: []string{
									"blackjack",
									"checkers",
								},
							},
							"b": {
								Leafs: []string{
									"pass",
								},
							},
						},
					},
					"two": {
						Leafs: []string{
							"a",
						},
					},
					"three": {
						Leafs: []string{
							"b",
						},
					},
				},
			},
		},
		{
			"overlap leaf + mid",
			&GameNode{
				Mids: map[string]*GameNode{
					"toplevelgame": {
						Leafs: []string{
							"",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"toplevelgame": {
						Leafs: []string{
							"blackjack",
							"memory",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"toplevelgame": {
						Leafs: []string{
							"",
							"blackjack",
							"memory",
						},
					},
				},
			},
		},
		{
			"key is mid and leafs (test normalize is called)",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Mids: map[string]*GameNode{
							"examples": {
								Leafs: []string{
									"blackjack",
									"memory",
								},
							},
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Leafs: []string{
							"other-dir/bar",
							"other-dir/baz",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Mids: map[string]*GameNode{
							"examples": {
								Leafs: []string{
									"blackjack",
									"memory",
								},
							},
							"other-dir": {
								Leafs: []string{
									"bar",
									"baz",
								},
							},
						},
					},
				},
			},
		},
		{
			"inner key should be merged to normalize",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros/boardgame/examples": {
						Leafs: []string{
							"blackjack",
							"checkers",
							"memory",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros/games": {
						Leafs: []string{
							"darwin",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Mids: map[string]*GameNode{
							"boardgame/examples": {
								Leafs: []string{
									"blackjack",
									"checkers",
									"memory",
								},
							},
							"games": {
								Leafs: []string{
									"darwin",
								},
							},
						},
					},
				},
			},
		},
		{
			"Single item",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros/boardgame/examples": {
						Leafs: []string{
							"checkers",
						},
					},
				},
			},
			nil,
			&GameNode{
				Leafs: []string{
					"github.com/jkomoros/boardgame/examples/checkers",
				},
			},
		},
	}

	for i, test := range tests {

		result := test.base.extend(test.other)
		if !assert.For(t, i, test.description).ThatActual(result).Equals(test.expected).ThenDiffOnFail().Passed() {
			spew.Dump(result)
		}

	}
}

func TestGameNodeList(t *testing.T) {
	tests := []struct {
		description string
		in          *GameNode
		expected    []string
	}{
		{
			"No nest",
			&GameNode{
				Leafs: []string{
					"b",
					"a",
					"c",
				},
			},
			[]string{
				"a",
				"b",
				"c",
			},
		},
		{
			"Single nest",
			&GameNode{
				Mids: map[string]*GameNode{
					"one": {
						Leafs: []string{
							"a",
							"b",
						},
					},
					"two": {
						Leafs: []string{
							"c",
							"d",
						},
					},
				},
			},
			[]string{
				"one/a",
				"one/b",
				"two/c",
				"two/d",
			},
		},
		{
			"Double nest with path separator in key",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Mids: map[string]*GameNode{
							"one": {
								Leafs: []string{
									"a",
									"b",
								},
							},
							"two": {
								Leafs: []string{
									"c",
									"d",
								},
							},
						},
					},
				},
			},
			[]string{
				"github.com/jkomoros/one/a",
				"github.com/jkomoros/one/b",
				"github.com/jkomoros/two/c",
				"github.com/jkomoros/two/d",
			},
		},
		{
			"Mixed leaf and min",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Mids: map[string]*GameNode{
							"toplevelgame": {
								Leafs: []string{
									"",
								},
							},
							"examples": {
								Leafs: []string{
									"blackjack",
									"memory",
								},
							},
						},
					},
				},
			},
			[]string{
				"github.com/jkomoros/examples/blackjack",
				"github.com/jkomoros/examples/memory",
				"github.com/jkomoros/toplevelgame",
			},
		},
		{
			"Mixed leaf and min with extra",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": {
						Mids: map[string]*GameNode{
							"toplevelgame": {
								Leafs: []string{
									"",
									"another",
								},
							},
							"examples": {
								Leafs: []string{
									"blackjack",
									"memory",
								},
							},
						},
					},
				},
			},
			[]string{
				"github.com/jkomoros/examples/blackjack",
				"github.com/jkomoros/examples/memory",
				"github.com/jkomoros/toplevelgame",
				"github.com/jkomoros/toplevelgame/another",
			},
		},
	}

	for i, test := range tests {
		result := test.in.List()
		assert.For(t, i, test.description).ThatActual(result).Equals(test.expected).ThenDiffOnFail()
	}
}
