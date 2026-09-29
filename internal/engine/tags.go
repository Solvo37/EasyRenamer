package engine

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var modifierNames = map[string]struct{}{
	"default": {}, "alt": {}, "append": {}, "prepend": {}, "suffix": {}, "prefix": {},
	"pad": {}, "trim": {}, "upper": {}, "lower": {}, "titlecase": {}, "substr": {},
	"rsubstr": {}, "insert": {}, "remove": {}, "replace": {}, "word": {}, "rword": {},
	"add": {}, "subtract": {}, "multiply": {}, "divide": {},
}

func renderToken(body string, ctx *TemplateContext) (string, error) {
	candidates, defaultText, sharedMods := parseFallback(body)

	for _, candidate := range candidates {
		value, err := resolveTagWithModifiers(candidate, ctx)
		if err != nil {
			return "", err
		}
		if value != "" {
			if len(sharedMods) > 0 {
				value, err = applyModifiers(value, sharedMods, ctx)
				if err != nil {
					return "", err
				}
			}
			return value, nil
		}
	}

	if defaultText != "" {
		value := defaultText
		var err error
		if len(sharedMods) > 0 {
			value, err = applyModifiers(value, sharedMods, ctx)
			if err != nil {
				return "", err
			}
		}
		return value, nil
	}
	return "", nil
}

func parseFallback(body string) (candidates []string, defaultText string, sharedMods []string) {
	body = strings.TrimSpace(body)
	defaultIndex := -1
	for i := 0; i < len(body); i++ {
		if body[i] != '|' {
			continue
		}
		if i+1 < len(body) && body[i+1] == '|' {
			i++
			continue
		}
		defaultIndex = i
		break
	}

	candidatePart := body
	if defaultIndex >= 0 {
		candidatePart = body[:defaultIndex]
		defaultText = body[defaultIndex+1:]
		defaultText, sharedMods = splitLiteralModifiers(defaultText)
	}
	for _, part := range strings.Split(candidatePart, "||") {
		if strings.TrimSpace(part) != "" {
			candidates = append(candidates, strings.TrimSpace(part))
		}
	}
	if len(candidates) == 0 {
		candidates = []string{body}
	}
	return
}

func splitLiteralModifiers(s string) (string, []string) {
	parts := strings.Split(s, ":")
	for i := 1; i < len(parts); i++ {
		if isModifier(parts[i]) {
			return strings.Join(parts[:i], ":"), parts[i:]
		}
	}
	return s, nil
}

func resolveTagWithModifiers(expr string, ctx *TemplateContext) (string, error) {
	core, mods := splitCoreAndModifiers(expr)
	value, err := resolveTagCore(core, ctx)
	if err != nil {
		return "", err
	}
	if len(mods) > 0 {
		return applyModifiers(value, mods, ctx)
	}
	return value, nil
}

func splitCoreAndModifiers(expr string) (string, []string) {
	parts := strings.Split(expr, ":")
	for i := 1; i < len(parts); i++ {
		if isModifier(parts[i]) {
			return strings.Join(parts[:i], ":"), parts[i:]
		}
	}
	return expr, nil
}

func isModifier(s string) bool {
	_, ok := modifierNames[strings.ToLower(strings.TrimSpace(s))]
	return ok
}

func resolveTagCore(expr string, ctx *TemplateContext) (string, error) {
	expr = strings.TrimSpace(expr)
	name, args := splitTagNameArgs(expr)
	lowerName := strings.ToLower(strings.TrimSpace(name))

	if lowerName == "date" && strings.EqualFold(filepath.Ext(ctx.Path), ".eml") {
		if value := metadataValue(ctx.Metadata, "date"); value != "" {
			return value, nil
		}
	}

	switch lowerName {
	case "name":
		return ctx.BaseName, nil
	case "ext":
		return strings.TrimPrefix(ctx.Extension, "."), nil
	case "foldername", "dirname", "parent":
		index := 1
		if len(args) > 0 {
			index, _ = strconv.Atoi(strings.TrimSpace(args[0]))
			if index == 0 {
				index = 1
			}
		}
		return folderNameAt(ctx.Path, index, ctx.ParentDir), nil
	case "inc", "inc nr":
		return formatCounterWithStep(ctx.GlobalIndex, strings.Join(args, ":"), 1, 1), nil
	case "inc nrdir":
		return formatCounterWithStep(ctx.DirIndex, strings.Join(args, ":"), 1, 1), nil
	case "dec nr":
		start, step := 1, 1
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			start, _ = strconv.Atoi(strings.TrimSpace(args[0]))
		}
		if len(args) > 1 && strings.TrimSpace(args[1]) != "" {
			step, _ = strconv.Atoi(strings.TrimSpace(args[1]))
		}
		return strconv.Itoa(start - (maxInt(ctx.GlobalIndex, 1)-1)*step), nil
	case "inc alpha":
		return alphaCounter(ctx.GlobalIndex, args), nil
	case "inc hex":
		start, step := parseStartStep(args, 1, 1)
		return strings.ToUpper(strconv.FormatInt(int64(start+(maxInt(ctx.GlobalIndex, 1)-1)*step), 16)), nil
	case "inc roman":
		start, step := parseStartStep(args, 1, 1)
		return toRoman(start + (maxInt(ctx.GlobalIndex, 1)-1)*step), nil
	case "rand":
		if len(args) >= 2 {
			minV, err1 := strconv.Atoi(strings.TrimSpace(args[0]))
			maxV, err2 := strconv.Atoi(strings.TrimSpace(args[1]))
			if err1 != nil || err2 != nil || maxV < minV {
				return "", fmt.Errorf("invalid random range in <%s>", expr)
			}
			return strconv.Itoa(minV + ctx.Rand.Intn(maxV-minV+1)), nil
		}
		return strconv.Itoa(ctx.Rand.Intn(10)), nil
	case "rand str":
		n := 8
		if len(args) > 0 {
			var err error
			n, err = strconv.Atoi(strings.TrimSpace(args[0]))
			if err != nil || n < 1 || n > 512 {
				return "", fmt.Errorf("invalid random string length in <%s>", expr)
			}
		}
		return randomChars(ctx.Rand, n, "abcdefghijklmnopqrstuvwxyz"), nil
	case "rand alpha":
		n := 8
		if len(args) > 0 {
			var err error
			n, err = strconv.Atoi(strings.TrimSpace(args[0]))
			if err != nil || n < 1 || n > 512 {
				return "", fmt.Errorf("invalid random alpha length in <%s>", expr)
			}
		}
		return randomChars(ctx.Rand, n, "abcdefghijklmnopqrstuvwxyz"), nil
	case "num items":
		return padNumber(int64(ctx.TotalItems), firstArg(args)), nil
	case "num files":
		n, _ := countDirectoryEntries(filepath.Dir(ctx.Path))
		return padNumber(int64(n), firstArg(args)), nil
	case "num dirs":
		_, n := countDirectoryEntries(filepath.Dir(ctx.Path))
		return padNumber(int64(n), firstArg(args)), nil
	case "word":
		return wordTag(ctx.BaseName, args, false), nil
	case "rword":
		return wordTag(ctx.BaseName, args, true), nil
	case "mediatype":
		if ctx.MediaType != "" {
			return ctx.MediaType, nil
		}
		return detectMediaType(ctx.Extension), nil
	case "substr":
		return substringTag(ctx.BaseName, args, false), nil
	case "rsubstr":
		return substringTag(ctx.BaseName, args, true), nil
	case "switch":
		if len(args) == 0 {
			return "", nil
		}
		return args[(maxInt(ctx.GlobalIndex, 1)-1)%len(args)], nil
	case "delimiter", "-":
		if len(args) > 0 {
			return strings.Join(args, ":"), nil
		}
		return " - ", nil
	case "subfolder":
		index := 1
		if len(args) > 0 {
			index, _ = strconv.Atoi(strings.TrimSpace(args[0]))
		}
		return subfolderAt(ctx.Path, index), nil
	case "file line":
		if len(args) == 0 {
			return "", nil
		}
		line, _ := strconv.Atoi(strings.TrimSpace(args[0]))
		return readFileLine(ctx.Path, line), nil
	case "file content":
		return readFileContentSlice(ctx.Path, args), nil
	case "metadata":
		if len(args) == 0 {
			return "", nil
		}
		return metadataValue(ctx.Metadata, strings.Join(args, ":")), nil
	}

	if isBatchDateTag(lowerName) {
		return renderDateTag(lowerName, args, ctx.BatchTime), nil
	}
	if isCreatedDateTag(lowerName) {
		return renderDateTag(strings.TrimSuffix(lowerName, " created"), args, ctx.CreatedTime), nil
	}
	if isModifiedDateTag(lowerName) {
		return renderDateTag(strings.TrimSuffix(lowerName, " modified"), args, ctx.ModifiedTime), nil
	}

	switch lowerName {
	case "filesize text":
		return humanFileSize(ctx.FileSize), nil
	case "filesize b":
		return padNumber(ctx.FileSize, firstArg(args)), nil
	case "filesize kb":
		return padNumber(ctx.FileSize/1024, firstArg(args)), nil
	case "filesize mb":
		return padNumber(ctx.FileSize/(1024*1024), firstArg(args)), nil
	case "filesize gb":
		return padNumber(ctx.FileSize/(1024*1024*1024), firstArg(args)), nil
	case "filesize tb":
		return padNumber(ctx.FileSize/(1024*1024*1024*1024), firstArg(args)), nil
	case "md5":
		return checksumFile(ctx.Path, "md5")
	case "sha1":
		return checksumFile(ctx.Path, "sha1")
	case "width":
		return padNumber(int64(ctx.Width), firstArg(args)), nil
	case "height":
		return padNumber(int64(ctx.Height), firstArg(args)), nil
	}

	if strings.HasPrefix(lowerName, "img ") {
		return resolveImageTag(lowerName, args, ctx), nil
	}
	if strings.HasPrefix(lowerName, "video ") || strings.HasPrefix(lowerName, "duration") || lowerName == "framerate" {
		return resolveVideoTag(lowerName, args, ctx), nil
	}
	if strings.HasPrefix(lowerName, "gps ") {
		return metadataValue(ctx.Metadata, lowerName), nil
	}
	if strings.HasPrefix(lowerName, "exe ") || lowerName == "description" {
		return metadataValue(ctx.Metadata, lowerName), nil
	}

	switch lowerName {
	case "album", "artist", "genre", "title", "audio year", "track", "trackcount", "disc", "disccount",
		"pages", "creator", "subject", "date", "from", "fromname", "fromemail",
		"to", "toname", "toemail", "cc", "ccname", "ccemail", "bcc", "bccname", "bccemail",
		"author", "copyright":
		key := lowerName
		if len(args) > 0 && (lowerName == "to" || lowerName == "toname" || lowerName == "toemail" ||
			lowerName == "cc" || lowerName == "ccname" || lowerName == "ccemail" ||
			lowerName == "bcc" || lowerName == "bccname" || lowerName == "bccemail") {
			key += " " + strings.TrimSpace(args[0])
		}
		value := metadataValue(ctx.Metadata, key)
		if value != "" && (lowerName == "track" || lowerName == "trackcount" || lowerName == "disc" || lowerName == "disccount" || lowerName == "pages") {
			if n, err := strconv.ParseInt(value, 10, 64); err == nil {
				value = padNumber(n, firstArg(args))
			}
		}
		return value, nil
	}

	if value := metadataValue(ctx.Metadata, lowerName); value != "" {
		return value, nil
	}
	return "", nil
}

func splitTagNameArgs(expr string) (string, []string) {
	parts := strings.Split(expr, ":")
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), nil
	}
	return strings.TrimSpace(parts[0]), parts[1:]
}

func parseStartStep(args []string, defaultStart, defaultStep int) (int, int) {
	start, step := defaultStart, defaultStep
	if len(args) > 0 {
		if n, err := strconv.Atoi(strings.TrimSpace(args[0])); err == nil {
			start = n
		}
	}
	if len(args) > 1 {
		if n, err := strconv.Atoi(strings.TrimSpace(args[1])); err == nil {
			step = n
		}
	}
	return start, step
}

func alphaCounter(index int, args []string) string {
	start := "A"
	step := 1
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		start = strings.TrimSpace(args[0])
	}
	if len(args) > 1 {
		if n, err := strconv.Atoi(strings.TrimSpace(args[1])); err == nil {
			step = n
		}
	}
	runes := []rune(start)
	if len(runes) == 0 {
		runes = []rune{'A'}
	}
	upper := unicode.IsUpper(runes[0])
	width := len(runes)
	first := unicode.ToUpper(runes[0])
	offset := int(first-'A') + (maxInt(index, 1)-1)*step
	if offset < 0 {
		offset = 0
	}
	var chars []rune
	n := offset
	for {
		chars = append([]rune{rune('A' + n%26)}, chars...)
		n = n/26 - 1
		if n < 0 {
			break
		}
	}
	for len(chars) < width {
		chars = append([]rune{'A'}, chars...)
	}
	value := string(chars)
	if !upper {
		value = strings.ToLower(value)
	}
	return value
}

func toRoman(n int) string {
	if n <= 0 {
		return strconv.Itoa(n)
	}
	values := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var b strings.Builder
	for i, v := range values {
		for n >= v {
			b.WriteString(symbols[i])
			n -= v
		}
	}
	return b.String()
}

func folderNameAt(path string, index int, fallback string) string {
	if path == "" {
		return fallback
	}
	dir := filepath.Clean(filepath.Dir(path))
	var parts []string
	for {
		base := filepath.Base(dir)
		if base == "." || base == string(filepath.Separator) || base == "" {
			break
		}
		parts = append(parts, base)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if len(parts) == 0 {
		return fallback
	}
	if index > 0 {
		if index <= len(parts) {
			return parts[index-1]
		}
		return ""
	}
	leftIndex := -index
	if leftIndex >= 1 && leftIndex <= len(parts) {
		return parts[len(parts)-leftIndex]
	}
	return ""
}

func countDirectoryEntries(dir string) (files, dirs int) {
	if dir == "" {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			dirs++
		} else {
			files++
		}
	}
	return
}

func subfolderAt(path string, index int) string {
	if index < 1 {
		index = 1
	}
	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var folders []string
	for _, e := range entries {
		if e.IsDir() {
			folders = append(folders, e.Name())
		}
	}
	if index > len(folders) {
		return ""
	}
	return folders[index-1]
}

func wordTag(s string, args []string, reverse bool) string {
	index, count := 1, 1
	separator := " _-.()[]{}"
	if len(args) > 0 {
		if n, err := strconv.Atoi(strings.TrimSpace(args[0])); err == nil && n > 0 {
			index = n
		}
	}
	if len(args) > 1 {
		if n, err := strconv.Atoi(strings.TrimSpace(args[1])); err == nil && n > 0 {
			count = n
		}
	}
	if len(args) > 2 && args[2] != "" {
		separator = args[2]
	}
	words := strings.FieldsFunc(s, func(r rune) bool { return strings.ContainsRune(separator, r) })
	if reverse {
		start := len(words) - index - count + 1
		if start < 0 {
			start = 0
		}
		end := len(words) - index + 1
		if end < 0 {
			return ""
		}
		if end > len(words) {
			end = len(words)
		}
		if start >= end {
			return ""
		}
		return strings.Join(words[start:end], " ")
	}
	start := index - 1
	if start >= len(words) {
		return ""
	}
	end := start + count
	if end > len(words) {
		end = len(words)
	}
	return strings.Join(words[start:end], " ")
}

func substringTag(s string, args []string, reverse bool) string {
	if len(args) == 0 {
		return ""
	}
	runes := []rune(s)
	pos := 1
	if n, err := strconv.Atoi(strings.Trim(strings.TrimSpace(args[0]), "\"")); err == nil {
		pos = n
	} else {
		needle := strings.Trim(strings.TrimSpace(args[0]), "\"")
		if reverse {
			idx := strings.LastIndex(s, needle)
			if idx < 0 {
				return ""
			}
			pos = len([]rune(s[:idx])) + 1
		} else {
			idx := strings.Index(s, needle)
			if idx < 0 {
				return ""
			}
			pos = len([]rune(s[:idx])) + 1
		}
	}
	count := len(runes)
	if len(args) > 1 {
		if n, err := strconv.Atoi(strings.TrimSpace(args[1])); err == nil {
			count = n
		} else {
			endNeedle := strings.Trim(strings.TrimSpace(args[1]), "\"")
			startIdx := maxInt(pos-1, 0)
			remaining := string(runes[startIdx:])
			idx := strings.Index(remaining, endNeedle)
			if idx < 0 {
				return ""
			}
			count = len([]rune(remaining[:idx]))
		}
	}
	if pos < 1 {
		pos = 1
	}
	start := pos - 1
	if reverse {
		start = len(runes) - pos - count + 1
	}
	if start < 0 {
		start = 0
	}
	if start >= len(runes) || count <= 0 {
		return ""
	}
	end := start + count
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

func isBatchDateTag(name string) bool {
	switch name {
	case "date", "time", "sec", "min", "hour", "day", "month", "year", "unixtimestamp", "unix":
		return true
	}
	return false
}

func isCreatedDateTag(name string) bool {
	return strings.HasSuffix(name, " created") && isBatchDateTag(strings.TrimSuffix(name, " created"))
}

func isModifiedDateTag(name string) bool {
	return strings.HasSuffix(name, " modified") && isBatchDateTag(strings.TrimSuffix(name, " modified"))
}

func renderDateTag(name string, args []string, t time.Time) string {
	if t.IsZero() {
		return ""
	}
	switch name {
	case "date":
		pattern := "yyyy-mm-dd"
		if len(args) > 0 {
			pattern = strings.Join(args, ":")
		}
		return formatAdvancedDate(t, pattern)
	case "time":
		pattern := "hh:nn:ss"
		if len(args) > 0 {
			pattern = strings.Join(args, ":")
		}
		return formatAdvancedDate(t, pattern)
	case "sec":
		return padNumber(int64(t.Second()), "00")
	case "min":
		return padNumber(int64(t.Minute()), "00")
	case "hour":
		return padNumber(int64(t.Hour()), "00")
	case "day":
		return padNumber(int64(t.Day()), "00")
	case "month":
		return formatMonth(t.Month(), firstArg(args))
	case "year":
		if firstArg(args) == "00" {
			return fmt.Sprintf("%02d", t.Year()%100)
		}
		return strconv.Itoa(t.Year())
	case "unixtimestamp", "unix":
		return strconv.FormatInt(t.Unix(), 10)
	}
	return ""
}

func formatAdvancedDate(t time.Time, pattern string) string {
	var b strings.Builder
	for i := 0; i < len(pattern); {
		lower := strings.ToLower(pattern[i:])
		switch {
		case strings.HasPrefix(lower, "yyyy"):
			b.WriteString(t.Format("2006"))
			i += 4
		case strings.HasPrefix(lower, "mmmm"):
			b.WriteString(t.Format("January"))
			i += 4
		case strings.HasPrefix(lower, "mmm"):
			b.WriteString(t.Format("Jan"))
			i += 3
		case strings.HasPrefix(lower, "dddd"):
			b.WriteString(t.Format("Monday"))
			i += 4
		case strings.HasPrefix(lower, "ddd"):
			b.WriteString(t.Format("Mon"))
			i += 3
		case strings.HasPrefix(lower, "yy"):
			b.WriteString(t.Format("06"))
			i += 2
		case strings.HasPrefix(lower, "mm"):
			b.WriteString(t.Format("01"))
			i += 2
		case strings.HasPrefix(lower, "dd"):
			b.WriteString(t.Format("02"))
			i += 2
		case strings.HasPrefix(lower, "hh"):
			b.WriteString(t.Format("15"))
			i += 2
		case strings.HasPrefix(lower, "nn"):
			b.WriteString(t.Format("04"))
			i += 2
		case strings.HasPrefix(lower, "ss"):
			b.WriteString(t.Format("05"))
			i += 2
		case strings.HasPrefix(lower, "m"):
			b.WriteString(strconv.Itoa(int(t.Month())))
			i++
		case strings.HasPrefix(lower, "d"):
			b.WriteString(strconv.Itoa(t.Day()))
			i++
		case strings.HasPrefix(lower, "h"):
			b.WriteString(strconv.Itoa(t.Hour()))
			i++
		case strings.HasPrefix(lower, "n"):
			b.WriteString(strconv.Itoa(t.Minute()))
			i++
		case strings.HasPrefix(lower, "s"):
			b.WriteString(strconv.Itoa(t.Second()))
			i++
		default:
			b.WriteByte(pattern[i])
			i++
		}
	}
	return b.String()
}

func formatMonth(month time.Month, spec string) string {
	switch spec {
	case "0":
		return strconv.Itoa(int(month))
	case "xxx":
		return strings.ToLower(month.String()[:3])
	case "Xxx":
		return month.String()[:3]
	case "XXX":
		return strings.ToUpper(month.String()[:3])
	case "xxxx":
		return strings.ToLower(month.String())
	case "Xxxx":
		return month.String()
	case "XXXX":
		return strings.ToUpper(month.String())
	default:
		return fmt.Sprintf("%02d", int(month))
	}
}

func resolveImageTag(name string, args []string, ctx *TemplateContext) string {
	switch name {
	case "img year":
		return renderImageDatePart(ctx, "year", args)
	case "img month":
		return renderImageDatePart(ctx, "month", args)
	case "img day":
		return renderImageDatePart(ctx, "day", args)
	case "img hour":
		return renderImageDatePart(ctx, "hour", args)
	case "img min":
		return renderImageDatePart(ctx, "min", args)
	case "img sec":
		return renderImageDatePart(ctx, "sec", args)
	case "img subsec":
		return metadataValue(ctx.Metadata, "img subsec")
	case "img dateoriginal", "img datecreate", "img timeoriginal", "img timecreate", "img dpi":
		value := metadataValue(ctx.Metadata, name)
		if value == "" {
			return ""
		}
		if strings.Contains(name, "date") || strings.Contains(name, "time") {
			if parsed, ok := parseMetadataTime(value); ok {
				pattern := "yyyy-mm-dd hh:nn:ss"
				if len(args) > 0 {
					pattern = strings.Join(args, ":")
				}
				return formatAdvancedDate(parsed, pattern)
			}
		}
		return value
	}
	return metadataValue(ctx.Metadata, name)
}

func renderImageDatePart(ctx *TemplateContext, part string, args []string) string {
	value := metadataValue(ctx.Metadata, "img dateoriginal")
	if value == "" {
		value = metadataValue(ctx.Metadata, "img datecreate")
	}
	t, ok := parseMetadataTime(value)
	if !ok {
		return ""
	}
	return renderDateTag(part, args, t)
}

func resolveVideoTag(name string, args []string, ctx *TemplateContext) string {
	if name == "framerate" {
		return metadataValue(ctx.Metadata, "framerate")
	}
	if strings.HasPrefix(name, "duration") {
		value := metadataValue(ctx.Metadata, "duration seconds")
		seconds, _ := strconv.ParseFloat(value, 64)
		switch name {
		case "duration":
			if len(args) > 0 {
				switch strings.ToLower(strings.TrimSpace(args[0])) {
				case "hours":
					return trimFloat(seconds / 3600)
				case "mins":
					return trimFloat(seconds / 60)
				case "secs":
					return trimFloat(seconds)
				}
			}
			total := int64(math.Round(seconds))
			return fmt.Sprintf("%dhour, %dmins, %dsec", total/3600, (total%3600)/60, total%60)
		case "duration hour":
			return padNumber(int64(seconds)/3600, firstArg(args))
		case "duration min":
			return padNumber((int64(seconds)%3600)/60, firstArg(args))
		case "duration sec":
			return padNumber(int64(seconds)%60, firstArg(args))
		}
	}
	if strings.HasPrefix(name, "video date") || strings.HasPrefix(name, "video time") {
		value := metadataValue(ctx.Metadata, "video date")
		t, ok := parseMetadataTime(value)
		if !ok {
			return ""
		}
		switch name {
		case "video date":
			return formatAdvancedDate(t, joinOr(args, "yyyy-mm-dd"))
		case "video time":
			return formatAdvancedDate(t, joinOr(args, "hh:nn:ss"))
		case "video date year":
			return renderDateTag("year", args, t)
		case "video date month":
			return renderDateTag("month", args, t)
		case "video date day":
			return renderDateTag("day", args, t)
		case "video date hour":
			return renderDateTag("hour", args, t)
		case "video date min":
			return renderDateTag("min", args, t)
		case "video date sec":
			return renderDateTag("sec", args, t)
		}
	}
	return metadataValue(ctx.Metadata, name)
}

func parseMetadataTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	layouts := []string{
		"2006:01:02 15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func applyModifiers(value string, mods []string, ctx *TemplateContext) (string, error) {
	for i := 0; i < len(mods); {
		name := strings.ToLower(strings.TrimSpace(mods[i]))
		i++
		switch name {
		case "default":
			arg := nextModifierArg(mods, &i)
			if value == "" {
				value = arg
			}
		case "alt":
			arg := nextModifierArg(mods, &i)
			if value == "" && arg != "" {
				alt, err := resolveTagWithModifiers(arg, ctx)
				if err != nil {
					return "", err
				}
				value = alt
			}
		case "append":
			arg := nextModifierArg(mods, &i)
			if value != "" {
				value += arg
			}
		case "prepend":
			arg := nextModifierArg(mods, &i)
			if value != "" {
				value = arg + value
			}
		case "suffix":
			arg := nextModifierArg(mods, &i)
			if value != "" && !strings.HasSuffix(value, arg) {
				value += arg
			}
		case "prefix":
			arg := nextModifierArg(mods, &i)
			if value != "" && !strings.HasPrefix(value, arg) {
				value = arg + value
			}
		case "pad":
			lengthText := nextModifierArg(mods, &i)
			char := nextModifierArgOptional(mods, &i, " ")
			length, _ := strconv.Atoi(lengthText)
			if length > len([]rune(value)) {
				pad := strings.Repeat(char, length-len([]rune(value)))
				value = pad + value
			}
		case "trim":
			value = strings.TrimSpace(value)
		case "upper":
			pos, count := optionalPositionCount(mods, &i)
			value = transformRange(value, pos, count, strings.ToUpper)
		case "lower":
			pos, count := optionalPositionCount(mods, &i)
			value = transformRange(value, pos, count, strings.ToLower)
		case "titlecase":
			pos, count := optionalPositionCount(mods, &i)
			value = transformRange(value, pos, count, titleCase)
		case "substr":
			pos := nextModifierArg(mods, &i)
			count := nextModifierArg(mods, &i)
			value = substringTag(value, []string{pos, count}, false)
		case "rsubstr":
			pos := nextModifierArg(mods, &i)
			count := nextModifierArg(mods, &i)
			value = substringTag(value, []string{pos, count}, true)
		case "insert":
			text := nextModifierArg(mods, &i)
			posText := nextModifierArg(mods, &i)
			pos, _ := strconv.Atoi(posText)
			value = insertAtRune(value, text, pos)
		case "remove":
			posText := nextModifierArg(mods, &i)
			countText := nextModifierArg(mods, &i)
			pos, _ := strconv.Atoi(posText)
			count, _ := strconv.Atoi(countText)
			value = removeRuneRange(value, pos, count)
		case "replace":
			old := nextModifierArg(mods, &i)
			newV := nextModifierArg(mods, &i)
			value = strings.ReplaceAll(value, old, newV)
		case "word":
			index := nextModifierArg(mods, &i)
			count := nextModifierArgOptional(mods, &i, "1")
			separator := nextModifierArgOptional(mods, &i, "")
			args := []string{index, count}
			if separator != "" {
				args = append(args, separator)
			}
			value = wordTag(value, args, false)
		case "rword":
			index := nextModifierArg(mods, &i)
			count := nextModifierArgOptional(mods, &i, "1")
			separator := nextModifierArgOptional(mods, &i, "")
			args := []string{index, count}
			if separator != "" {
				args = append(args, separator)
			}
			value = wordTag(value, args, true)
		case "add", "subtract", "multiply", "divide":
			arg := nextModifierArg(mods, &i)
			value = numericModifier(value, name, arg)
		default:
			return "", fmt.Errorf("unknown tag modifier %q", name)
		}
	}
	return value, nil
}

func nextModifierArg(mods []string, index *int) string {
	if *index >= len(mods) {
		return ""
	}
	value := mods[*index]
	*index = *index + 1
	return value
}

func nextModifierArgOptional(mods []string, index *int, fallback string) string {
	if *index >= len(mods) || isModifier(mods[*index]) {
		return fallback
	}
	value := mods[*index]
	*index = *index + 1
	return value
}

func optionalPositionCount(mods []string, index *int) (int, int) {
	pos, count := 1, -1
	if *index < len(mods) && !isModifier(mods[*index]) {
		if n, err := strconv.Atoi(mods[*index]); err == nil {
			pos = n
			*index = *index + 1
		}
	}
	if *index < len(mods) && !isModifier(mods[*index]) {
		if n, err := strconv.Atoi(mods[*index]); err == nil {
			count = n
			*index = *index + 1
		}
	}
	return pos, count
}

func transformRange(value string, pos, count int, transform func(string) string) string {
	if count < 0 {
		return transform(value)
	}
	runes := []rune(value)
	if pos < 1 {
		pos = 1
	}
	start := pos - 1
	if start >= len(runes) {
		return value
	}
	end := start + count
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[:start]) + transform(string(runes[start:end])) + string(runes[end:])
}

func insertAtRune(value, text string, pos int) string {
	runes := []rune(value)
	if pos < 1 {
		pos = 1
	}
	index := pos - 1
	if index > len(runes) {
		index = len(runes)
	}
	return string(runes[:index]) + text + string(runes[index:])
}

func numericModifier(value, op, arg string) string {
	v, err1 := strconv.ParseFloat(strings.TrimSpace(value), 64)
	n, err2 := strconv.ParseFloat(strings.TrimSpace(arg), 64)
	if err1 != nil || err2 != nil {
		return value
	}
	switch op {
	case "add":
		v += n
	case "subtract":
		v -= n
	case "multiply":
		v *= n
	case "divide":
		if n == 0 {
			return value
		}
		v /= n
	}
	return trimFloat(v)
}

func trimFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func checksumFile(path, algorithm string) (string, error) {
	if path == "" {
		return "", nil
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	switch algorithm {
	case "md5":
		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	case "sha1":
		h := sha1.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	}
	return "", nil
}

func readFileLine(path string, line int) string {
	if line < 1 || path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if line > len(lines) {
		return ""
	}
	return lines[line-1]
}

func readFileContentSlice(path string, args []string) string {
	if path == "" || len(args) < 2 {
		return ""
	}
	pos, _ := strconv.Atoi(strings.TrimSpace(args[0]))
	count, _ := strconv.Atoi(strings.TrimSpace(args[1]))
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	runes := []rune(string(data))
	if pos < 1 {
		pos = 1
	}
	start := pos - 1
	if start >= len(runes) || count <= 0 {
		return ""
	}
	end := start + count
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

func metadataValue(metadata map[string]string, key string) string {
	if metadata == nil {
		return ""
	}
	if value, ok := metadata[strings.ToLower(strings.TrimSpace(key))]; ok {
		return value
	}
	return ""
}

func padNumber(value int64, spec string) string {
	if spec == "" {
		return strconv.FormatInt(value, 10)
	}
	width := len(spec)
	if n, err := strconv.Atoi(spec); err == nil && n > 0 && !strings.HasPrefix(spec, "0") {
		width = n
	}
	if width <= 0 {
		return strconv.FormatInt(value, 10)
	}
	return fmt.Sprintf("%0*d", width, value)
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.TrimSpace(args[0])
}

func detectMediaType(ext string) string {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	switch ext {
	case "jpg", "jpeg", "jpe", "jfif", "png", "gif", "bmp", "webp", "tif", "tiff", "heic", "heif", "avif", "dng", "raw":
		return "image"
	case "mp3", "flac", "wav", "m4a", "aac", "ogg", "opus", "wma", "aiff", "ape":
		return "audio"
	case "mp4", "mkv", "avi", "mov", "wmv", "webm", "m4v", "mpeg", "mpg", "3gp", "ts", "mts", "m2ts":
		return "video"
	case "pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "rtf", "odt", "ods", "odp", "csv", "md", "epub", "eml":
		return "document"
	default:
		return "other"
	}
}

func humanFileSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d %s", size, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func joinOr(parts []string, fallback string) string {
	if len(parts) == 0 {
		return fallback
	}
	return strings.Join(parts, ":")
}
