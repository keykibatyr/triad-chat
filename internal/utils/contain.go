package utils

import "regexp"

func ContainsAi(content string) bool {
	re := regexp.MustCompile(`^/ai`)

	match := re.FindString(content)
	if match != "" {
		return true
	}

	return false
}	