package imagegen

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
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
	lockedImage := filepath.Join(dir, "style.png")
	lock, lockFiles, hasSourceManifest, err := buildStyleLockFiles(sourceLock.Image, lockedImage, now)
	if err != nil {
		return nil, err
	}
	entry := &HouseStyle{SchemaVersion: 1, Name: name, Slug: slug, Brief: filepath.Join(slug, "style.md"), StyleLock: filepath.Join(slug, "style.png.style-lock.json"), ImageSHA256: lock.ImageSHA256}
	entryData, err := marshalJSON(entry)
	if err != nil {
		return nil, err
	}
	catalogData, err := houseStyleCatalogData(root, entry)
	if err != nil {
		return nil, err
	}
	specs := map[string]fileutil.FileSpec{
		filepath.Join(slug, "style.md"):         {Contents: brief, Mode: 0o644, Exclusive: !force},
		filepath.Join(slug, "house-style.json"): {Contents: entryData, Mode: 0o644, Exclusive: !force},
		"catalog.json":                          {Contents: catalogData, Mode: 0o644},
	}
	for name, contents := range lockFiles {
		specs[filepath.Join(slug, name)] = fileutil.FileSpec{Contents: contents, Mode: 0o644, Exclusive: !force}
	}
	if force && !hasSourceManifest {
		specs[filepath.Join(slug, "style.png.imagegen.json")] = fileutil.FileSpec{Delete: true}
	}
	if err := fileutil.WriteFileSetAtomic(root, specs, true); err != nil {
		return nil, err
	}
	return entry, nil
}

func rebuildHouseStyleCatalog(root string) error {
	data, err := houseStyleCatalogData(root, nil)
	if err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(filepath.Join(root, "catalog.json"), data, 0o644)
}

func houseStyleCatalogData(root string, replacement *HouseStyle) ([]byte, error) {
	dirs, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		dirs = nil
	} else if err != nil {
		return nil, err
	}
	styles := make([]HouseStyle, 0, len(dirs))
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if replacement != nil && dir.Name() == replacement.Slug {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, dir.Name(), "house-style.json"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var style HouseStyle
		if err := json.Unmarshal(data, &style); err != nil {
			return nil, err
		}
		styles = append(styles, style)
	}
	if replacement != nil {
		styles = append(styles, *replacement)
	}
	sort.Slice(styles, func(i, j int) bool { return styles[i].Slug < styles[j].Slug })
	return marshalJSON(HouseStyleCatalog{SchemaVersion: 1, Styles: styles})
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
	data, err := marshalJSON(value)
	if err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(path, data, 0o644)
}

func marshalJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	return data, nil
}
