package inline

import (
	"html"
	"strings"
)

// allowedSchemes are the only URL schemes that may appear in an href.
// Everything else (javascript:, data:, vbscript:, file:, blob:, ...) is
// rejected. Scheme-less references (relative paths, #fragments, ?queries,
// //host) are allowed.
var allowedSchemes = map[string]bool{
	"http":   true,
	"https":  true,
	"mailto": true,
}

// isSafeURL reports whether raw may be emitted into an href attribute.
// Both the raw form and its HTML-entity-decoded form must be safe, so
// obfuscations like "&#106;avascript:" are caught as well.
func isSafeURL(raw string) bool {
	return schemeOK(raw) && schemeOK(html.UnescapeString(raw))
}

// schemeOK reports whether s passes the URL scheme allowlist. Characters
// of 0x20 and below (whitespace and controls) are dropped first so that
// split schemes like "java\tscript:" or "java script:" are caught; the
// comparison is ASCII case-insensitive.
func schemeOK(s string) bool {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] > 0x20 {
			b.WriteByte(s[i])
		}
	}
	cleaned := strings.ToLower(b.String())

	// A scheme is present only when ":" comes before any "/", "?" or "#"
	// and the prefix is a valid scheme name. Anything else is a relative
	// reference and safe.
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
