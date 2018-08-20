package config

import (
	"github.com/workfit/tester/assert"
	"testing"
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
					ConfigModeCommon{
						"AllowedOriginsDev",
						"DefaultPortDev",
						"DefaultStaticPortDev",
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
					},
					nil,
				},
				nil,
				"",
			},
			&Config{
				&ConfigMode{
					ConfigModeCommon{
						"AllowedOriginsDev",
						"DefaultPortDev",
						"DefaultStaticPortDev",
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
					},
					nil,
				},
				&ConfigMode{
					ConfigModeCommon{
						AllowedOrigins: "*",
						DefaultPort:    "8080",
						Storage:        make(map[string]string),
					},
					nil,
				},
				nil,
				nil,
			},
		},
		{
			"Simple derive",
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon{
						"AllowedOriginsBase",
						"DefaultPortBase",
						"DefaultStaticPortBase",
						[]string{
							"AdminUserIdBase1",
							"AdminUserIdDev2",
						},
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
					ConfigModeCommon{
						"AllowedOriginsDev",
						"",
						"",
						[]string{
							"AdminUserIdDev1",
						},
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
				&ConfigMode{
					ConfigModeCommon{
						"AllowedOriginsDev",
						"DefaultPortBase",
						"DefaultStaticPortBase",
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
					},
					nil,
				},
				&ConfigMode{
					ConfigModeCommon{
						"AllowedOriginsBase",
						"DefaultPortBase",
						"DefaultStaticPortBase",
						[]string{
							"AdminUserIdBase1",
							"AdminUserIdDev2",
						},
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

		assert.For(t, i, test.description).ThatActual(out).Equals(test.out).ThenDiffOnFail()
	}

}

func TestApiHostDerivation(t *testing.T) {

	tests := []struct {
		description string
		prodMode    bool
		in          *RawConfigMode
		out         *ConfigMode
	}{
		{
			"No op",
			false,
			&RawConfigMode{
				ConfigModeCommon{
					ApiHost:     "provided",
					DefaultPort: "8888",
				},
				nil,
			},
			&ConfigMode{
				ConfigModeCommon{
					ApiHost:        "provided",
					DefaultPort:    "8888",
					AllowedOrigins: "*",
					Storage:        make(map[string]string),
				},
				nil,
			},
		},
		{
			"dev",
			false,
			&RawConfigMode{
				ConfigModeCommon{
					DefaultPort: "8888",
				},
				nil,
			},
			&ConfigMode{
				ConfigModeCommon{
					ApiHost:        "http://localhost:8888",
					DefaultPort:    "8888",
					AllowedOrigins: "*",
					Storage:        make(map[string]string),
				},
				nil,
			},
		},
		{
			"prod non default port",
			true,
			&RawConfigMode{
				ConfigModeCommon{
					ApiHost:     "",
					DefaultPort: "8080",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
				},
				nil,
			},
			&ConfigMode{
				ConfigModeCommon{
					ApiHost:     "https://example-boardgame.appspot.com:8080",
					DefaultPort: "8080",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
					AllowedOrigins: "*",
					Storage:        make(map[string]string),
				},
				nil,
			},
		},
		{
			"prod no default port",
			true,
			&RawConfigMode{
				ConfigModeCommon{
					ApiHost: "",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
					DefaultPort: "80",
				},
				nil,
			},
			&ConfigMode{
				ConfigModeCommon{
					ApiHost: "https://example-boardgame.appspot.com",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
					DefaultPort:    "80",
					AllowedOrigins: "*",
					Storage:        make(map[string]string),
				},
				nil,
			},
		},
		{
			"prod default port 80",
			true,
			&RawConfigMode{
				ConfigModeCommon{
					ApiHost:     "",
					DefaultPort: "80",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
				},
				nil,
			},
			&ConfigMode{
				ConfigModeCommon{
					ApiHost:     "https://example-boardgame.appspot.com",
					DefaultPort: "80",
					Firebase: &FirebaseConfig{
						StorageBucket: "example-boardgame.appspot.com",
					},
					AllowedOrigins: "*",
					Storage:        make(map[string]string),
				},
				nil,
			},
		},
	}

	for i, test := range tests {
		out := test.in.Derive(test.prodMode)

		assert.For(t, i, test.description).ThatActual(out).Equals(test.out).ThenDiffOnFail()
	}

}
