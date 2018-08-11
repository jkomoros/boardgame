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
					[]string{
						"AdminUserIdDev1",
						"AdminUserIdDev2",
					},
					true,
					map[string]string{
						"Config1": "Dev",
						"Config2": "Dev",
					},
					"bolt",
					"gastring",
					nil,
					"https://localhost",
					nil,
					nil,
				},
				nil,
			},
			&Config{
				nil,
				&ConfigMode{
					"AllowedOriginsDev",
					"DefaultPortDev",
					[]string{
						"AdminUserIdDev1",
						"AdminUserIdDev2",
					},
					true,
					map[string]string{
						"Config1": "Dev",
						"Config2": "Dev",
					},
					"bolt",
					"gastring",
					nil,
					"https://localhost",
					nil,
					nil,
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
					"bolt",
					"gastring",
					nil,
					"https://localhost",
					nil,
					nil,
				},
				&ConfigMode{
					"AllowedOriginsDev",
					"",
					[]string{
						"AdminUserIdDev1",
					},
					false,
					map[string]string{
						"Config2": "Dev",
					},
					"bolt",
					"gastring",
					nil,
					"https://localhost",
					nil,
					nil,
				},
				nil,
			},
			&Config{
				&ConfigMode{
					"AllowedOriginsBase",
					"DefaultPortBase",
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
					"bolt",
					"gastring",
					nil,
					"https://localhost",
					nil,
					nil,
				},
				&ConfigMode{
					"AllowedOriginsDev",
					"DefaultPortBase",
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
					"bolt",
					"gastring",
					nil,
					"https://localhost",
					nil,
					nil,
				},
				&ConfigMode{
					"AllowedOriginsBase",
					"DefaultPortBase",
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
					"bolt",
					"gastring",
					nil,
					"https://localhost",
					nil,
					nil,
				},
			},
		},
	}

	for i, test := range tests {
		test.in.derive()
		assert.For(t, i, test.description).ThatActual(test.in).Equals(test.out).ThenDiffOnFail()
	}

}

func TestApiHostDerivation(t *testing.T) {

	tests := []struct {
		description string
		prodMode    bool
		in          *ConfigMode
		out         *ConfigMode
	}{
		{
			"No op",
			false,
			&ConfigMode{
				ApiHost:     "provided",
				DefaultPort: "8888",
			},
			&ConfigMode{
				ApiHost:     "provided",
				DefaultPort: "8888",
			},
		},
		{
			"dev",
			false,
			&ConfigMode{
				DefaultPort: "8888",
			},
			&ConfigMode{
				ApiHost:     "http://localhost:8888",
				DefaultPort: "8888",
			},
		},
		{
			"prod non default port",
			true,
			&ConfigMode{
				ApiHost:     "",
				DefaultPort: "8080",
				Firebase: &FirebaseConfig{
					StorageBucket: "example-boardgame.appspot.com",
				},
			},
			&ConfigMode{
				ApiHost:     "https://example-boardgame.appspot.com:8080",
				DefaultPort: "8080",
				Firebase: &FirebaseConfig{
					StorageBucket: "example-boardgame.appspot.com",
				},
			},
		},
		{
			"prod no default port",
			true,
			&ConfigMode{
				ApiHost: "",
				Firebase: &FirebaseConfig{
					StorageBucket: "example-boardgame.appspot.com",
				},
			},
			&ConfigMode{
				ApiHost: "https://example-boardgame.appspot.com",
				Firebase: &FirebaseConfig{
					StorageBucket: "example-boardgame.appspot.com",
				},
			},
		},
		{
			"prod default port 80",
			true,
			&ConfigMode{
				ApiHost:     "",
				DefaultPort: "80",
				Firebase: &FirebaseConfig{
					StorageBucket: "example-boardgame.appspot.com",
				},
			},
			&ConfigMode{
				ApiHost:     "https://example-boardgame.appspot.com",
				DefaultPort: "80",
				Firebase: &FirebaseConfig{
					StorageBucket: "example-boardgame.appspot.com",
				},
			},
		},
	}

	for i, test := range tests {
		test.in.derive(test.prodMode)
		assert.For(t, i, test.description).ThatActual(test.in).Equals(test.out).ThenDiffOnFail()
	}

}
