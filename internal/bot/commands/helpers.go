package commands

import "strings"

func isValidRepoURL(url string) bool {
	return strings.HasPrefix(url, "https://github.com/") ||
		strings.HasPrefix(url, "https://gitlab.com/") ||
		strings.HasPrefix(url, "https://stackoverflow.com/") ||
		strings.HasPrefix(url, "https://reddit.com/") ||
		strings.HasPrefix(url, "https://www.youtube.com/") ||
		strings.HasPrefix(url, "https://youtube.com/")
}
