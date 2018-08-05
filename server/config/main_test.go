package config

import (
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
