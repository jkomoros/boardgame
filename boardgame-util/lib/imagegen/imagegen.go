// Package imagegen provides a reproducible, provider-backed image generation
// pipeline for boardgame projects. It deliberately keeps credentials out of
// project configuration and records enough provenance to review or revise an
// asset later.
package imagegen

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
)

const (
	DefaultEndpoint = "https://generativelanguage.googleapis.com/v1beta"
	DefaultModel    = "gemini-3-pro-image-preview"
)

var validRatios = map[string]bool{
	"1:1": true, "1:4": true, "4:1": true, "1:8": true, "8:1": true,
	"2:3": true, "3:2": true, "3:4": true, "4:3": true, "4:5": true,
	"5:4": true, "9:16": true, "16:9": true, "21:9": true,
}

var validSizes = map[string]bool{"512": true, "1K": true, "2K": true, "4K": true}

type Request struct {
	Mode         string
	Prompt       string
	PromptFile   string
	Output       string
	References   []string
	Model        string
	AspectRatio  string
	ImageSize    string
	Endpoint     string
	APIKey       string
	CleanRoom    bool
	SourceAssets bool
}

type Reference struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	MIME   string `json:"mime_type"`
}

type Manifest struct {
	SchemaVersion int         `json:"schema_version"`
	CreatedAt     string      `json:"created_at"`
	Provider      string      `json:"provider"`
	Model         string      `json:"model"`
	Mode          string      `json:"mode"`
	Prompt        string      `json:"prompt"`
	PromptFile    string      `json:"prompt_file,omitempty"`
	PromptSHA256  string      `json:"prompt_sha256"`
	References    []Reference `json:"references,omitempty"`
	AspectRatio   string      `json:"aspect_ratio"`
	ImageSize     string      `json:"image_size"`
	Output        string      `json:"output"`
	OutputMIME    string      `json:"output_mime_type"`
	OutputSHA256  string      `json:"output_sha256"`
	CleanRoom     bool        `json:"clean_room"`
	SourceAssets  bool        `json:"source_assets_used"`
}

type Client struct {
	HTTPClient *http.Client
	Now        func() time.Time
}

type inlineData struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"`
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type apiRequest struct {
	Contents []struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	} `json:"contents"`
	GenerationConfig struct {
		ResponseModalities []string `json:"responseModalities"`
		ImageConfig        struct {
			AspectRatio string `json:"aspectRatio"`
			ImageSize   string `json:"imageSize"`
		} `json:"imageConfig"`
	} `json:"generationConfig"`
}

type apiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []part `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func Validate(r *Request) error {
	if r.Mode == "" {
		r.Mode = "generate"
	}
	if r.Mode != "generate" && r.Mode != "edit" && r.Mode != "style-sheet" && r.Mode != "matte" {
		return fmt.Errorf("unsupported mode %q", r.Mode)
	}
	if strings.TrimSpace(r.Prompt) == "" {
		return errors.New("prompt is required")
	}
	if r.Output == "" {
		return errors.New("output path is required")
	}
	if r.Mode == "edit" && len(r.References) == 0 {
		return errors.New("edit mode requires at least one reference image")
	}
	if r.Model == "" {
		r.Model = DefaultModel
	}
	if r.AspectRatio == "" {
		r.AspectRatio = "1:1"
	}
	if !validRatios[r.AspectRatio] {
		return fmt.Errorf("unsupported aspect ratio %q", r.AspectRatio)
	}
	if r.ImageSize == "" {
		r.ImageSize = "2K"
	}
	if !validSizes[r.ImageSize] {
		return fmt.Errorf("unsupported image size %q", r.ImageSize)
	}
	if r.Endpoint == "" {
		r.Endpoint = DefaultEndpoint
	}
	if r.APIKey == "" {
		return errors.New("GEMINI_API_KEY is not set")
	}
	return nil
}

func StyleSheetPrompt(brief string) string {
	return `Create a reusable visual style reference sheet for an original tabletop game. ` +
		`Arrange clearly separated examples of palette, line treatment, texture, lighting, ` +
		`materials, environments, characters or creatures, and UI ornament. Include no logos, ` +
		`brand names, copyrighted characters, card layouts, or readable prose. This sheet is a ` +
		`production reference, not a game component. Follow this original art direction:\n\n` + brief
}

func (c Client) Generate(ctx context.Context, r Request) (*Manifest, error) {
	if err := Validate(&r); err != nil {
		return nil, err
	}
	if r.Mode == "style-sheet" {
		r.Prompt = StyleSheetPrompt(r.Prompt)
	} else if r.Mode == "matte" {
		r.Prompt = MattePrompt(r.Prompt, "#00FF00")
	}

	references, apiParts, err := loadReferences(r.References)
	if err != nil {
		return nil, err
	}
	apiParts = append([]part{{Text: r.Prompt}}, apiParts...)

	var payload apiRequest
	payload.Contents = make([]struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}, 1)
	payload.Contents[0].Role = "user"
	payload.Contents[0].Parts = apiParts
	payload.GenerationConfig.ResponseModalities = []string{"IMAGE"}
	payload.GenerationConfig.ImageConfig.AspectRatio = r.AspectRatio
	payload.GenerationConfig.ImageConfig.ImageSize = r.ImageSize
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := strings.TrimRight(r.Endpoint, "/") + "/models/" + r.Model + ":generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", r.APIKey)
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Gemini request failed: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	var decoded apiResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return nil, fmt.Errorf("Gemini returned invalid JSON (HTTP %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(decoded.Error.Message)
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return nil, fmt.Errorf("Gemini returned HTTP %d: %s", resp.StatusCode, message)
	}
	imageBytes, responseMIME, err := firstImage(decoded)
	if err != nil {
		return nil, err
	}
	imageBytes, outputMIME, err := normalizeImage(imageBytes, responseMIME, r.Output)
	if err != nil {
		return nil, err
	}
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	manifest := &Manifest{
		SchemaVersion: 1, CreatedAt: now().UTC().Format(time.RFC3339), Provider: "Google Gemini API",
		Model: r.Model, Mode: r.Mode, Prompt: r.Prompt, PromptFile: r.PromptFile,
		PromptSHA256: digest([]byte(r.Prompt)), References: references,
		AspectRatio: r.AspectRatio, ImageSize: r.ImageSize, Output: r.Output,
		OutputMIME: outputMIME, OutputSHA256: digest(imageBytes), CleanRoom: r.CleanRoom, SourceAssets: r.SourceAssets,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	manifestBytes = append(manifestBytes, '\n')
	root := filepath.Dir(r.Output)
	outputName := filepath.Base(r.Output)
	if err := fileutil.WriteFilesAtomic(root, map[string][]byte{
		outputName:                    imageBytes,
		outputName + ".imagegen.json": manifestBytes,
	}, true, 0o644); err != nil {
		return nil, err
	}
	return manifest, nil
}

func loadReferences(paths []string) ([]Reference, []part, error) {
	refs := make([]Reference, 0, len(paths))
	parts := make([]part, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read reference %s: %w", path, err)
		}
		mimeType := http.DetectContentType(data)
		if !strings.HasPrefix(mimeType, "image/") {
			mimeType = mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
		}
		if !strings.HasPrefix(mimeType, "image/") {
			return nil, nil, fmt.Errorf("reference %s is not an image", path)
		}
		refs = append(refs, Reference{Path: path, SHA256: digest(data), MIME: mimeType})
		parts = append(parts, part{InlineData: &inlineData{MIMEType: mimeType, Data: base64.StdEncoding.EncodeToString(data)}})
	}
	return refs, parts, nil
}

func firstImage(response apiResponse) ([]byte, string, error) {
	for _, candidate := range response.Candidates {
		for _, p := range candidate.Content.Parts {
			if p.InlineData != nil && strings.HasPrefix(p.InlineData.MIMEType, "image/") {
				result, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
				if err != nil {
					return nil, "", fmt.Errorf("decode generated image: %w", err)
				}
				return result, p.InlineData.MIMEType, nil
			}
		}
	}
	return nil, "", errors.New("Gemini returned no image")
}

// normalizeImage makes the bytes agree with the requested output extension.
// Gemini may return JPEG data even when callers request a .png path.
func normalizeImage(data []byte, reportedMIME, output string) ([]byte, string, error) {
	targetMIME := mime.TypeByExtension(strings.ToLower(filepath.Ext(output)))
	if separator := strings.IndexByte(targetMIME, ';'); separator >= 0 {
		targetMIME = targetMIME[:separator]
	}
	detectedMIME := http.DetectContentType(data)
	if targetMIME == "" {
		if strings.HasPrefix(detectedMIME, "image/") {
			return data, detectedMIME, nil
		}
		return data, reportedMIME, nil
	}
	if detectedMIME == targetMIME {
		return data, targetMIME, nil
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decode Gemini image for %s output: %w", targetMIME, err)
	}
	var encoded bytes.Buffer
	switch targetMIME {
	case "image/png":
		err = png.Encode(&encoded, decoded)
	case "image/jpeg":
		err = jpeg.Encode(&encoded, decoded, &jpeg.Options{Quality: 95})
	default:
		return nil, "", fmt.Errorf("Gemini returned %s data; cannot convert it to requested %s output", detectedMIME, targetMIME)
	}
	if err != nil {
		return nil, "", fmt.Errorf("encode generated %s: %w", targetMIME, err)
	}
	return encoded.Bytes(), targetMIME, nil
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
