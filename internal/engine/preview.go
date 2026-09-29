package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

var invalidChars = regexp.MustCompile(`[<>:"/\\|?*]`)

func Preview(cfg Config) ([]*Item, error) {
	if cfg.Root == "" {
		return nil, errors.New("folder is required")
	}
	st, err := os.Stat(cfg.Root)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, errors.New("selected path is not a folder")
	}
	if cfg.BatchTime.IsZero() {
		cfg.BatchTime = time.Now()
	}

	files, err := scanFiles(cfg)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(files, func(i, j int) bool {
		di, dj := filepath.Dir(files[i]), filepath.Dir(files[j])
		if !strings.EqualFold(di, dj) {
			return NaturalLess(di, dj)
		}
		return NaturalLess(filepath.Base(files[i]), filepath.Base(files[j]))
	})

	rng := rand.New(rand.NewSource(cfg.BatchTime.UnixNano()))
	items := make([]*Item, 0, len(files))
	perDir := map[string]int{}
	targets := map[string]*Item{}

	for i, path := range files {
		dir := filepath.Dir(path)
		perDir[dir]++
		oldName := filepath.Base(path)
		newName, err := buildName(cfg, path, i+1, perDir[dir], rng)
		item := &Item{
			SourcePath:  path,
			Folder:      filepath.Base(dir),
			OldName:     oldName,
			NewName:     newName,
			DirIndex:    perDir[dir],
			GlobalIndex: i + 1,
			Checked:     true,
			Status:      StatusOK,
		}
		if err != nil {
			item.Status = StatusInvalid
			item.Error = err.Error()
			item.Checked = false
			items = append(items, item)
			continue
		}

		if strings.EqualFold(oldName, newName) {
			item.Status = StatusUnchanged
			item.Checked = false
		}
		if err := validateWindowsName(newName); err != nil {
			item.Status = StatusInvalid
			item.Error = err.Error()
			item.Checked = false
		}

		target := filepath.Join(dir, newName)
		key := strings.ToLower(target)
		if prev, ok := targets[key]; ok {
			item.Status = StatusConflict
			item.Error = "duplicate target name"
			item.Checked = false
			prev.Status = StatusConflict
			prev.Error = "duplicate target name"
			prev.Checked = false
		} else {
			targets[key] = item
		}
		if item.Status == StatusOK {
			if _, err := os.Stat(target); err == nil && !strings.EqualFold(target, path) {
				item.Status = StatusConflict
				item.Error = "target already exists"
				item.Checked = false
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func scanFiles(cfg Config) ([]string, error) {
	var out []string
	if cfg.Recursive {
		err := filepath.WalkDir(cfg.Root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if MatchesCategory(cfg.Category, cfg.CustomExtensions, filepath.Ext(d.Name())) {
				out = append(out, path)
			}
			return nil
		})
		return out, err
	}
	entries, err := os.ReadDir(cfg.Root)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if MatchesCategory(cfg.Category, cfg.CustomExtensions, filepath.Ext(e.Name())) {
			out = append(out, filepath.Join(cfg.Root, e.Name()))
		}
	}
	return out, nil
}

func buildName(cfg Config, path string, globalIndex, dirIndex int, rng *rand.Rand) (string, error) {
	old := filepath.Base(path)
	base, ext := BaseAndExt(old)
	parent := filepath.Base(filepath.Dir(path))

	switch cfg.Method {
	case MethodTemplate:
		return RenderTemplate(cfg.Template, TemplateContext{
			BaseName: base, Extension: ext, ParentDir: parent,
			DirIndex: dirIndex, GlobalIndex: globalIndex,
			BatchTime: cfg.BatchTime, Rand: rng,
		})
	case MethodReplace:
		if cfg.Find == "" {
			return old, nil
		}
		var renamed string
		if cfg.UseRegex {
			re, err := regexp.Compile(cfg.Find)
			if err != nil {
				return "", fmt.Errorf("invalid regex: %w", err)
			}
			renamed = re.ReplaceAllString(base, cfg.ReplaceWith)
		} else {
			renamed = strings.ReplaceAll(base, cfg.Find, cfg.ReplaceWith)
		}
		return renamed + ext, nil
	case MethodPrefixSuffix:
		return cfg.Prefix + base + cfg.Suffix + ext, nil
	case MethodCase:
		switch cfg.CaseMode {
		case CaseLower:
			return strings.ToLower(base) + ext, nil
		case CaseUpper:
			return strings.ToUpper(base) + ext, nil
		case CaseTitle:
			return titleCase(base) + ext, nil
		default:
			return old, nil
		}
	default:
		return old, nil
	}
}

func titleCase(s string) string {
	var b strings.Builder
	upperNext := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if upperNext {
				b.WriteRune(unicode.ToUpper(r))
				upperNext = false
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
		} else {
			upperNext = true
			b.WriteRune(r)
		}
	}
	return b.String()
}

func validateWindowsName(name string) error {
	if name == "" {
		return errors.New("empty file name")
	}
	if invalidChars.MatchString(name) {
		return errors.New("name contains Windows-forbidden characters")
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return errors.New("name cannot end with a dot or space")
	}
	for _, r := range name {
		if r < 32 {
			return errors.New("name contains control characters")
		}
	}
	base := strings.ToUpper(strings.TrimSuffix(name, filepath.Ext(name)))
	reserved := map[string]struct{}{
		"CON": {}, "PRN": {}, "AUX": {}, "NUL": {},
		"COM1": {}, "COM2": {}, "COM3": {}, "COM4": {}, "COM5": {}, "COM6": {}, "COM7": {}, "COM8": {}, "COM9": {},
		"LPT1": {}, "LPT2": {}, "LPT3": {}, "LPT4": {}, "LPT5": {}, "LPT6": {}, "LPT7": {}, "LPT8": {}, "LPT9": {},
	}
	if _, ok := reserved[base]; ok {
		return errors.New("name is reserved by Windows")
	}
	return nil
}
