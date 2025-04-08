package sanitize

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	//	Non-breaking space: https://www.compart.com/en/unicode/U+00A0
	NBSP_UTF8		= "\xC2\xA0"
	//	Zero-width space: https://www.compart.com/en/unicode/U+200B
	ZWSP_UTF8		= "\xE2\x80\x8B"
	
	//	Soft hyphen: https://www.compart.com/en/unicode/U+00AD
	SHY_UTF8		= "\xC2\xAD"
	//	Non-breaking hyphen: https://www.compart.com/en/unicode/U+2011
	NBHY_UTF8		= "\xE2\x80\x91"
	//	En dash: https://www.compart.com/en/unicode/U+2013
	ENDASH_UTF8		= "\xE2\x80\x93"
)

var (
	re_space 			= regexp.MustCompile(NBSP_UTF8+`|`+ZWSP_UTF8)
	re_hyphen 			= regexp.MustCompile(SHY_UTF8+`|`+NBHY_UTF8+`|`+ENDASH_UTF8)
	re_reduce_newlines 	= regexp.MustCompile(`\n{3,}`)
	re_reduce_spaces 	= regexp.MustCompile(` +`)
	
	//	https://www.regular-expressions.info/unicode.html
	re_printable 		= regexp.MustCompile(`[\P{C}\n\r\t]`)
	re_non_printable	= regexp.MustCompile(`[^\P{C}\n\r\t]`)
)

func Filter_utf8mb3(s string) string {
	//	Strip invalid UTF8 chars
	s = strings.ToValidUTF8(s, "")
	//	Remvoe NULL bytes
	s = strings.Replace(s, "\x00", "", -1)
	//	Strip chars with more than 3 bytes
	l := len(s)
	out := make([]rune, 0, l)
	for i := 0; i < l; {
		r, n := utf8.DecodeRune([]byte(s[i:]))
		if n < 4 {
			out = append(out, r)
		}
		i += n
	}
	return string(out)
}

func Trim(s string, allow_newlines bool) string {
	s = normalize(s)
	
	has_newline := strings.Contains(s, "\n")
	if has_newline {
		if allow_newlines {
			lines := strings.Split(s, "\n")
			for i, v := range lines {
				lines[i] = strings.TrimSpace(v)
			}
			s = strings.Join(lines, "\n")
		} else {
			s = strings.Replace(s, "\n", " ", -1)
			s = strings.TrimSpace(s)
		}
	}
	
	if has_newline && allow_newlines && strings.Contains(s, "\n\n\n") {
		s = re_reduce_newlines.ReplaceAllString(s, "\n\n")
	}
	
	if strings.Contains(s, "  ") {
		s = re_reduce_spaces.ReplaceAllString(s, " ")
	}
	
	return s
}

func Non_printable(s string) string {
	return re_printable.ReplaceAllString(s, "")
}

func Strip_non_printable(s string) string {
	return re_non_printable.ReplaceAllString(s, "")
}

func normalize(s string) string {
	//	Remove carriage return
	s = strings.Replace(s, "\r", "", -1)
	s = re_space.ReplaceAllString(s, " ")
	s = re_hyphen.ReplaceAllString(s, "-")
	return s
}