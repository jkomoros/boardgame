package imagegen

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProduceAlphaKeysAndDespillsMatte(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "matte.png"), filepath.Join(dir, "alpha.png")
	img := image.NewNRGBA(image.Rect(0, 0, 9, 9))
	for y := 0; y < 9; y++ {
		for x := 0; x < 9; x++ {
			img.SetNRGBA(x, y, color.NRGBA{0, 255, 0, 255})
		}
	}
	for y := 3; y <= 5; y++ {
		for x := 3; x <= 5; x++ {
			img.SetNRGBA(x, y, color.NRGBA{30, 30, 30, 255})
		}
	}
	file, err := os.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
	file.Close()
	manifest, err := ProduceAlpha(AlphaOptions{Input: input, Output: output, KeyColor: "#00FF00", CoreTolerance: 20, FeatherTolerance: 100, ChromaGate: 70}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.TransparentPixels == 0 || manifest.OpaquePixels == 0 || manifest.MaxEdgeAlpha != 0 {
		t.Fatalf("bad alpha stats: %#v", manifest)
	}
	decodedFile, _ := os.Open(output)
	defer decodedFile.Close()
	decoded, _ := png.Decode(decodedFile)
	if _, _, _, a := decoded.At(0, 0).RGBA(); a != 0 {
		t.Fatalf("corner alpha=%d", a)
	}
	if _, _, _, a := decoded.At(4, 4).RGBA(); a != 0xffff {
		t.Fatalf("subject alpha=%d", a)
	}
}

func TestMattePromptRequiresUniformKey(t *testing.T) {
	prompt := MattePrompt("an original beetle", "#00FF00")
	for _, want := range []string{"an original beetle", "#00FF00", "perfectly uniform", "production matte"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestContractAlphaErodesMaskByRadius(t *testing.T) {
	mask := []uint8{
		0, 0, 0, 0, 0,
		0, 255, 255, 255, 0,
		0, 255, 255, 255, 0,
		0, 255, 255, 255, 0,
		0, 0, 0, 0, 0,
	}
	contracted := contractAlpha(mask, 5, 5, 1)
	for index, alpha := range contracted {
		want := uint8(0)
		if index == 12 {
			want = 255
		}
		if alpha != want {
			t.Fatalf("alpha[%d]=%d, want %d", index, alpha, want)
		}
	}
}
