package engine

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
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
	if len(cfg.Sources) == 0 && strings.TrimSpace(cfg.Root) == "" {
		return nil, errors.New("at least one file or folder is required")
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

		if st, statErr := os.Stat(path); statErr == nil {
			item.Size = st.Size()
			item.Modified = st.ModTime()
		}
		if width, height, ok := readImageDimensions(path); ok {
			item.Width = width
			item.Height = height
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

func readImageDimensions(path string) (int, int, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".jpe", ".jfif", ".png", ".gif":
	default:
		return 0, 0, false
	}

	f, err := os.Open(path)
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

func scanFiles(cfg Config) ([]string, error) {
	sources := append([]string(nil), cfg.Sources...)
	if len(sources) == 0 && strings.TrimSpace(cfg.Root) != "" {
		sources = []string{cfg.Root}
	}

	seen := make(map[string]struct{})
	out := make([]string, 0)
	add := func(path string) {
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
		path = filepath.Clean(path)
		key := strings.ToLower(path)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, path)
	}

	for _, source := range sources {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}
		st, err := os.Stat(source)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", source, err)
		}

		if !st.IsDir() {
			if MatchesCategory(cfg.Category, cfg.CustomExtensions, filepath.Ext(source)) {
				add(source)
			}
			continue
		}

		if cfg.Recursive {
			err := filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if MatchesCategory(cfg.Category, cfg.CustomExtensions, filepath.Ext(d.Name())) {
					add(path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		entries, err := os.ReadDir(source)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if MatchesCategory(cfg.Category, cfg.CustomExtensions, filepath.Ext(e.Name())) {
				add(filepath.Join(source, e.Name()))
			}
		}
	}

	return out, nil
}

func buildName(cfg Config, path string, globalIndex, dirIndex int, rng *rand.Rand) (string, error) {
	methods := cfg.Methods
	if len(methods) == 0 {
		methodType := cfg.Method
		if methodType == "" {
			methodType = MethodTemplate
		}
		methods = []RenameMethod{{
			Type:        methodType,
			Template:    cfg.Template,
			Find:        cfg.Find,
			ReplaceWith: cfg.ReplaceWith,
			UseRegex:    cfg.UseRegex,
			Prefix:      cfg.Prefix,
			Suffix:      cfg.Suffix,
			CaseMode:    cfg.CaseMode,
		}}
	}

	current := filepath.Base(path)
	parent := filepath.Base(filepath.Dir(path))
	for _, method := range methods {
		if method.Disabled {
			continue
		}
		var err error
		current, err = applyMethod(method, current, parent, globalIndex, dirIndex, cfg.BatchTime, rng)
		if err != nil {
			return "", err
		}
	}
	return current, nil
}

func applyMethod(method RenameMethod, current, parent string, globalIndex, dirIndex int, batchTime time.Time, rng *rand.Rand) (string, error) {
	base, ext := BaseAndExt(current)

	switch method.Type {
	case MethodTemplate:
		tpl := method.Template
		if tpl == "" {
			tpl = "<Name>"
		}
		return RenderTemplate(tpl, TemplateContext{
			BaseName: base, Extension: ext, ParentDir: parent,
			DirIndex: dirIndex, GlobalIndex: globalIndex,
			BatchTime: batchTime, Rand: rng,
		})
	case MethodReplace:
		if method.Find == "" {
			return current, nil
		}
		var renamed string
		if method.UseRegex {
			re, err := regexp.Compile(method.Find)
			if err != nil {
				return "", fmt.Errorf("invalid regex: %w", err)
			}
			renamed = re.ReplaceAllString(base, method.Find)
			renamed = re.ReplaceAllString(base, method.ReplaceWith)
		} else {
			renamed = strings.ReplaceAll(base, method.Find, method.ReplaceWith)
		}
		return renamed + ext, nil
	case MethodPrefixSuffix:
		return method.Prefix + base + method.Suffix + ext, nil
	case MethodCase:
		switch method.CaseMode {
		case CaseUpper:
			return strings.ToUpper(base) + ext, nil
		case CaseTitle:
			return titleCase(base) + ext, nil
		case CaseLower, "":
			return strings.ToLower(base) + ext, nil
		default:
			return current, nil
		}
	default:
		return current, nil
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
