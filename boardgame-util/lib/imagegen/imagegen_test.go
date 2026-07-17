package imagegen

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (r roundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return r(req) }

func TestGenerateWritesImageAndManifest(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "asset.png")
	generated := image.NewRGBA(image.Rect(0, 0, 2, 2))
	generated.Set(0, 0, color.RGBA{R: 220, G: 30, B: 20, A: 255})
	var jpegBytes bytes.Buffer
	if err := jpeg.Encode(&jpegBytes, generated, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	client := Client{
		Now: func() time.Time { return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC) },
		HTTPClient: &http.Client{Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
			if got := req.Header.Get("x-goog-api-key"); got != "secret" {
				t.Fatalf("API key header = %q", got)
			}
			body, _ := io.ReadAll(req.Body)
			if strings.Contains(string(body), "secret") {
				t.Fatal("API key leaked into request body")
			}
			response := `{"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/jpeg","data":"` + base64.StdEncoding.EncodeToString(jpegBytes.Bytes()) + `"}}]}}]}`
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(response)), Header: make(http.Header)}, nil
		})},
	}
	manifest, err := client.Generate(context.Background(), Request{
		Prompt: "an original beetle", Output: output, APIKey: "secret", CleanRoom: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if gotMIME := http.DetectContentType(got); gotMIME != "image/png" {
		t.Fatalf("output MIME = %q, want image/png", gotMIME)
	}
	if !manifest.CleanRoom || manifest.OutputMIME != "image/png" || manifest.OutputSHA256 == "" || manifest.PromptSHA256 == "" {
		t.Fatalf("incomplete manifest: %#v", manifest)
	}
	if _, err := os.Stat(output + ".imagegen.json"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateEditRequiresReference(t *testing.T) {
	r := Request{Mode: "edit", Prompt: "change it", Output: "out.png", APIKey: "x"}
	if err := Validate(&r); err == nil || !strings.Contains(err.Error(), "reference") {
		t.Fatalf("expected reference error, got %v", err)
	}
}

func TestStyleSheetPromptHasSafetyBoundary(t *testing.T) {
	prompt := StyleSheetPrompt("Graphite naturalist sketches")
	for _, required := range []string{"original tabletop game", "no logos", "Graphite naturalist sketches"} {
		if !strings.Contains(prompt, required) {
			t.Errorf("prompt missing %q", required)
		}
	}
}
