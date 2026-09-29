package engine

import "strings"

var categoryExtensions = map[Category]map[string]struct{}{
	CategoryImages:    setOf(".jpg", ".jpeg", ".jpe", ".jfif", ".png", ".webp", ".gif", ".bmp", ".tif", ".tiff", ".heic", ".heif", ".avif", ".raw", ".dng"),
	CategoryVideos:    setOf(".mp4", ".mkv", ".avi", ".mov", ".wmv", ".webm", ".m4v", ".mpeg", ".mpg", ".3gp", ".ts", ".mts", ".m2ts"),
	CategoryAudio:     setOf(".mp3", ".flac", ".wav", ".m4a", ".aac", ".ogg", ".opus", ".wma", ".aiff", ".ape"),
	CategoryDocuments: setOf(".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".rtf", ".odt", ".ods", ".odp", ".csv", ".md", ".epub"),
	CategoryArchives:  setOf(".zip", ".7z", ".rar", ".tar", ".gz", ".bz2", ".xz", ".tgz", ".zst"),
}

func setOf(values ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(values))
	for _, v := range values {
		m[strings.ToLower(v)] = struct{}{}
	}
	return m
}

func MatchesCategory(cat Category, custom, ext string) bool {
	ext = strings.ToLower(ext)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	switch cat {
	case CategoryAll:
		return true
	case CategoryCustom:
		for _, part := range strings.FieldsFunc(custom, func(r rune) bool {
			return r == ',' || r == ';' || r == ' ' || r == '
' || r == '	'
		}) {
			part = strings.TrimSpace(strings.ToLower(part))
			if part == "" {
				continue
			}
			if !strings.HasPrefix(part, ".") {
				part = "." + part
			}
			if part == ext {
				return true
			}
		}
		return false
	default:
		_, ok := categoryExtensions[cat][ext]
		return ok
	}
}

func Categories() []Category {
	return []Category{
		CategoryAll,
		CategoryImages,
		CategoryVideos,
		CategoryAudio,
		CategoryDocuments,
		CategoryArchives,
		CategoryCustom,
	}
}
