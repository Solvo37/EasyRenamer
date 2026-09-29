package i18n

import (
	"strings"
	"testing"
)

func TestAllLanguagesCoverEnglishUIKeys(t *testing.T) {
	base := translations[English]
	for _, lang := range supported {
		if lang == English {
			continue
		}
		table := translations[lang]
		for key := range base {
			if strings.TrimSpace(table[key]) == "" {
				t.Errorf("%s is missing translation key %q", lang, key)
			}
		}
	}
}
