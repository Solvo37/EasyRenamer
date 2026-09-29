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
	BaseName    string
	Extension   string
	ParentDir   string
	DirIndex    int
	GlobalIndex int
	BatchTime   time.Time
	Rand        *rand.Rand
}

var tokenRE = regexp.MustCompile(`<([^<>]+)>`)

func RenderTemplate(tpl string, ctx TemplateContext) (string, error) {
	if ctx.BatchTime.IsZero() {
		ctx.BatchTime = time.Now()
	}
	if ctx.Rand == nil {
		ctx.Rand = rand.New(rand.NewSource(ctx.BatchTime.UnixNano()))
	}

	var renderErr error
	result := tokenRE.ReplaceAllStringFunc(tpl, func(raw string) string {
		if renderErr != nil {
			return raw
		}
		body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "<"), ">"))
		val, err := renderToken(body, ctx)
		if err != nil {
			renderErr = err
			return raw
		}
		return val
	})
	if renderErr != nil {
		return "", renderErr
	}

	if !containsToken(tpl, "Ext") {
		result += ctx.Extension
	}
	return result, nil
}

func containsToken(tpl, token string) bool {
	return strings.Contains(strings.ToLower(tpl), "<"+strings.ToLower(token)+">")
}

func renderToken(body string, ctx TemplateContext) (string, error) {
	lower := strings.ToLower(body)
	switch lower {
	case "name":
		return ctx.BaseName, nil
	case "ext":
		return strings.TrimPrefix(ctx.Extension, "."), nil
	case "unixtimestamp", "unix":
		return strconv.FormatInt(ctx.BatchTime.Unix(), 10), nil
	case "dirname:1", "parent":
		return ctx.ParentDir, nil
	case "rand":
		return strconv.Itoa(ctx.Rand.Intn(10)), nil
	}

	if strings.HasPrefix(lower, "inc nrdir:") {
		fmtSpec := strings.TrimSpace(body[len("Inc NrDir:"):])
		return formatCounter(ctx.DirIndex, fmtSpec), nil
	}
	if strings.HasPrefix(lower, "inc:") {
		fmtSpec := strings.TrimSpace(body[len("Inc:"):])
		return formatCounter(ctx.GlobalIndex, fmtSpec), nil
	}
	if strings.HasPrefix(lower, "rand str:") {
		n, err := strconv.Atoi(strings.TrimSpace(body[len("Rand Str:"):]))
		if err != nil || n < 1 || n > 128 {
			return "", fmt.Errorf("invalid random string length in <%s>", body)
		}
		return randomChars(ctx.Rand, n, "abcdefghijklmnopqrstuvwxyz0123456789"), nil
	}
	if strings.HasPrefix(lower, "rand alpha:") {
		n, err := strconv.Atoi(strings.TrimSpace(body[len("Rand Alpha:"):]))
		if err != nil || n < 1 || n > 128 {
			return "", fmt.Errorf("invalid random alpha length in <%s>", body)
		}
		return randomChars(ctx.Rand, n, "abcdefghijklmnopqrstuvwxyz"), nil
	}
	if strings.HasPrefix(lower, "date:") {
		pattern := strings.TrimSpace(body[len("Date:"):])
		return ctx.BatchTime.Format(toGoTimeLayout(pattern)), nil
	}

	return "", fmt.Errorf("unknown token <%s>", body)
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
