package cli

import (
	"reflect"
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

func TestOverlayContentPadWithAlignment(t *testing.T) {
	tests := []struct {
		input          *overlayContent
		alignment      []columnAlignment
		expected       *overlayContent
		expectedWidths []int
	}{
		{
			input: &overlayContent{
				{
					"00",
					"1",
				},
				{
					"0",
					"1",
				},
			},
			alignment: []columnAlignment{
				alignLeft,
				alignLeft,
			},
			expected: &overlayContent{
				{
					"00",
					"1",
				},
				{
					"0 ",
					"1",
				},
			},
			expectedWidths: []int{
				2,
				1,
			},
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
			alignment: []columnAlignment{
				alignLeft,
				alignLeft,
			},
			expected: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"0",
					"1",
				},
			},
			expectedWidths: []int{
				1,
				1,
			},
		},
		{
			input: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"    0",
					"1   ",
				},
			},
			alignment: []columnAlignment{
				alignRight,
				alignLeft,
			},
			expected: &overlayContent{
				{
					"    0",
					"1   ",
				},
				{
					"    0",
					"1   ",
				},
			},
			expectedWidths: []int{
				5,
				4,
			},
		},
		{
			input: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"    0",
					"1   ",
				},
			},
			alignment: []columnAlignment{
				alignRight,
			},
			expected: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"    0",
					"1   ",
				},
			},
			expectedWidths: []int{
				5,
				4,
			},
		},
	}

	for i, test := range tests {
		test.input.PadWithAlignment(test.alignment...)

		if !reflect.DeepEqual(*test.input, *test.expected) {
			t.Error("Mismatch in test", i, "got", test.input, "wanted", test.expected)
		}

		widths := test.input.ColumnWidths()

		if !reflect.DeepEqual(widths, test.expectedWidths) {
			t.Error("Width mismatch in test", i, "got", widths, "wanted", test.expectedWidths)
		}
	}
}

func TestOverlayContentDefaultPad(t *testing.T) {
	tests := []struct {
		input    *overlayContent
		expected *overlayContent
	}{
		{
			input: &overlayContent{
				{
					"0",
					"1",
				},
				{
					"  0",
					"  1",
				},
			},
			expected: &overlayContent{
				{
					"0  ",
					"1  ",
				},
				{
					"  0",
					"  1",
				},
			},
		},
	}

	for i, test := range tests {
		test.input.DefaultPad()

		if !reflect.DeepEqual(*test.input, *test.expected) {
			t.Error("Mismatch in test", i, "got", test.input, "expected", test.expected)
		}
	}
}

func TestOverlayContentString(t *testing.T) {
	tests := []struct {
		input    *overlayContent
		expected string
	}{
		{
			input: &overlayContent{
				{
					"0",
					"1 ",
				}, {
					"0",
					"1",
				},
			},
			expected: "01 \n01 ",
		},
	}

	for i, test := range tests {
		str := test.input.String()

		if str != test.expected {
			t.Error("Mismatch in test", i, "got", str, "expected", test.expected)
		}
	}
}
