package strip

// Removes trailing slashes from a URL
//
// For example: "http://example.com" == TrailingSlashes("http://example.com//")
func TrailingSlashes(url string) string {
	r := []rune(url)
	for len(r) > 0 && r[len(r) - 1] == '/' {
		r = r[:len(r) - 1]
	}

	return string(r)
}
