package config

import (
	"encoding/json"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBaseExtend(t *testing.T) {

	tests := []struct {
		description string
		in          *Config
		out         *Config
	}{
		{
			"No op",
			&Config{
				nil,
				&ConfigMode{
					"AllowedOriginsDev",
					"DefaultPortDev",
					"FirebaseProjectIDDev",
					[]string{
						"AdminUserIdDev1",
						"AdminUserIdDev2",
					},
					true,
					map[string]string{
						"Config1": "Dev",
						"Config2": "Dev",
					},
				},
				nil,
			},
			&Config{
				nil,
				&ConfigMode{
					"AllowedOriginsDev",
					"DefaultPortDev",
					"FirebaseProjectIDDev",
					[]string{
						"AdminUserIdDev1",
						"AdminUserIdDev2",
					},
					true,
					map[string]string{
						"Config1": "Dev",
						"Config2": "Dev",
					},
				},
				nil,
			},
		},
		{
			"Simple derive",
			&Config{
				&ConfigMode{
					"AllowedOriginsBase",
					"DefaultPortBase",
					"FirebaseProjectIDBase",
					[]string{
						"AdminUserIdBase1",
						"AdminUserIdDev2",
					},
					true,
					map[string]string{
						"Config1": "Base",
						"Config2": "Base",
						"Config3": "Base",
					},
				},
				&ConfigMode{
					"AllowedOriginsDev",
					"",
					"FirebaseProjectIDDev",
					[]string{
						"AdminUserIdDev1",
					},
					false,
					map[string]string{
						"Config2": "Dev",
					},
				},
				nil,
			},
			&Config{
				&ConfigMode{
					"AllowedOriginsBase",
					"DefaultPortBase",
					"FirebaseProjectIDBase",
					[]string{
						"AdminUserIdBase1",
						"AdminUserIdDev2",
					},
					true,
					map[string]string{
						"Config1": "Base",
						"Config2": "Base",
						"Config3": "Base",
					},
				},
				&ConfigMode{
					"AllowedOriginsDev",
					"DefaultPortBase",
					"FirebaseProjectIDDev",
					[]string{
						"AdminUserIdBase1",
						"AdminUserIdDev2",
						"AdminUserIdDev1",
					},
					true,
					map[string]string{
						"Config1": "Base",
						"Config2": "Dev",
						"Config3": "Base",
					},
				},
				&ConfigMode{
					"AllowedOriginsBase",
					"DefaultPortBase",
					"FirebaseProjectIDBase",
					[]string{
						"AdminUserIdBase1",
						"AdminUserIdDev2",
					},
					true,
					map[string]string{
						"Config1": "Base",
						"Config2": "Base",
						"Config3": "Base",
					},
				},
			},
		},
	}

	for i, test := range tests {
		test.in.derive()
		assert.For(t, i, test.description).ThatActual(test.in).Equals(test.out).ThenDiffOnFail()
	}

}

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
