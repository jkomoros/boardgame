package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/imagegen"
)

type imagegenCmd struct {
	baseSubCommand

	Prompt           string
	PromptFile       string
	Output           string
	OutputDir        string
	References       []string
	StyleLock        string
	HouseStyle       string
	HouseStylesDir   string
	StyleName        string
	Slug             string
	Model            string
	AspectRatio      string
	ImageSize        string
	Endpoint         string
	SecretFile       string
	SecretField      string
	SourceAssets     bool
	Force            bool
	KeyColor         string
	CoreTolerance    int
	FeatherTolerance int
	ChromaGate       int
}

func (i *imagegenCmd) Name() string { return "imagegen" }

func (i *imagegenCmd) Description() string {
	return "Generate, edit, and document original game art with Gemini"
}

func (i *imagegenCmd) Usage() string { return "ACTION [LFS-PATTERN]..." }

func (i *imagegenCmd) HelpText() string {
	return `imagegen is a reproducible art pipeline for board-game projects.

ACTION is style-explore, style-iterate, style-lock, house-style-add, generate,
edit, style-sheet, matte-generate, alpha, or lfs-init. Generation writes OUTPUT
and OUTPUT.imagegen.json, containing the exact prompt, model, reference hashes,
settings, output hash, and clean-room declaration. edit requires one or more
--reference images. style-sheet adds a standard production-reference layout to
the supplied original art direction. Credentials come from GEMINI_API_KEY or
dev.gemini_api_key in config.SECRET.json and are never written to provenance.

style-explore creates four meaningfully different style candidates and a
gallery. style-iterate uses a selected --reference and creates four narrower
variants. style-lock copies the selected candidate to --output and writes a
hash-verified lock manifest. Later generation can use --style-lock directly.
house-style-add promotes a locked style and its prose brief into BOARDGAME's
shared, deterministic house-style catalog; later games use --house-style.
matte-generate produces an isolated subject on a strict chroma field. alpha
turns that preserved matte into a true-alpha PNG with border key estimation,
feathering, despill, edge QA, and a deterministic production manifest.

lfs-init installs Git LFS for the current repository and tracks the remaining
positional patterns. With no patterns it tracks PNG, JPEG, and WebP files.

Examples:
  boardgame-util imagegen style-explore --prompt-file art/style.md --output-dir art/round-1
  boardgame-util imagegen style-iterate --prompt 'Keep the tactile linework; simplify color' --reference art/round-1/04-tactile.png --output-dir art/round-2
  boardgame-util imagegen style-lock --reference art/round-2/01-faithful.png --output art/style.png
  boardgame-util imagegen house-style-add --name 'Field Archive' --slug field-archive --prompt-file art/style.md --style-lock art/style.png.style-lock.json
  boardgame-util imagegen generate --prompt-file art/card.md --style-lock art/style.png.style-lock.json --output art/card.png
  boardgame-util imagegen matte-generate --prompt-file art/token.md --output art/token-matte.png
  boardgame-util imagegen alpha --reference art/token-matte.png --output art/token.png
  boardgame-util imagegen edit --prompt 'make the background cooler' --reference art/card.png --output art/card-v2.png
  boardgame-util imagegen lfs-init 'art/generated/**'`
}

func (i *imagegenCmd) WritOptions() []*writ.Option {
	return []*writ.Option{
		{Names: []string{"prompt"}, Description: "Prompt text. Mutually exclusive with --prompt-file.", Decoder: writ.NewOptionDecoder(&i.Prompt)},
		{Names: []string{"prompt-file"}, Description: "UTF-8 file containing the prompt or style brief.", Decoder: writ.NewOptionDecoder(&i.PromptFile)},
		{Names: []string{"output", "o"}, Description: "Generated image path; provenance is written beside it.", Decoder: writ.NewOptionDecoder(&i.Output)},
		{Names: []string{"output-dir"}, Description: "Directory for four style candidates and gallery.html.", Decoder: writ.NewOptionDecoder(&i.OutputDir)},
		{Names: []string{"reference", "r"}, Description: "Reference image path. Repeat for multiple references.", Decoder: writ.NewOptionDecoder(&i.References)},
		{Names: []string{"style-lock"}, Description: "Hash-verified style-lock manifest to use as a reference.", Decoder: writ.NewOptionDecoder(&i.StyleLock)},
		{Names: []string{"house-style"}, Description: "House-style slug, directory, or lock manifest to use as a reference.", Decoder: writ.NewOptionDecoder(&i.HouseStyle)},
		{Names: []string{"house-styles-dir"}, Description: "House-style catalog root. Defaults to art/house-styles.", Decoder: writ.NewOptionDecoder(&i.HouseStylesDir)},
		{Names: []string{"name"}, Description: "Human-readable house-style name.", Decoder: writ.NewOptionDecoder(&i.StyleName)},
		{Names: []string{"slug"}, Description: "Lowercase hyphenated house-style identifier.", Decoder: writ.NewOptionDecoder(&i.Slug)},
		{Names: []string{"model"}, Description: "Gemini image model.", Decoder: writ.NewOptionDecoder(&i.Model)},
		{Names: []string{"aspect-ratio"}, Description: "Output aspect ratio, such as 1:1, 3:2, or 16:9.", Decoder: writ.NewOptionDecoder(&i.AspectRatio)},
		{Names: []string{"image-size"}, Description: "Output size: 512, 1K, 2K, or 4K.", Decoder: writ.NewOptionDecoder(&i.ImageSize)},
		{Names: []string{"endpoint"}, Description: "Gemini API base URL, primarily for testing.", Decoder: writ.NewOptionDecoder(&i.Endpoint)},
		{Names: []string{"secret-file"}, Description: "JSON secret file. Defaults to config.SECRET.json when present.", Decoder: writ.NewOptionDecoder(&i.SecretFile)},
		{Names: []string{"secret-field"}, Description: "Dot-separated JSON key path. Defaults to dev.gemini_api_key.", Decoder: writ.NewOptionDecoder(&i.SecretField)},
		{Names: []string{"source-assets"}, Description: "Record that source-game assets were used. Omit for clean-room generation.", Decoder: writ.NewFlagDecoder(&i.SourceAssets), Flag: true},
		{Names: []string{"force", "f"}, Description: "Replace an existing locked style output.", Decoder: writ.NewFlagDecoder(&i.Force), Flag: true},
		{Names: []string{"key-color"}, Description: "Six-digit chroma key; defaults to #00FF00.", Decoder: writ.NewOptionDecoder(&i.KeyColor)},
		{Names: []string{"core-tolerance"}, Description: "Distance from the estimated key that becomes fully transparent.", Decoder: writ.NewOptionDecoder(&i.CoreTolerance)},
		{Names: []string{"feather-tolerance"}, Description: "Outer key distance for the partial-alpha ramp.", Decoder: writ.NewOptionDecoder(&i.FeatherTolerance)},
		{Names: []string{"chroma-gate"}, Description: "Minimum key-direction chroma needed before a pixel is eligible for removal.", Decoder: writ.NewOptionDecoder(&i.ChromaGate)},
	}
}

func (i *imagegenCmd) Run(p writ.Path, positional []string) {
	if len(positional) == 0 {
		p.Last().ExitHelp(errors.New("ACTION is required"))
	}
	action := positional[0]
	if action == "lfs-init" {
		if err := initLFS(positional[1:]); err != nil {
			i.Base().errAndQuit("Couldn't configure Git LFS: " + err.Error())
		}
		fmt.Println("Git LFS is configured; commit .gitattributes with the project")
		return
	}
	if action == "style-lock" {
		if len(i.References) != 1 {
			i.Base().errAndQuit("style-lock requires exactly one --reference candidate")
		}
		lock, err := imagegen.CreateStyleLock(i.References[0], i.Output, i.Force, nil)
		if err != nil {
			i.Base().errAndQuit("Couldn't lock style: " + err.Error())
		}
		fmt.Printf("Locked %s (%s) from %s\n", lock.Image, lock.ImageSHA256, lock.SelectedFrom)
		return
	}
	if action == "house-style-add" {
		if i.StyleLock == "" || i.PromptFile == "" {
			i.Base().errAndQuit("house-style-add requires --style-lock and --prompt-file")
		}
		entry, err := imagegen.AddHouseStyle(i.HouseStylesDir, i.StyleName, i.Slug, i.PromptFile, i.StyleLock, i.Force, nil)
		if err != nil {
			i.Base().errAndQuit("Couldn't add house style: " + err.Error())
		}
		fmt.Printf("Added house style %s (%s); commit its prose, lock, provenance, and LFS image\n", entry.Name, entry.Slug)
		return
	}
	if action == "alpha" {
		if len(i.References) != 1 {
			i.Base().errAndQuit("alpha requires exactly one --reference matte")
		}
		manifest, err := imagegen.ProduceAlpha(imagegen.AlphaOptions{Input: i.References[0], Output: i.Output, KeyColor: i.KeyColor, CoreTolerance: i.CoreTolerance, FeatherTolerance: i.FeatherTolerance, ChromaGate: i.ChromaGate}, nil)
		if err != nil {
			i.Base().errAndQuit("Alpha production failed: " + err.Error())
		}
		fmt.Printf("Wrote %s and alpha provenance (%d transparent, %d partial, %d opaque pixels)\n", manifest.Output, manifest.TransparentPixels, manifest.PartialPixels, manifest.OpaquePixels)
		return
	}
	if len(positional) != 1 {
		p.Last().ExitHelp(errors.New("generation actions do not accept positional arguments"))
	}
	if i.Prompt != "" && i.PromptFile != "" {
		i.Base().errAndQuit("Use either --prompt or --prompt-file, not both")
	}
	prompt := i.Prompt
	if i.PromptFile != "" {
		data, err := os.ReadFile(i.PromptFile)
		if err != nil {
			i.Base().errAndQuit("Couldn't read prompt file: " + err.Error())
		}
		prompt = string(data)
	}
	if i.StyleLock != "" {
		lock, err := imagegen.ReadStyleLock(i.StyleLock)
		if err != nil {
			i.Base().errAndQuit("Couldn't read style lock: " + err.Error())
		}
		i.References = append([]string{lock.Image}, i.References...)
	}
	if i.HouseStyle != "" {
		lock, err := imagegen.ResolveHouseStyle(i.HouseStyle, i.HouseStylesDir)
		if err != nil {
			i.Base().errAndQuit("Couldn't read house style: " + err.Error())
		}
		i.References = append([]string{lock.Image}, i.References...)
	}
	apiKey, err := imagegen.ResolveAPIKey(
		os.Getenv("GEMINI_API_KEY"),
		imagegenSecretFile(i.SecretFile),
		imagegenSecretField(i.SecretField),
	)
	if err != nil {
		i.Base().errAndQuit("Couldn't load Gemini credential: " + err.Error())
	}
	if action == "style-explore" || action == "style-iterate" {
		if i.OutputDir == "" {
			i.Base().errAndQuit("style exploration requires --output-dir")
		}
		var candidates []imagegen.StyleCandidate
		if action == "style-explore" {
			candidates = imagegen.ExploreStyles(prompt)
		} else {
			if len(i.References) == 0 {
				i.Base().errAndQuit("style-iterate requires a selected --reference or --style-lock")
			}
			candidates = imagegen.IterateStyles(prompt)
		}
		for _, candidate := range candidates {
			output := filepath.Join(i.OutputDir, candidate.Slug+".png")
			manifest, err := (imagegen.Client{}).Generate(context.Background(), imagegen.Request{
				Mode: "style-sheet", Prompt: candidate.Prompt, PromptFile: i.PromptFile,
				Output: output, References: i.References, Model: i.Model,
				AspectRatio: i.AspectRatio, ImageSize: i.ImageSize, Endpoint: i.Endpoint,
				APIKey: apiKey, CleanRoom: !i.SourceAssets,
				SourceAssets: i.SourceAssets,
			})
			if err != nil {
				i.Base().errAndQuit("Style generation failed for " + candidate.Label + ": " + err.Error())
			}
			fmt.Printf("Wrote %s (%s)\n", manifest.Output, candidate.Label)
		}
		if err := imagegen.WriteGallery(i.OutputDir, "Image style candidates", candidates); err != nil {
			i.Base().errAndQuit("Couldn't write style gallery: " + err.Error())
		}
		fmt.Printf("Review %s, then iterate or lock one candidate\n", filepath.Join(i.OutputDir, "gallery.html"))
		return
	}
	mode := action
	if action == "matte-generate" {
		mode = "matte"
	}
	manifest, err := (imagegen.Client{}).Generate(context.Background(), imagegen.Request{
		Mode: mode, Prompt: prompt, PromptFile: i.PromptFile, Output: i.Output,
		References: i.References, Model: i.Model, AspectRatio: i.AspectRatio,
		ImageSize: i.ImageSize, Endpoint: i.Endpoint, APIKey: apiKey,
		CleanRoom: !i.SourceAssets, SourceAssets: i.SourceAssets,
	})
	if err != nil {
		i.Base().errAndQuit("Image generation failed: " + err.Error())
	}
	fmt.Printf("Wrote %s and %s.imagegen.json with model %s\n", manifest.Output, manifest.Output, manifest.Model)
}

func imagegenSecretFile(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if value := os.Getenv("BOARDGAME_IMAGEGEN_SECRET_FILE"); value != "" {
		return value
	}
	const local = "config.SECRET.json"
	if _, err := os.Stat(local); err == nil {
		return local
	}
	return ""
}

func imagegenSecretField(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if value := os.Getenv("BOARDGAME_IMAGEGEN_SECRET_FIELD"); value != "" {
		return value
	}
	return "dev.gemini_api_key"
}

func initLFS(patterns []string) error {
	if len(patterns) == 0 {
		patterns = []string{"*.png", "*.jpg", "*.jpeg", "*.webp"}
	}
	if err := runGit("rev-parse", "--show-toplevel"); err != nil {
		return errors.New("current directory is not inside a Git repository")
	}
	if err := runGit("lfs", "version"); err != nil {
		return errors.New("git-lfs is not installed")
	}
	if err := runGit("lfs", "install", "--local"); err != nil {
		return err
	}
	for _, pattern := range patterns {
		if strings.TrimSpace(pattern) == "" {
			return errors.New("LFS pattern may not be empty")
		}
		if err := runGit("lfs", "track", pattern); err != nil {
			return err
		}
	}
	return nil
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
