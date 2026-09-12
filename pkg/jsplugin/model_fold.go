package jsplugin

// asciiFold lowercases ASCII letters only. Unicode case folding is avoided so
// non-ASCII characters cannot impersonate an ASCII model name.
func asciiFold(s string) string {
	var buf []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			if buf == nil {
				buf = []byte(s)
			}
			buf[i] = c + ('a' - 'A')
		}
	}
	if buf == nil {
		return s
	}
	return string(buf)
}

// ASCIIFold is the exported form for consumers outside this package.
func ASCIIFold(s string) string {
	return asciiFold(s)
}
