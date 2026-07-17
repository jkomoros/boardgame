package imagegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type AlphaOptions struct {
	Input, Output    string
	KeyColor         string
	CoreTolerance    int
	FeatherTolerance int
	ChromaGate       int
}

type AlphaManifest struct {
	SchemaVersion     int    `json:"schema_version"`
	CreatedAt         string `json:"created_at"`
	Input             string `json:"input"`
	InputSHA256       string `json:"input_sha256"`
	Output            string `json:"output"`
	OutputSHA256      string `json:"output_sha256"`
	KeyRequested      string `json:"key_requested"`
	KeyEstimated      string `json:"key_estimated"`
	CoreTolerance     int    `json:"core_tolerance"`
	FeatherTolerance  int    `json:"feather_tolerance"`
	ChromaGate        int    `json:"chroma_gate"`
	TransparentPixels int    `json:"transparent_pixels"`
	PartialPixels     int    `json:"partial_pixels"`
	OpaquePixels      int    `json:"opaque_pixels"`
	MaxEdgeAlpha      uint8  `json:"max_edge_alpha"`
}

func MattePrompt(prompt, key string) string {
	return prompt + `\n\nTRANSPARENCY MATTE — CRITICAL: Render the subject isolated on a perfectly uniform, flat ` + key +
		` chroma-key background filling every pixel outside the subject. No texture, gradient, shadow, floor, border, paper, glow, or environmental color in the background. Keep the subject palette far from the key color. This is a production matte that will be converted deterministically to true alpha.`
}

func ProduceAlpha(options AlphaOptions, now func() time.Time) (*AlphaManifest, error) {
	if options.Input == "" || options.Output == "" {
		return nil, errors.New("alpha production requires input and output paths")
	}
	key, err := parseHexColor(options.KeyColor)
	if err != nil {
		return nil, err
	}
	if options.CoreTolerance == 0 {
		options.CoreTolerance = 40
	}
	if options.FeatherTolerance == 0 {
		options.FeatherTolerance = 100
	}
	if options.ChromaGate == 0 {
		options.ChromaGate = 80
	}
	if options.CoreTolerance < 0 || options.FeatherTolerance <= options.CoreTolerance || options.ChromaGate < 0 {
		return nil, errors.New("alpha tolerances must satisfy 0 <= core < feather and chroma gate >= 0")
	}
	inputBytes, err := os.ReadFile(options.Input)
	if err != nil {
		return nil, err
	}
	source, _, err := image.Decode(bytes.NewReader(inputBytes))
	if err != nil {
		return nil, fmt.Errorf("decode matte: %w", err)
	}
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width < 3 || height < 3 {
		return nil, errors.New("matte is too small")
	}
	estimated := estimateKey(source, key, options.ChromaGate)
	pixels := image.NewNRGBA(image.Rect(0, 0, width, height))
	rawAlpha := make([]uint8, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := source.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			r8, g8, b8 := float64(r>>8), float64(g>>8), float64(b>>8)
			alpha := 255.0
			if chromaProjection(r8, g8, b8, key) >= float64(options.ChromaGate) {
				dr, dg, db := r8-estimated[0], g8-estimated[1], b8-estimated[2]
				distance := math.Sqrt(dr*dr + dg*dg + db*db)
				if distance <= float64(options.CoreTolerance) {
					alpha = 0
				} else if distance < float64(options.FeatherTolerance) {
					alpha = 255 * (distance - float64(options.CoreTolerance)) / float64(options.FeatherTolerance-options.CoreTolerance)
				}
			}
			a := uint8(clampFloat(alpha))
			rawAlpha[y*width+x] = a
			if a >= 16 && a < 255 {
				af := float64(a) / 255
				r8 = clampFloat((r8 - (1-af)*estimated[0]) / af)
				g8 = clampFloat((g8 - (1-af)*estimated[1]) / af)
				b8 = clampFloat((b8 - (1-af)*estimated[2]) / af)
			}
			pixels.SetNRGBA(x, y, color.NRGBA{uint8(r8), uint8(g8), uint8(b8), a})
		}
	}
	smoothed := blurAlpha(rawAlpha, width, height)
	transparent, partial, opaque := 0, 0, 0
	var maxEdge uint8
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			a := smoothed[idx]
			if rawAlpha[idx] == 0 {
				a = 0
			}
			px := pixels.NRGBAAt(x, y)
			px.A = a
			pixels.SetNRGBA(x, y, px)
			switch {
			case a == 0:
				transparent++
			case a == 255:
				opaque++
			default:
				partial++
			}
			if x == 0 || y == 0 || x == width-1 || y == height-1 {
				if a > maxEdge {
					maxEdge = a
				}
			}
		}
	}
	if transparent == 0 || opaque == 0 {
		return nil, errors.New("alpha QA failed: output must contain both transparent background and opaque subject pixels")
	}
	if maxEdge > 8 {
		return nil, fmt.Errorf("alpha QA failed: subject or matte residue touches the frame (edge alpha %d)", maxEdge)
	}
	if err := os.MkdirAll(filepath.Dir(options.Output), 0o755); err != nil {
		return nil, err
	}
	file, err := os.Create(options.Output)
	if err != nil {
		return nil, err
	}
	if err := png.Encode(file, pixels); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	outputBytes, err := os.ReadFile(options.Output)
	if err != nil {
		return nil, err
	}
	if now == nil {
		now = time.Now
	}
	manifest := &AlphaManifest{SchemaVersion: 1, CreatedAt: now().UTC().Format(time.RFC3339), Input: options.Input, InputSHA256: digest(inputBytes), Output: options.Output, OutputSHA256: digest(outputBytes), KeyRequested: strings.ToUpper(options.KeyColor), KeyEstimated: fmt.Sprintf("#%02X%02X%02X", int(estimated[0]), int(estimated[1]), int(estimated[2])), CoreTolerance: options.CoreTolerance, FeatherTolerance: options.FeatherTolerance, ChromaGate: options.ChromaGate, TransparentPixels: transparent, PartialPixels: partial, OpaquePixels: opaque, MaxEdgeAlpha: maxEdge}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	if err := os.WriteFile(options.Output+".alpha.json", data, 0o644); err != nil {
		return nil, err
	}
	return manifest, nil
}

func parseHexColor(value string) ([3]float64, error) {
	if value == "" {
		value = "#00FF00"
	}
	value = strings.TrimPrefix(value, "#")
	if len(value) != 6 {
		return [3]float64{}, errors.New("key color must be six hexadecimal digits")
	}
	parsed, err := strconv.ParseUint(value, 16, 24)
	if err != nil {
		return [3]float64{}, errors.New("key color must be six hexadecimal digits")
	}
	return [3]float64{float64(parsed >> 16), float64((parsed >> 8) & 255), float64(parsed & 255)}, nil
}

func chromaProjection(r, g, b float64, key [3]float64) float64 {
	keyMean := (key[0] + key[1] + key[2]) / 3
	pixelMean := (r + g + b) / 3
	k0, k1, k2 := key[0]-keyMean, key[1]-keyMean, key[2]-keyMean
	norm := math.Sqrt(k0*k0 + k1*k1 + k2*k2)
	if norm == 0 {
		return 0
	}
	return ((r-pixelMean)*k0 + (g-pixelMean)*k1 + (b-pixelMean)*k2) / norm
}

func estimateKey(source image.Image, requested [3]float64, gate int) [3]float64 {
	b := source.Bounds()
	inset := 10
	var sum [3]float64
	count := 0
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if x >= inset && y >= inset && x < b.Dx()-inset && y < b.Dy()-inset {
				continue
			}
			r, g, bl, _ := source.At(b.Min.X+x, b.Min.Y+y).RGBA()
			r8, g8, b8 := float64(r>>8), float64(g>>8), float64(bl>>8)
			if chromaProjection(r8, g8, b8, requested) < float64(gate) {
				continue
			}
			sum[0] += r8
			sum[1] += g8
			sum[2] += b8
			count++
		}
	}
	if count == 0 {
		return requested
	}
	return [3]float64{sum[0] / float64(count), sum[1] / float64(count), sum[2] / float64(count)}
}

func blurAlpha(source []uint8, width, height int) []uint8 {
	tmp, result := make([]uint8, len(source)), make([]uint8, len(source))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			left, right := x-1, x+1
			if left < 0 {
				left = 0
			}
			if right >= width {
				right = width - 1
			}
			tmp[y*width+x] = uint8((int(source[y*width+left]) + 2*int(source[y*width+x]) + int(source[y*width+right])) / 4)
		}
	}
	for y := 0; y < height; y++ {
		top, bottom := y-1, y+1
		if top < 0 {
			top = 0
		}
		if bottom >= height {
			bottom = height - 1
		}
		for x := 0; x < width; x++ {
			result[y*width+x] = uint8((int(tmp[top*width+x]) + 2*int(tmp[y*width+x]) + int(tmp[bottom*width+x])) / 4)
		}
	}
	return result
}

func clampFloat(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return math.Round(value)
}
