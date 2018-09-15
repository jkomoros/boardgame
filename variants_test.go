package boardgame

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestVariants(t *testing.T) {
	tests := []struct {
		description string
		in          VariantConfig
		errExpected bool
		expected    VariantConfig
	}{
		{
			"Basic expansion no default, one nil value",
			VariantConfig{
				"color": &VariantKey{
					VariantDisplayInfo: VariantDisplayInfo{
						Description: "The color of the thing",
					},
					Values: map[string]*VariantDisplayInfo{
						"blue": {
							Description: "The color blue",
						},
						"red": nil,
					},
				},
			},
			false,
			VariantConfig{
				"color": &VariantKey{
					VariantDisplayInfo: VariantDisplayInfo{
						Name:        "color",
						DisplayName: "Color",
						Description: "The color of the thing",
					},
					Values: map[string]*VariantDisplayInfo{
						"blue": {
							Name:        "blue",
							DisplayName: "Blue",
							Description: "The color blue",
						},
						"red": {
							Name:        "red",
							DisplayName: "Red",
						},
					},
				},
			},
		},
		{
			"Basic expansion with provided display name",
			VariantConfig{
				"color": &VariantKey{
					Values: map[string]*VariantDisplayInfo{
						"blue": {
							Description: "The color blue",
							DisplayName: "Override Blue",
						},
					},
				},
			},
			false,
			VariantConfig{
				"color": &VariantKey{
					VariantDisplayInfo: VariantDisplayInfo{
						Name:        "color",
						DisplayName: "Color",
					},
					Values: map[string]*VariantDisplayInfo{
						"blue": {
							Name:        "blue",
							Description: "The color blue",
							DisplayName: "Override Blue",
						},
					},
				},
			},
		},
		{
			"Basic expansion with multi word name",
			VariantConfig{
				"color": &VariantKey{
					Values: map[string]*VariantDisplayInfo{
						"long-blue-name": nil,
					},
				},
			},
			false,
			VariantConfig{
				"color": &VariantKey{
					VariantDisplayInfo: VariantDisplayInfo{
						Name:        "color",
						DisplayName: "Color",
					},
					Values: map[string]*VariantDisplayInfo{
						"long-blue-name": {
							Name:        "long-blue-name",
							DisplayName: "Long Blue Name",
						},
					},
				},
			},
		},
		{
			"Invalid default",
			VariantConfig{
				"color": &VariantKey{
					Default: "green",
					Values: map[string]*VariantDisplayInfo{
						"blue": nil,
					},
				},
			},
			true,
			nil,
		},
		{
			"OK default name",
			VariantConfig{
				"color": &VariantKey{
					Default: "blue",
					Values: map[string]*VariantDisplayInfo{
						"blue": nil,
					},
				},
			},
			false,
			VariantConfig{
				"color": &VariantKey{
					VariantDisplayInfo: VariantDisplayInfo{
						Name:        "color",
						DisplayName: "Color",
					},
					Default: "blue",
					Values: map[string]*VariantDisplayInfo{
						"blue": {
							Name:        "blue",
							DisplayName: "Blue",
						},
					},
				},
			},
		},
		{
			"Name mismatch key (should work because SetName will overwrite wrong name",
			VariantConfig{
				"color": &VariantKey{
					VariantDisplayInfo: VariantDisplayInfo{
						Name: "not color",
					},
					Values: map[string]*VariantDisplayInfo{
						"blue": nil,
					},
				},
			},
			false,
			VariantConfig{
				"color": &VariantKey{
					VariantDisplayInfo: VariantDisplayInfo{
						Name:        "color",
						DisplayName: "Color",
					},
					Values: map[string]*VariantDisplayInfo{
						"blue": {
							Name:        "blue",
							DisplayName: "Blue",
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		test.in.Initalize()

		err := test.in.Valid()

		if test.errExpected {
			assert.For(t, i, test.description).ThatActual(err).IsNotNil()
			continue
		}

		assert.For(t, i, test.description).ThatActual(err).IsNil()
		assert.For(t, i, test.description).ThatActual(test.in).Equals(test.expected).ThenDiffOnFail()
	}

}
