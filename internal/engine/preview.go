package engine

import (
	"context"
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

	govaluate "gopkg.in/Knetic/govaluate.v3"
)

var invalidChars = regexp.MustCompile(`[<>:"/\\|?*]`)

func Preview(cfg Config) ([]*Item, error) {
	return PreviewContext(context.Background(), cfg)
}

func PreviewContext(ctx context.Context, cfg Config) ([]*Item, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(cfg.Sources) == 0 && strings.TrimSpace(cfg.Root) == "" {
		return nil, errors.New("at least one file or folder is required")
	}
	if cfg.BatchTime.IsZero() {
		cfg.BatchTime = time.Now()
	}

	files, err := scanFilesContext(ctx, cfg)
	if err != nil {
		return nil, err
	}

	infos := prepareSortInfo(files, cfg)

	rng := rand.New(rand.NewSource(cfg.BatchTime.UnixNano()))
	items := make([]*Item, 0, len(infos))
	perDir := map[string]int{}
	targets := map[string]*Item{}

	for i, info := range infos {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		path := info.Path
		dir := info.Dir
		perDir[dir]++
		oldName := info.Name

		newName := oldName
		var buildErr error
		if info.StatErr == nil {
			newName, buildErr = buildName(cfg, path, i+1, perDir[dir], len(infos), rng)
		}

		item := &Item{
			SourcePath:  path,
			Folder:      filepath.Base(dir),
			OldName:     oldName,
			NewName:     newName,
			DirIndex:    perDir[dir],
			GlobalIndex: i + 1,
			Checked:     true,
			Status:      StatusOK,
			Size:        info.Size,
			Created:     info.Created,
			Modified:    info.Modified,
		}

		if info.StatErr != nil {
			item.Status = StatusInvalid
			if os.IsNotExist(info.StatErr) {
				item.Error = "file not found"
			} else {
				item.Error = info.StatErr.Error()
			}
			item.Checked = false
			items = append(items, item)
			continue
		}

		if buildErr != nil {
			item.Status = StatusInvalid
			item.Error = buildErr.Error()
			item.Checked = false
			items = append(items, item)
			continue
		}

		if !cfg.SkipImageDimensions {
			if width, height, ok := readImageDimensions(path); ok {
				item.Width = width
				item.Height = height
			}
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

		policy := cfg.CollisionPolicy
		if policy == "" {
			policy = CollisionSkip
		}

		target := filepath.Join(dir, item.NewName)
		if item.Status == StatusOK && policy == CollisionAutoNumber && targetConflicts(target, path, targets) {
			resolved, resolvedTarget := nextAvailableName(dir, item.NewName, path, targets)
			item.NewName = resolved
			target = resolvedTarget
		}

		if len([]rune(target)) > 32767 {
			item.Status = StatusInvalid
			item.Error = "destination path is too long"
			item.Checked = false
		}

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
				// Overwrite stays blocked until it can preserve and restore the replaced file.
				item.Status = StatusConflict
				if policy == CollisionOverwrite {
					item.Error = "safe overwrite is not available"
				} else {
					item.Error = "target already exists"
				}
				item.Checked = false
			}
		}
		items = append(items, item)
	}
	return items, nil
}


func targetConflicts(target, source string, targets map[string]*Item) bool {
	if _, ok := targets[strings.ToLower(target)]; ok {
		return true
	}
	if _, err := os.Stat(target); err == nil && !strings.EqualFold(target, source) {
		return true
	}
	return false
}

func nextAvailableName(dir, name, source string, targets map[string]*Item) (string, string) {
	base, ext := BaseAndExt(name)
	for n := 1; n < 1000000; n++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, n, ext)
		target := filepath.Join(dir, candidate)
		if !targetConflicts(target, source, targets) {
			return candidate, target
		}
	}
	return name, filepath.Join(dir, name)
}

type previewSortInfo struct {
	Path     string
	Dir      string
	Name     string
	Ext      string
	Size     int64
	Created  time.Time
	Modified time.Time
	Added    int
	Manual   int
	StatErr  error
}

func prepareSortInfo(files []string, cfg Config) []previewSortInfo {
	manualRank := make(map[string]int, len(cfg.ManualOrder))
	for index, path := range cfg.ManualOrder {
		manualRank[strings.ToLower(filepath.Clean(path))] = index
	}

	infos := make([]previewSortInfo, 0, len(files))
	for index, path := range files {
		manual := len(manualRank) + index
		if rank, ok := manualRank[strings.ToLower(filepath.Clean(path))]; ok {
			manual = rank
		}
		info := previewSortInfo{
			Path:   path,
			Dir:    filepath.Dir(path),
			Name:   filepath.Base(path),
			Ext:    strings.ToLower(filepath.Ext(path)),
			Added:  index,
			Manual: manual,
		}
		st, err := os.Stat(path)
		if err != nil {
			info.StatErr = err
		} else {
			info.Size = st.Size()
			info.Created = createdTime(path, st)
			info.Modified = st.ModTime()
		}
		infos = append(infos, info)
	}

	mode := cfg.SortBy
	legacyPerFolder := mode == ""
	if mode == "" {
		mode = SortName
	}
	perFolder := cfg.SortPerFolder || legacyPerFolder

	sort.SliceStable(infos, func(i, j int) bool {
		a, b := infos[i], infos[j]
		if perFolder && !strings.EqualFold(a.Dir, b.Dir) {
			return NaturalLess(a.Dir, b.Dir)
		}
		if a.StatErr != nil || b.StatErr != nil {
			if a.StatErr != nil && b.StatErr == nil {
				return false
			}
			if a.StatErr == nil && b.StatErr != nil {
				return true
			}
		}
		cmp := comparePreviewSort(a, b, mode)
		if cmp == 0 {
			cmp = naturalCompare(a.Path, b.Path)
		}
		if cfg.SortDescending {
			return cmp > 0
		}
		return cmp < 0
	})
	return infos
}

func comparePreviewSort(a, b previewSortInfo, mode SortMode) int {
	switch mode {
	case SortCreated:
		return timeCompare(a.Created, b.Created)
	case SortModified:
		return timeCompare(a.Modified, b.Modified)
	case SortSize:
		if a.Size < b.Size {
			return -1
		}
		if a.Size > b.Size {
			return 1
		}
		return naturalCompare(a.Name, b.Name)
	case SortExtension:
		if cmp := naturalCompare(a.Ext, b.Ext); cmp != 0 {
			return cmp
		}
		return naturalCompare(a.Name, b.Name)
	case SortPath:
		return naturalCompare(a.Path, b.Path)
	case SortManual:
		if a.Manual < b.Manual {
			return -1
		}
		if a.Manual > b.Manual {
			return 1
		}
		return 0
	case SortAdded:
		if a.Added < b.Added {
			return -1
		}
		if a.Added > b.Added {
			return 1
		}
		return 0
	case SortName:
		fallthrough
	default:
		return naturalCompare(a.Name, b.Name)
	}
}

func naturalCompare(a, b string) int {
	if strings.EqualFold(a, b) {
		return 0
	}
	if NaturalLess(a, b) {
		return -1
	}
	return 1
}

func timeCompare(a, b time.Time) int {
	if a.Equal(b) {
		return 0
	}
	if a.Before(b) {
		return -1
	}
	return 1
}

// ImageDimensions reads only the selected image header and is intentionally
// not called during bulk preview generation.
func ImageDimensions(path string) (int, int, bool) {
	return readImageDimensions(path)
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
	return scanFilesContext(context.Background(), cfg)
}

func scanFilesContext(ctx context.Context, cfg Config) ([]string, error) {
	sources := append([]string(nil), cfg.Sources...)
	if len(sources) == 0 && strings.TrimSpace(cfg.Root) != "" {
		sources = []string{cfg.Root}
	}

	seen := make(map[string]struct{})
	excluded := make(map[string]struct{}, len(cfg.ExcludedPaths))
	for _, path := range cfg.ExcludedPaths {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		excluded[strings.ToLower(filepath.Clean(path))] = struct{}{}
	}
	out := make([]string, 0)
	add := func(path string) {
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
		path = filepath.Clean(path)
		key := strings.ToLower(path)
		if _, ok := excluded[key]; ok {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, path)
	}

	for _, source := range sources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}
		st, err := os.Stat(source)
		if err != nil {
			if os.IsNotExist(err) && MatchesCategory(cfg.Category, cfg.CustomExtensions, filepath.Ext(source)) {
				add(source)
				continue
			}
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
				if err := ctx.Err(); err != nil {
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
			if err := ctx.Err(); err != nil {
				return nil, err
			}
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

func buildName(cfg Config, path string, globalIndex, dirIndex, totalItems int, rng *rand.Rand) (string, error) {
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
	modified := cfg.BatchTime
	if st, err := os.Stat(path); err == nil {
		modified = st.ModTime()
	}

	for _, method := range methods {
		if method.Disabled {
			continue
		}
		var err error
		current, err = applyMethod(method, current, path, parent, globalIndex, dirIndex, totalItems, cfg.BatchTime, modified, rng)
		if err != nil {
			return "", err
		}
	}
	return current, nil
}

func applyMethod(method RenameMethod, current, path, parent string, globalIndex, dirIndex, totalItems int, batchTime, modified time.Time, rng *rand.Rand) (string, error) {
	base, ext := BaseAndExt(current)

	switch method.Type {
	case MethodTemplate:
		tpl := method.Template
		if tpl == "" {
			tpl = "<Name>"
		}
		return RenderTemplate(tpl, TemplateContext{
			Path: path, OriginalName: filepath.Base(path),
			BaseName: base, Extension: ext, ParentDir: parent,
			DirIndex: dirIndex, GlobalIndex: globalIndex, TotalItems: totalItems,
			BatchTime: batchTime, ModifiedTime: modified, Rand: rng,
		})
	case MethodList:
		lines := splitMethodLines(method.ListText)
		idx := globalIndex - 1
		if idx < 0 || idx >= len(lines) {
			return "", fmt.Errorf("list has no name for item %d", globalIndex)
		}
		name := lines[idx]
		if method.ListIncludeExtension {
			return name, nil
		}
		return name + ext, nil
	case MethodListReplace:
		renamed, err := applyListReplace(base, method)
		if err != nil {
			return "", err
		}
		return renamed + ext, nil
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
	case MethodRemove:
		return removeRuneRange(base, method.RemoveStart, method.RemoveCount) + ext, nil
	case MethodRemovePattern:
		if method.RemovePattern == "" {
			return current, nil
		}
		if method.RemovePatternRegex {
			re, err := regexp.Compile(method.RemovePattern)
			if err != nil {
				return "", fmt.Errorf("invalid remove regex: %w", err)
			}
			return re.ReplaceAllString(base, "") + ext, nil
		}
		return strings.ReplaceAll(base, method.RemovePattern, "") + ext, nil
	case MethodRenumber:
		start := method.RenumberStart
		step := method.RenumberStep
		padding := method.RenumberPadding
		if step == 0 {
			step = 1
		}
		if padding < 1 {
			padding = 1
		}
		index := globalIndex
		if method.RenumberPerDir {
			index = dirIndex
		}
		value := start + (index-1)*step
		number := fmt.Sprintf("%0*d", padding, value)
		separator := method.RenumberSeparator
		if method.RenumberPosition == PositionSuffix {
			return base + separator + number + ext, nil
		}
		return number + separator + base + ext, nil
	case MethodTrim:
		trimmed := strings.TrimSpace(base)
		if method.TrimNormalizeSpaces {
			trimmed = strings.Join(strings.Fields(trimmed), " ")
		}
		return trimmed + ext, nil
	case MethodTimestamp:
		stampTime := modified
		if method.TimestampSource == TimestampBatch {
			stampTime = batchTime
		}
		format := strings.TrimSpace(method.TimestampFormat)
		if format == "" {
			format = "yyyyMMdd-HHmmss"
		}
		stamp := stampTime.Format(toGoTimeLayout(format))
		separator := method.TimestampSeparator
		if method.TimestampPosition == PositionPrefix {
			return stamp + separator + base + ext, nil
		}
		return base + separator + stamp + ext, nil
	case MethodMove:
		moved, err := moveRuneRange(base, method.MoveStart, method.MoveCount, method.MoveTo)
		if err != nil {
			return "", err
		}
		return moved + ext, nil
	case MethodSwap:
		swapped, err := swapBySeparator(base, method.SwapSeparator, method.SwapOccurrence)
		if err != nil {
			return "", err
		}
		return swapped + ext, nil
	case MethodScript:
		return evaluateScriptExpression(method.ScriptExpression, current, parent, globalIndex, dirIndex, batchTime, modified)
	default:
		return current, nil
	}
}

func removeRuneRange(s string, start, count int) string {
	if count <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	if start < 1 {
		start = 1
	}
	from := start - 1
	if from >= len(runes) {
		return s
	}
	to := from + count
	if to > len(runes) {
		to = len(runes)
	}
	return string(append(append([]rune(nil), runes[:from]...), runes[to:]...))
}

func moveRuneRange(s string, start, count, destination int) (string, error) {
	if count <= 0 {
		return s, nil
	}
	runes := []rune(s)
	if start < 1 {
		start = 1
	}
	from := start - 1
	if from >= len(runes) {
		return s, nil
	}
	to := from + count
	if to > len(runes) {
		to = len(runes)
	}

	part := append([]rune(nil), runes[from:to]...)
	rest := append(append([]rune(nil), runes[:from]...), runes[to:]...)

	if destination < 1 {
		destination = 1
	}
	insertAt := destination - 1
	if insertAt > len(rest) {
		insertAt = len(rest)
	}

	result := make([]rune, 0, len(runes))
	result = append(result, rest[:insertAt]...)
	result = append(result, part...)
	result = append(result, rest[insertAt:]...)
	return string(result), nil
}

func splitMethodLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func applyListReplace(base string, method RenameMethod) (string, error) {
	result := base
	lines := splitMethodLines(method.ListReplaceText)
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var find, replace string
		if parts := strings.SplitN(raw, "\t", 2); len(parts) == 2 {
			find, replace = parts[0], parts[1]
		} else if parts := strings.SplitN(raw, "=>", 2); len(parts) == 2 {
			find, replace = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		} else {
			return "", fmt.Errorf("invalid list replace rule on line %d; use find => replace", i+1)
		}
		if find == "" {
			continue
		}

		if method.ListReplaceRegex {
			pattern := find
			if !method.ListReplaceCaseSensitive {
				pattern = "(?i)" + pattern
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return "", fmt.Errorf("invalid list replace regex on line %d: %w", i+1, err)
			}
			result = re.ReplaceAllString(result, replace)
			continue
		}

		if method.ListReplaceCaseSensitive {
			result = strings.ReplaceAll(result, find, replace)
		} else {
			re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(find))
			result = re.ReplaceAllStringFunc(result, func(string) string { return replace })
		}
	}
	return result, nil
}

func swapBySeparator(base, separator string, occurrence int) (string, error) {
	if separator == "" {
		return "", errors.New("swap separator is required")
	}
	if occurrence < 1 {
		occurrence = 1
	}

	start := 0
	index := -1
	for i := 0; i < occurrence; i++ {
		found := strings.Index(base[start:], separator)
		if found < 0 {
			return base, nil
		}
		index = start + found
		start = index + len(separator)
	}

	left := base[:index]
	right := base[index+len(separator):]
	return right + separator + left, nil
}

func evaluateScriptExpression(expr, current, parent string, globalIndex, dirIndex int, batchTime, modified time.Time) (string, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return current, nil
	}

	base, ext := BaseAndExt(current)
	functions := map[string]govaluate.ExpressionFunction{
		"lower": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("lower() expects 1 argument")
			}
			return strings.ToLower(fmt.Sprint(args[0])), nil
		},
		"upper": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("upper() expects 1 argument")
			}
			return strings.ToUpper(fmt.Sprint(args[0])), nil
		},
		"trim": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("trim() expects 1 argument")
			}
			return strings.TrimSpace(fmt.Sprint(args[0])), nil
		},
		"replace": func(args ...interface{}) (interface{}, error) {
			if len(args) != 3 {
				return nil, errors.New("replace() expects 3 arguments")
			}
			return strings.ReplaceAll(fmt.Sprint(args[0]), fmt.Sprint(args[1]), fmt.Sprint(args[2])), nil
		},
		"concat": func(args ...interface{}) (interface{}, error) {
			var b strings.Builder
			for _, arg := range args {
				b.WriteString(fmt.Sprint(arg))
			}
			return b.String(), nil
		},
		"substr": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 || len(args) > 3 {
				return nil, errors.New("substr() expects 2 or 3 arguments")
			}
			s := []rune(fmt.Sprint(args[0]))
			start, ok := numericArg(args[1])
			if !ok {
				return nil, errors.New("substr() start must be numeric")
			}
			if start < 0 {
				start = 0
			}
			if start > len(s) {
				start = len(s)
			}
			end := len(s)
			if len(args) == 3 {
				count, ok := numericArg(args[2])
				if !ok {
					return nil, errors.New("substr() count must be numeric")
				}
				if count < 0 {
					count = 0
				}
				end = start + count
				if end > len(s) {
					end = len(s)
				}
			}
			return string(s[start:end]), nil
		},
	}

	evaluable, err := govaluate.NewEvaluableExpressionWithFunctions(expr, functions)
	if err != nil {
		return "", fmt.Errorf("invalid script expression: %w", err)
	}

	params := map[string]interface{}{
		"Name":          base,
		"Ext":           ext,
		"FullName":      current,
		"Index":         globalIndex,
		"DirIndex":      dirIndex,
		"DirName":       parent,
		"UnixTimestamp": batchTime.Unix(),
		"ModifiedUnix":  modified.Unix(),
		"BatchDate":     batchTime.Format("2006-01-02 15:04:05"),
		"ModifiedDate":  modified.Format("2006-01-02 15:04:05"),
	}

	value, err := evaluable.Evaluate(params)
	if err != nil {
		return "", fmt.Errorf("script evaluation failed: %w", err)
	}
	if value == nil {
		return "", errors.New("script returned no value")
	}

	result := fmt.Sprint(value)
	if result == "" {
		return "", errors.New("script returned an empty filename")
	}
	if filepath.Ext(result) == "" && ext != "" {
		result += ext
	}
	return result, nil
}

func numericArg(value interface{}) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	case int32:
		return int(v), true
	default:
		return 0, false
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
	if strings.TrimSpace(name) == "" {
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
	base := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
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
