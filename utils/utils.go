package utils

import "strings"

func StripExtraWS(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")

	// remove all the extra spaces
	for {
		ss := strings.ReplaceAll(s, "  ", " ")
		if s == ss {
			break
		}
		s = ss
	}

	return s
}
