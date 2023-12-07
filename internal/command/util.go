package command

import (
	"regexp"
	"unicode"
)

var (
	sourceExpression = regexp.MustCompile(`^(?:(?P<apiVersion>(?:[a-z0-9-.]+/)?[a-z0-9-.]+):)?(?P<kind>[A-Za-z0-9-.]+):(?:(?P<namespace>[a-z0-9-.]+)/)?(?P<name>[a-z0-9-.]+)(?:$|[?].*$)`)
	disallowedChars  = regexp.MustCompile(`[^a-z0-9-]`)
)

func isDisallowedStartEndChar(rune rune) bool {
	return !unicode.IsLetter(rune) && !unicode.IsNumber(rune)
}
