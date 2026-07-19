package imagegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
)

type StyleCandidate struct {
	Slug   string
	Label  string
	Prompt string
}

type StyleLock struct {
	SchemaVersion  int    `json:"schema_version"`
	LockedAt       string `json:"locked_at"`
	Image          string `json:"image"`
	ImageSHA256    string `json:"image_sha256"`
	SelectedFrom   string `json:"selected_from"`
	SourceManifest string `json:"source_manifest,omitempty"`
}

func ExploreStyles(brief string) []StyleCandidate {
	return []StyleCandidate{
		{Slug: "01-observational", Label: "Observational", Prompt: brief + "\n\nSTYLE AXIS: observational and naturalistic; precise forms, restrained palette, scientific credibility, quiet editorial composition."},
		{Slug: "02-graphic", Label: "Graphic", Prompt: brief + "\n\nSTYLE AXIS: bold and graphic; simplified silhouettes, strong value structure, limited palette, clear readability at game-component scale."},
		{Slug: "03-atmospheric", Label: "Atmospheric", Prompt: brief + "\n\nSTYLE AXIS: painterly and atmospheric; expressive light, weather, depth, and color while retaining functional negative space."},
		{Slug: "04-tactile", Label: "Tactile", Prompt: brief + "\n\nSTYLE AXIS: tactile printmaking and handmade materials; visible process, texture, registration, and an ownable physical identity."},
	}
}

func IterateStyles(refinement string) []StyleCandidate {
	return []StyleCandidate{
		{Slug: "01-faithful", Label: "Faithful", Prompt: refinement + "\n\nITERATION AXIS: preserve the selected style very closely; make only the requested refinements."},
		{Slug: "02-bolder", Label: "Bolder", Prompt: refinement + "\n\nITERATION AXIS: preserve the selected identity while increasing contrast, silhouette clarity, and distinctiveness."},
		{Slug: "03-quieter", Label: "Quieter", Prompt: refinement + "\n\nITERATION AXIS: preserve the selected identity while simplifying texture and creating more calm functional negative space."},
		{Slug: "04-production", Label: "Production", Prompt: refinement + "\n\nITERATION AXIS: preserve the selected identity while optimizing consistency, repeatability, and legibility across many game assets."},
	}
}

func WriteGallery(dir, title string, candidates []StyleCandidate) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	const page = `<!doctype html>
<html><head><meta charset="utf-8"><title>{{.Title}}</title>
<style>body{font:16px system-ui;margin:2rem;background:#eee;color:#222}main{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:1.5rem}figure{margin:0;background:white;padding:1rem;border-radius:.5rem;box-shadow:0 2px 12px #0002}img{display:block;width:100%;height:auto}figcaption{font-weight:650;margin-top:.75rem}code{font-weight:400}</style></head>
<body><h1>{{.Title}}</h1><p>Choose one candidate to iterate or lock. Every image has a provenance sidecar.</p><main>
{{range .Candidates}}<figure><img src="{{.Slug}}.png" alt="{{.Label}}"><figcaption>{{.Label}} — <code>{{.Slug}}.png</code></figcaption></figure>{{end}}
</main></body></html>`
	tmpl, err := template.New("gallery").Parse(page)
	if err != nil {
		return err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, struct {
		Title      string
		Candidates []StyleCandidate
	}{title, candidates}); err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(filepath.Join(dir, "gallery.html"), output.Bytes(), 0o644)
}

func CreateStyleLock(selected, output string, force bool, now func() time.Time) (*StyleLock, error) {
	lock, files, hasSourceManifest, err := buildStyleLockFiles(selected, output, now)
	if err != nil {
		return nil, err
	}
	specs := make(map[string]fileutil.FileSpec, len(files)+1)
	for name, contents := range files {
		specs[name] = fileutil.FileSpec{Contents: contents, Mode: 0o644, Exclusive: !force}
	}
	if force && !hasSourceManifest {
		specs[filepath.Base(output)+".imagegen.json"] = fileutil.FileSpec{Delete: true}
	}
	if err := fileutil.WriteFileSetAtomic(filepath.Dir(output), specs, true); err != nil {
		return nil, err
	}
	return lock, nil
}

func buildStyleLockFiles(selected, output string, now func() time.Time) (*StyleLock, map[string][]byte, bool, error) {
	if selected == "" || output == "" {
		return nil, nil, false, errors.New("selected reference and output are required")
	}
	data, err := readLimitedFile(selected, maxImageAssetBytes)
	if err != nil {
		return nil, nil, false, err
	}
	if err := validateImageDimensions(data); err != nil {
		return nil, nil, false, fmt.Errorf("validate selected style image: %w", err)
	}
	sourceManifest := selected + ".imagegen.json"
	manifestData, manifestErr := readLimitedFile(sourceManifest, maxMetadataBytes)
	if os.IsNotExist(manifestErr) {
		sourceManifest = ""
	} else if manifestErr != nil {
		return nil, nil, false, fmt.Errorf("read source image manifest: %w", manifestErr)
	}
	if now == nil {
		now = time.Now
	}
	// The locked image always lives beside this manifest. Store a portable
	// relative reference so a style can move between GAMES and BOARDGAME.
	lock := &StyleLock{SchemaVersion: 1, LockedAt: now().UTC().Format(time.RFC3339), Image: filepath.Base(output), ImageSHA256: digest(data), SelectedFrom: selected, SourceManifest: sourceManifest}
	encoded, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return nil, nil, false, err
	}
	encoded = append(encoded, '\n')
	outputName := filepath.Base(output)
	files := map[string][]byte{
		outputName:                      data,
		outputName + ".style-lock.json": encoded,
	}
	if sourceManifest != "" {
		files[outputName+".imagegen.json"] = manifestData
	}
	return lock, files, sourceManifest != "", nil
}

func ReadStyleLock(path string) (*StyleLock, error) {
	data, err := readLimitedFile(path, maxMetadataBytes)
	if err != nil {
		return nil, err
	}
	var lock StyleLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}
	imagePath := lock.Image
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(filepath.Dir(path), imagePath)
	}
	image, err := readLimitedFile(imagePath, maxImageAssetBytes)
	if err != nil {
		return nil, fmt.Errorf("read locked style image: %w", err)
	}
	if err := validateImageDimensions(image); err != nil {
		return nil, fmt.Errorf("validate locked style image: %w", err)
	}
	if digest(image) != lock.ImageSHA256 {
		return nil, errors.New("locked style image hash does not match the lock manifest")
	}
	lock.Image = imagePath
	return &lock, nil
}
