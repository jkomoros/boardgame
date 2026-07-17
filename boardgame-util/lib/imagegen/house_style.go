package imagegen

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

var houseStyleSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type HouseStyle struct {
	SchemaVersion int    `json:"schema_version"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Brief         string `json:"brief"`
	StyleLock     string `json:"style_lock"`
	ImageSHA256   string `json:"image_sha256"`
}

type HouseStyleCatalog struct {
	SchemaVersion int          `json:"schema_version"`
	Styles        []HouseStyle `json:"styles"`
}

func AddHouseStyle(root, name, slug, briefPath, sourceLockPath string, force bool, now func() time.Time) (*HouseStyle, error) {
	if root == "" {
		root = filepath.Join("art", "house-styles")
	}
	if name == "" || !houseStyleSlug.MatchString(slug) {
		return nil, errors.New("house style requires a name and a lowercase hyphenated slug")
	}
	brief, err := os.ReadFile(briefPath)
	if err != nil {
		return nil, err
	}
	sourceLock, err := ReadStyleLock(sourceLockPath)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, slug)
	if !force {
		if _, err := os.Stat(filepath.Join(dir, "house-style.json")); err == nil {
			return nil, errors.New("house style already exists; use --force to replace it")
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "style.md"), brief, 0o644); err != nil {
		return nil, err
	}
	lockedImage := filepath.Join(dir, "style.png")
	lock, err := CreateStyleLock(sourceLock.Image, lockedImage, force, now)
	if err != nil {
		return nil, err
	}
	entry := &HouseStyle{SchemaVersion: 1, Name: name, Slug: slug, Brief: filepath.Join(slug, "style.md"), StyleLock: filepath.Join(slug, "style.png.style-lock.json"), ImageSHA256: lock.ImageSHA256}
	if err := writeJSON(filepath.Join(dir, "house-style.json"), entry); err != nil {
		return nil, err
	}
	if err := rebuildHouseStyleCatalog(root); err != nil {
		return nil, err
	}
	return entry, nil
}

func rebuildHouseStyleCatalog(root string) error {
	dirs, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	styles := make([]HouseStyle, 0, len(dirs))
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, dir.Name(), "house-style.json"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		var style HouseStyle
		if err := json.Unmarshal(data, &style); err != nil {
			return err
		}
		styles = append(styles, style)
	}
	sort.Slice(styles, func(i, j int) bool { return styles[i].Slug < styles[j].Slug })
	return writeJSON(filepath.Join(root, "catalog.json"), HouseStyleCatalog{SchemaVersion: 1, Styles: styles})
}

func ResolveHouseStyle(input, root string) (*StyleLock, error) {
	if input == "" {
		return nil, errors.New("house style is required")
	}
	path := input
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, "style.png.style-lock.json")
	} else if err != nil && !filepath.IsAbs(input) {
		if root == "" {
			root = filepath.Join("art", "house-styles")
		}
		path = filepath.Join(root, input, "style.png.style-lock.json")
	}
	return ReadStyleLock(path)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
