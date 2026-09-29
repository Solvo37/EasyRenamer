package engine

import (
	"os"
	"path/filepath"
	"strings"
)

func populateTemplateContext(ctx *TemplateContext, tpl string) {
	if ctx == nil {
		return
	}
	if ctx.Metadata == nil {
		ctx.Metadata = make(map[string]string)
	}
	if ctx.Path == "" {
		if ctx.MediaType == "" {
			ctx.MediaType = detectMediaType(ctx.Extension)
		}
		return
	}

	if st, err := os.Stat(ctx.Path); err == nil {
		if ctx.FileSize == 0 {
			ctx.FileSize = st.Size()
		}
		if ctx.ModifiedTime.IsZero() {
			ctx.ModifiedTime = st.ModTime()
		}
		if ctx.CreatedTime.IsZero() {
			ctx.CreatedTime = createdTime(ctx.Path, st)
		}
	}
	if ctx.MediaType == "" {
		ctx.MediaType = detectMediaType(filepath.Ext(ctx.Path))
	}

	lower := strings.ToLower(tpl)
	needImage := ctx.MediaType == "image" && (strings.Contains(lower, "<width") ||
		strings.Contains(lower, "<height") || strings.Contains(lower, "<img ") ||
		strings.Contains(lower, "<author") || strings.Contains(lower, "<copyright") ||
		strings.Contains(lower, "<subject") || strings.Contains(lower, "<title") ||
		strings.Contains(lower, "<gps "))
	needVideo := ctx.MediaType == "video" && (strings.Contains(lower, "<width") ||
		strings.Contains(lower, "<height") || strings.Contains(lower, "<video ") ||
		strings.Contains(lower, "<duration") || strings.Contains(lower, "<framerate") ||
		strings.Contains(lower, "<title") || strings.Contains(lower, "<genre"))
	needAudio := ctx.MediaType == "audio" && (strings.Contains(lower, "<album") ||
		strings.Contains(lower, "<artist") || strings.Contains(lower, "<genre") ||
		strings.Contains(lower, "<title") || strings.Contains(lower, "<audio year") ||
		strings.Contains(lower, "<track") || strings.Contains(lower, "<disc"))
	needDocs := ctx.MediaType == "document" && (strings.Contains(lower, "<pages") ||
		strings.Contains(lower, "<creator") || strings.Contains(lower, "<title") ||
		strings.Contains(lower, "<subject") || strings.Contains(lower, "<from") ||
		strings.Contains(lower, "<to") || strings.Contains(lower, "<cc") ||
		strings.Contains(lower, "<bcc") || strings.Contains(lower, "<date"))
	needExe := strings.EqualFold(filepath.Ext(ctx.Path), ".exe") &&
		(strings.Contains(lower, "<exe ") || strings.Contains(lower, "<description"))
	needGenericMetadata := strings.Contains(lower, "<metadata:")

	if needImage {
		if ctx.Width == 0 || ctx.Height == 0 {
			if w, h, ok := readImageDimensions(ctx.Path); ok {
				ctx.Width, ctx.Height = w, h
			}
		}
		mergeMetadata(ctx.Metadata, readEXIFMetadata(ctx.Path))
	}
	if needVideo {
		meta := readVideoMetadata(ctx.Path)
		mergeMetadata(ctx.Metadata, meta)
		if ctx.Width == 0 {
			ctx.Width = atoiMetadata(meta, "width")
		}
		if ctx.Height == 0 {
			ctx.Height = atoiMetadata(meta, "height")
		}
	}
	if needAudio {
		mergeMetadata(ctx.Metadata, readAudioMetadata(ctx.Path))
	}
	if needDocs {
		mergeMetadata(ctx.Metadata, readDocumentMetadata(ctx.Path))
	}
	if needExe {
		mergeMetadata(ctx.Metadata, readExecutableMetadata(ctx.Path))
	}
	if needGenericMetadata {
		mergeMetadata(ctx.Metadata, readEXIFMetadata(ctx.Path))
		mergeMetadata(ctx.Metadata, readAudioMetadata(ctx.Path))
		mergeMetadata(ctx.Metadata, readVideoMetadata(ctx.Path))
		mergeMetadata(ctx.Metadata, readDocumentMetadata(ctx.Path))
		mergeMetadata(ctx.Metadata, readExecutableMetadata(ctx.Path))
	}
}

func mergeMetadata(dst, src map[string]string) {
	for key, value := range src {
		if value == "" {
			continue
		}
		dst[strings.ToLower(strings.TrimSpace(key))] = value
	}
}

func atoiMetadata(meta map[string]string, key string) int {
	value := meta[strings.ToLower(key)]
	n := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
