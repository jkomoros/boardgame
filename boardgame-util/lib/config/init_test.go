package config

import (
	"github.com/workfit/tester/assert"
	"path/filepath"
	"testing"
)

//If upddateGolden is true, save new goldens in testdata
const updateGolden = false

type initialConfigConstructor func(string) *Config

func TestSampleConfigs(t *testing.T) {

	tests := []struct {
		constructor initialConfigConstructor
		filename    string
	}{
		{
			DefaultStarterConfig,
			"default.json",
		},
		{
			MinimalStarterConfig,
			"minimal.json",
		},
		{
			SampleStarterConfig,
			"sample.json",
		},
	}

	for i, test := range tests {

		filename := filepath.Join("testdata", test.filename)

		c := test.constructor(filename)

		if updateGolden {

			if err := c.Save(); err != nil {
				t.Error("Couldn't save " + filename + ": " + err.Error())
			}

			continue
		}

		golden, err := GetConfig(filename, "", false)
		assert.For(t, i).ThatActual(err).IsNil()
		assert.For(t, i).ThatActual(c).Equals(golden).ThenDiffOnFail()

	}

}
