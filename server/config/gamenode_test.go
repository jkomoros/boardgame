package config

import (
	"encoding/json"
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
					"github.com/jkomoros": &GameNode{
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
					"github.com/jkomoros": &GameNode{
						Mids: map[string]*GameNode{
							"boardgame": &GameNode{
								Leafs: []string{
									"checkers",
									"blackjack",
								},
							},
							"other-repo": &GameNode{
								Leafs: []string{
									"pass",
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
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
				},
			},
			nil,
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
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
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
				},
			},
		},
		{
			"Simple no nesting one duplicate",
			&GameNode{
				Leafs: []string{
					"checkers",
					"blackjack",
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
					"checkers",
					"blackjack",
					"pass",
				},
			},
		},
		{
			"One level nesting no overlap",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/bob": &GameNode{
						Leafs: []string{
							"pass",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
					"github.com/bob": &GameNode{
						Leafs: []string{
							"pass",
						},
					},
				},
			},
		},
		{
			"One level nesting partial overlap",
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
					"github.com/a": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"memory",
						},
					},
					"github.com/b": &GameNode{
						Leafs: []string{
							"pass",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"github.com/jkomoros": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
							"memory",
						},
					},
					"github.com/a": &GameNode{
						Leafs: []string{
							"checkers",
							"blackjack",
						},
					},
					"github.com/b": &GameNode{
						Leafs: []string{
							"pass",
						},
					},
				},
			},
		},
		{
			"overlap at second level",
			&GameNode{
				Mids: map[string]*GameNode{
					"one": &GameNode{
						Mids: map[string]*GameNode{
							"github.com/jkomoros": &GameNode{
								Leafs: []string{
									"checkers",
									"blackjack",
								},
							},
							"github.com/a": &GameNode{
								Leafs: []string{
									"checkers",
									"blackjack",
								},
							},
						},
					},
					"two": &GameNode{
						Leafs: []string{
							"a",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"one": &GameNode{
						Mids: map[string]*GameNode{
							"github.com/jkomoros": &GameNode{
								Leafs: []string{
									"checkers",
									"memory",
								},
							},
							"github.com/b": &GameNode{
								Leafs: []string{
									"pass",
								},
							},
						},
					},
					"three": &GameNode{
						Leafs: []string{
							"b",
						},
					},
				},
			},
			&GameNode{
				Mids: map[string]*GameNode{
					"one": &GameNode{
						Mids: map[string]*GameNode{
							"github.com/jkomoros": &GameNode{
								Leafs: []string{
									"checkers",
									"blackjack",
									"memory",
								},
							},
							"github.com/a": &GameNode{
								Leafs: []string{
									"checkers",
									"blackjack",
								},
							},
							"github.com/b": &GameNode{
								Leafs: []string{
									"pass",
								},
							},
						},
					},
					"two": &GameNode{
						Leafs: []string{
							"a",
						},
					},
					"three": &GameNode{
						Leafs: []string{
							"b",
						},
					},
				},
			},
		},
	}

	for i, test := range tests {

		result := test.base.extend(test.other)
		assert.For(t, i, test.description).ThatActual(result).Equals(test.expected).ThenDiffOnFail()

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
					"one": &GameNode{
						Leafs: []string{
							"a",
							"b",
						},
					},
					"two": &GameNode{
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
					"github.com/jkomoros": &GameNode{
						Mids: map[string]*GameNode{
							"one": &GameNode{
								Leafs: []string{
									"a",
									"b",
								},
							},
							"two": &GameNode{
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
	}

	for i, test := range tests {
		result := test.in.List()
		assert.For(t, i, test.description).ThatActual(result).Equals(test.expected).ThenDiffOnFail()
	}
}
