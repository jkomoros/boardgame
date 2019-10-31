package config

import (
	"testing"

	"github.com/workfit/tester/assert"
)

func TestBaseExtend(t *testing.T) {

	tests := []struct {
		description string
		in          *RawConfig
		out         *Config
	}{
		{
			"No op",
			&RawConfig{
				nil,
				&RawConfigMode{
					ModeCommon{
						"AllowedOriginsDev",
						"DefaultPortDev",
						"DefaultStaticPortDev",
						[]string{
							"AdminUserIdDev1",
							"AdminUserIdDev2",
						},
						true,
						true,
						map[string]string{
							"Config1": "Dev",
							"Config2": "Dev",
						},
						"bolt",
						"gastring",
						nil,
						"https://localhost",
					},
					nil,
				},
				nil,
				"",
			},
			&Config{
				&Mode{
					ModeCommon{
						"AllowedOriginsDev",
						"DefaultPortDev",
						"DefaultStaticPortDev",
						[]string{
							"AdminUserIdDev1",
							"AdminUserIdDev2",
						},
						true,
						true,
						map[string]string{
							"Config1": "Dev",
							"Config2": "Dev",
						},
						"bolt",
						"gastring",
						nil,
						"https://localhost",
					},
					nil,
					nil,
				},
				&Mode{
					ModeCommon{
						AllowedOrigins:    "*",
						DefaultPort:       "8080",
						DefaultStaticPort: "80",
						Storage:           make(map[string]string),
					},
					nil,
					nil,
				},
				nil,
				nil,
				nil,
			},
		},
		{
			"Simple derive",
			&RawConfig{
				&RawConfigMode{
					ModeCommon{
						"AllowedOriginsBase",
						"DefaultPortBase",
						"DefaultStaticPortBase",
						[]string{
							"AdminUserIdBase1",
							"AdminUserIdDev2",
						},
						false,
						false,
						map[string]string{
							"Config1": "Base",
							"Config2": "Base",
							"Config3": "Base",
						},
						"bolt",
						"gastring",
						nil,
						"https://localhost",
					},
					nil,
				},
				&RawConfigMode{
					ModeCommon{
						"AllowedOriginsDev",
						"",
						"",
						[]string{
							"AdminUserIdDev1",
						},
						true,
						true,
						map[string]string{
							"Config2": "Dev",
						},
						"bolt",
						"gastring",
						nil,
						"https://localhost",
					},
					nil,
				},
				nil,
				"",
			},
			&Config{
				&Mode{
					ModeCommon{
						"AllowedOriginsDev",
						"DefaultPortBase",
						"DefaultStaticPortBase",
						[]string{
							"AdminUserIdBase1",
							"AdminUserIdDev2",
							"AdminUserIdDev1",
						},
						true,
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
					},
					nil,
					nil,
				},
				&Mode{
					ModeCommon{
						"AllowedOriginsBase",
						"DefaultPortBase",
						"DefaultStaticPortBase",
						[]string{
							"AdminUserIdBase1",
							"AdminUserIdDev2",
						},
						false,
						false,
						map[string]string{
							"Config1": "Base",
							"Config2": "Base",
							"Config3": "Base",
						},
						"bolt",
						"gastring",
						nil,
						"https://localhost",
					},
					nil,
					nil,
				},
				nil,
				nil,
				nil,
			},
		},
	}

	for i, test := range tests {

		out := NewConfig(test.in, nil)

		if out == nil {
			continue
		}

		//Chjeat and zero these out so we skip them in the compare
		out.rawSecretConfig = nil
		out.rawPublicConfig = nil
		out.Prod.parentConfig = nil
		out.Dev.parentConfig = nil

		assert.For(t, i, test.description).ThatActual(out).Equals(test.out).ThenDiffOnFail()
	}

}

func TestApiHostDerivation(t *testing.T) {

	tests := []struct {
		description string
		prodMode    bool
		in          *RawConfigMode
		out         *Mode
	}{
		{
			"No op",
			false,
			&RawConfigMode{
				ModeCommon{
					APIHost:     "provided",
					DefaultPort: "8888",
				},
				nil,
			},
			&Mode{
				ModeCommon{
					APIHost:           "provided",
					DefaultPort:       "8888",
					DefaultStaticPort: "8080",
					AllowedOrigins:    "*",
					Storage:           make(map[string]string),
				},
				nil,
				nil,
			},
		},
		{
			"dev",
			false,
			&RawConfigMode{
				ModeCommon{
					DefaultPort: "8888",
				},
				nil,
			},
			&Mode{
				ModeCommon{
					APIHost:           "http://localhost:8888",
					DefaultPort:       "8888",
					DefaultStaticPort: "8080",
					AllowedOrigins:    "*",
					Storage:           make(map[string]string),
				},
				nil,
				nil,
			},
		},
		{
			"prod non default port",
			true,
			&RawConfigMode{
				ModeCommon{
					APIHost:     "",
					DefaultPort: "8080",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
				},
				nil,
			},
			&Mode{
				ModeCommon{
					APIHost:           "https://example-boardgame.appspot.com:8080",
					DefaultPort:       "8080",
					DefaultStaticPort: "80",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
					AllowedOrigins: "*",
					Storage:        make(map[string]string),
				},
				nil,
				nil,
			},
		},
		{
			"prod no default port",
			true,
			&RawConfigMode{
				ModeCommon{
					APIHost: "",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
					DefaultPort: "80",
				},
				nil,
			},
			&Mode{
				ModeCommon{
					APIHost: "https://example-boardgame.appspot.com",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
					DefaultPort:       "80",
					DefaultStaticPort: "80",
					AllowedOrigins:    "*",
					Storage:           make(map[string]string),
				},
				nil,
				nil,
			},
		},
		{
			"prod default port 80",
			true,
			&RawConfigMode{
				ModeCommon{
					APIHost:     "",
					DefaultPort: "80",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
				},
				nil,
			},
			&Mode{
				ModeCommon{
					APIHost:           "https://example-boardgame.appspot.com",
					DefaultPort:       "80",
					DefaultStaticPort: "80",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
					AllowedOrigins: "*",
					Storage:        make(map[string]string),
				},
				nil,
				nil,
			},
		},
	}

	for i, test := range tests {
		out := test.in.Derive(nil, test.prodMode)

		assert.For(t, i, test.description).ThatActual(out).Equals(test.out).ThenDiffOnFail()
	}

}
