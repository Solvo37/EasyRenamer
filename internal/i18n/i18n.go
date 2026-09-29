package i18n

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Language string

const (
	English  Language = "en"
	Russian  Language = "ru"
	Spanish  Language = "es"
	Chinese  Language = "zh"
)

var supported = []Language{English, Russian, Spanish, Chinese}

var (
	mu      sync.RWMutex
	current = English
)

func init() {
	current = loadLanguage()
}

func Current() Language {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

func Set(lang Language) error {
	if !IsSupported(lang) {
		lang = English
	}
	mu.Lock()
	current = lang
	mu.Unlock()
	path := settingsPath()
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(lang), 0o644)
}

func Languages() []Language {
	return append([]Language(nil), supported...)
}

func IsSupported(lang Language) bool {
	for _, item := range supported {
		if item == lang {
			return true
		}
	}
	return false
}

func LanguageName(lang Language) string {
	switch lang {
	case Russian:
		return "Русский"
	case Spanish:
		return "Español"
	case Chinese:
		return "中文"
	default:
		return "English"
	}
}

func T(key string) string {
	mu.RLock()
	lang := current
	mu.RUnlock()
	if table, ok := translations[lang]; ok {
		if value := table[key]; value != "" {
			return value
		}
	}
	if value := translations[English][key]; value != "" {
		return value
	}
	return key
}

func loadLanguage() Language {
	if path := settingsPath(); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			lang := Language(strings.TrimSpace(string(data)))
			if IsSupported(lang) {
				return lang
			}
		}
	}
	detected := detectSystemLanguage()
	if IsSupported(detected) {
		return detected
	}
	return English
}

func settingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "EasyRenamer", "language.txt")
}

func languageFromLocale(locale string) Language {
	locale = strings.ToLower(strings.TrimSpace(locale))
	switch {
	case strings.HasPrefix(locale, "ru"):
		return Russian
	case strings.HasPrefix(locale, "es"):
		return Spanish
	case strings.HasPrefix(locale, "zh"):
		return Chinese
	default:
		return English
	}
}
