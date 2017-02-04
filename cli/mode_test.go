package cli

import (
	"testing"
)

func TestOverlayContentValid(t *testing.T) {

	tests := []struct {
		input    *overlayContent
		expected bool
	}{
		{
			input: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"0",
				},
			},
			expected: false,
		},
		{
			input: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"0",
					"1",
				},
			},
			expected: true,
		},
		{
			input:    &overlayContent{},
			expected: true,
		},
	}

	for i, test := range tests {
		if test.input.Valid() != test.expected {
			t.Error("Test", i, "got wrong result. Got", test.input.Valid(), "wanted", test.expected)
		}
	}

}

func TestOverlayContentAligned(t *testing.T) {
	tests := []struct {
		input    *overlayContent
		expected bool
	}{
		{
			input: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"0",
				},
			},
			expected: false,
		},
		{
			input: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"0",
					"1",
				},
			},
			expected: true,
		},
		{
			input: &overlayContent{
				{
					" 0",
					"1",
				},
				{
					"0",
					"1",
				},
			},
			expected: false,
		},
		{
			input: &overlayContent{
				{
					" 0",
					"1",
				},
				{
					"0 ",
					"1",
				},
			},
			expected: true,
		},
	}

	for i, test := range tests {
		if test.input.Aligned() != test.expected {
			t.Error("Test", i, "got wrong result. Got", test.input.Aligned(), "wanted", test.expected)
		}
	}
}
