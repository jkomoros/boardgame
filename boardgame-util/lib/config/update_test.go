package config

import (
	"github.com/workfit/tester/assert"
	"path/filepath"
	"testing"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		description    string
		inPublic       *RawConfig
		inSecret       *RawConfig
		inType         ConfigModeType
		inIsSecret     bool
		inUpdater      ConfigUpdater
		errExpected    bool
		expectedPublic *RawConfig
		expectedSecret *RawConfig
	}{
		{
			"Simple public",
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", publicConfigFileName),
			},
			nil,
			TypeBase,
			false,
			SetString("allowedORIGINs ", "after"),
			false,
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "after",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", publicConfigFileName),
			},
			nil,
		},
		{
			"Simple secret",
			nil,
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", privateConfigFileName),
			},
			TypeBase,
			true,
			SetString("allowedORIGINs ", "after"),
			false,
			nil,
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "after",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", privateConfigFileName),
			},
		},
		{
			"Set on nil mode",
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", publicConfigFileName),
			},
			nil,
			TypeDev,
			false,
			SetString("allowedORIGINs ", "after"),
			false,
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "after",
					},
				},
				nil,
				filepath.Join("folder", publicConfigFileName),
			},
			nil,
		},
		{
			"Nil updater",
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", publicConfigFileName),
			},
			nil,
			TypeBase,
			false,
			SetString("NOTAPROPERTs ", "after"),
			true,
			nil,
			nil,
		},
		{
			"Public on fully nil config",
			nil,
			nil,
			TypeBase,
			false,
			SetString("allowedORIGINs ", "after"),
			false,
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "after",
					},
				},
				nil,
				nil,
				publicConfigFileName,
			},
			nil,
		},
		{
			"Set on nil secret with non-nil public",
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", publicConfigFileName),
			},
			nil,
			TypeBase,
			true,
			SetString("allowedORIGINs ", "after"),
			false,
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", publicConfigFileName),
			},
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "after",
					},
				},
				nil,
				nil,
				filepath.Join("folder", privateConfigFileName),
			},
		},
		{
			"Set on nil public with non-nil secret",
			nil,
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", privateConfigFileName),
			},
			TypeBase,
			false,
			SetString("allowedORIGINs ", "after"),
			false,
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "after",
					},
				},
				nil,
				nil,
				filepath.Join("folder", publicConfigFileName),
			},
			&RawConfig{
				&RawConfigMode{
					ConfigModeCommon: ConfigModeCommon{
						AllowedOrigins: "before",
						DefaultPort:    "before",
					},
				},
				nil,
				nil,
				filepath.Join("folder", privateConfigFileName),
			},
		},
	}

	for i, test := range tests {
		config := NewConfig(test.inPublic, test.inSecret)

		err := config.Update(test.inType, test.inIsSecret, test.inUpdater)

		if test.errExpected {
			assert.For(t, i, test.description).ThatActual(err).IsNotNil()
			continue
		} else {
			assert.For(t, i, test.description).ThatActual(err).IsNil()
		}

		assert.For(t, i, test.description).ThatActual(config.RawConfig()).Equals(test.expectedPublic)
		assert.For(t, i, test.description).ThatActual(config.RawSecretConfig()).Equals(test.expectedSecret)

	}
}
