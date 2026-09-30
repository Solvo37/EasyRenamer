package engine

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TemplateContext struct {
	Path         string
	OriginalName string
	BaseName     string
	Extension    string
	ParentDir    string
	DirIndex     int
	GlobalIndex  int
	TotalItems   int
	BatchTime    time.Time
	ModifiedTime time.Time
	CreatedTime  time.Time
	FileSize     int64
	Width        int
	Height       int
	MediaType    string
	Metadata     map[string]string
	Rand         *rand.Rand
}

var tokenRE = regexp.MustCompile(`<([^<>]+)>`)

func RenderTemplate(tpl string, ctx TemplateContext) (string, error) {
	if ctx.BatchTime.IsZero() {
		ctx.BatchTime = time.Now()
	}
	if ctx.Rand == nil {
		ctx.Rand = rand.New(rand.NewSource(ctx.BatchTime.UnixNano()))
	}
	if ctx.OriginalName == "" {
		ctx.OriginalName = ctx.BaseName + ctx.Extension
	}
	populateTemplateContext(&ctx, tpl)

	var renderErr error
	result := tokenRE.ReplaceAllStringFunc(tpl, func(raw string) string {
		if renderErr != nil {
			return raw
		}
		body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "<"), ">"))
		val, err := renderToken(body, &ctx)
		if err != nil {
			renderErr = err
			return raw
		}
		return val
	})
	if renderErr != nil {
		return "", renderErr
	}

	if !templateContainsExtensionTag(tpl) && filepath.Ext(result) == "" {
		result += ctx.Extension
	}
	return result, nil
}

func templateContainsExtensionTag(tpl string) bool {
	for _, match := range tokenRE.FindAllStringSubmatch(tpl, -1) {
		if len(match) < 2 {
			continue
		}
		body := strings.TrimSpace(match[1])
		lower := strings.ToLower(body)
		if lower == "ext" || strings.HasPrefix(lower, "ext:") || strings.HasPrefix(lower, "ext|") {
			return true
		}
	}
	return false
}

func formatCounterWithStep(index int, spec string, defaultStart, defaultStep int) string {
	if index < 1 {
		index = 1
	}
	start := defaultStart
	step := defaultStep
	width := 0
	parts := strings.Split(spec, ":")
	if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
		startText := strings.TrimSpace(parts[0])
		width = len(startText)
		if n, err := strconv.Atoi(startText); err == nil {
			start = n
		}
	}
	if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			step = n
		}
	}
	value := start + (index-1)*step
	if width > 1 && strings.HasPrefix(strings.TrimSpace(parts[0]), "0") {
		return fmt.Sprintf("%0*d", width, value)
	}
	return strconv.Itoa(value)
}

func formatCounter(index int, spec string) string {
	if index < 1 {
		index = 1
	}
	if spec == "" {
		return strconv.Itoa(index)
	}
	width := len(spec)
	start, err := strconv.Atoi(spec)
	if err != nil {
		start = 1
	}
	value := start + index - 1
	return fmt.Sprintf("%0*d", width, value)
}

func randomChars(r *rand.Rand, n int, alphabet string) string {
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[r.Intn(len(alphabet))])
	}
	return b.String()
}

func toGoTimeLayout(pattern string) string {
	r := strings.NewReplacer(
		"yyyy", "2006",
		"yy", "06",
		"MM", "01",
		"dd", "02",
		"HH", "15",
		"mm", "04",
		"ss", "05",
	)
	return r.Replace(pattern)
}

func BaseAndExt(name string) (string, string) {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext), ext
}
