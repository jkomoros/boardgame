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
		//TODO: test simple set secret
		//TODO: test set on nil config mode
		//TODO: test nil updater
		//TODO: test totally nil configs in
		//TODO: test set on nil secret with non-nil public
		//TODO: test set on nil public with nil secret
		//TODO: test set on nil public with non-nil secret
		//TODO: test invalid type
	}

	for i, test := range tests {
		config, err := NewConfig(test.inPublic, test.inSecret)

		assert.For(t, i, test.description).ThatActual(err).IsNil()

		err = config.Update(test.inType, test.inIsSecret, test.inUpdater)

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
