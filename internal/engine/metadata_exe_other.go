//go:build !windows

package engine

func readExecutableMetadata(path string) map[string]string {
	return map[string]string{}
}
