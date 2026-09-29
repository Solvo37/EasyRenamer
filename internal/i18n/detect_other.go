//go:build !windows

package i18n

import "os"

func detectSystemLanguage() Language {
	if value := os.Getenv("LC_ALL"); value != "" {
		return languageFromLocale(value)
	}
	if value := os.Getenv("LC_MESSAGES"); value != "" {
		return languageFromLocale(value)
	}
	return languageFromLocale(os.Getenv("LANG"))
}
