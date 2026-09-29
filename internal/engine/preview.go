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

	govaluate "gopkg.in/Knetic/govaluate.v3"
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
	modified := cfg.BatchTime
	if st, err := os.Stat(path); err == nil {
		modified = st.ModTime()
	}

	for _, method := range methods {
		if method.Disabled {
			continue
		}
		var err error
		current, err = applyMethod(method, current, parent, globalIndex, dirIndex, cfg.BatchTime, modified, rng)
		if err != nil {
			return "", err
		}
	}
	return current, nil
}

func applyMethod(method RenameMethod, current, parent string, globalIndex, dirIndex int, batchTime, modified time.Time, rng *rand.Rand) (string, error) {
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
