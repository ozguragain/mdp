package inline

import (
	"html"
	"strings"
)

// allowedSchemes are the only URL schemes allowed in an href.
var allowedSchemes = map[string]bool{
	"http":   true,
	"https":  true,
	"mailto": true,
}

// isSafeURL checks both the raw form and its HTML-entity-decoded form
// against the scheme allowlist, so obfuscations like "&#106;avascript:" are
// caught as well.
func isSafeURL(raw string) bool {
	return schemeOK(raw) && schemeOK(html.UnescapeString(raw))
}

// schemeOK checks s against the allowlist after dropping bytes <= 0x20 and
// lowercasing, so split schemes like "java\tscript:" are caught too.
func schemeOK(s string) bool {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] > 0x20 {
			b.WriteByte(s[i])
		}
	}
	cleaned := strings.ToLower(b.String())

	// A scheme is present only when ":" comes before any "/", "?" or "#"
	// and the prefix is a valid scheme name; otherwise the reference is
	// relative and safe.
	colon := strings.IndexByte(cleaned, ':')
	if colon < 0 {
		return true
	}
	if cut := strings.IndexAny(cleaned, "/?#"); cut >= 0 && cut < colon {
		return true
	}
	scheme := cleaned[:colon]
	if !isSchemeName(scheme) {
		return true
	}
	return allowedSchemes[scheme]
}

// isSchemeName reports whether s (already lowercased) matches
// [A-Za-z][A-Za-z0-9+.-]*.
func isSchemeName(s string) bool {
	if s == "" || !isAlpha(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !isAlpha(c) && (c < '0' || c > '9') && c != '+' && c != '-' && c != '.' {
			return false
		}
	}
	return true
}

func isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z'
}
