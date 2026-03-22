package commands

import "strings"

func isValidRepoURL(url string) bool {
	return strings.HasPrefix(url, "https://github.com/") ||
		strings.HasPrefix(url, "https://gitlab.com/")
}
