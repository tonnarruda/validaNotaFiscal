package api

import "strings"

func DefragmentText(s string) string {
	return strings.ReplaceAll(s, " ", "")
}
