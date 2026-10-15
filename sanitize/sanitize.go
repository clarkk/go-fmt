package sanitize

import (
	"regexp"
	"strings"
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
	re_reduce_newlines 	= regexp.MustCompile(`\n{3,}`)
	re_reduce_spaces 	= regexp.MustCompile(` {2,}`)
	
	//	https://www.regular-expressions.info/unicode.html
	re_printable 		= regexp.MustCompile(`[\P{C}\n\r\t]`)
	re_non_printable	= regexp.MustCompile(`[^\P{C}\n\r\t]`)
	
	normalizer = strings.NewReplacer(
		"\r", "",
		NBSP_UTF8, " ",
		ZWSP_UTF8, " ",
		SHY_UTF8, "",
		NBHY_UTF8, "-",
		ENDASH_UTF8, "-",
	)
)

func Filter_utf8mb3(s string) string {
	s = strings.ToValidUTF8(s, "")
	return strings.Map(func(r rune) rune {
		if r == 0 || r > 0xFFFF {
			return -1
		}
		return r
	}, s)
}

func Trim(s string, allow_newlines bool) string {
	s = normalize(s)
	
	has_newline := strings.Contains(s, "\n")
	if has_newline {
		if allow_newlines {
			lines := strings.Split(s, "\n")
			//	Trim each line individually
			for i, v := range lines {
				lines[i] = strings.TrimSpace(v)
			}
			//	Remove empty lines in the beginning
			for len(lines) > 0 && lines[0] == "" {
				lines = lines[1:]
			}
			//	Remove empty lines in the end
			for len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}
			s = strings.Join(lines, "\n")
		} else {
			s = strings.ReplaceAll(s, "\n", " ")
			s = strings.TrimSpace(s)
		}
	} else {
		s = strings.TrimSpace(s)
	}
	
	//	Reduce consecutive newlines to at most two
	if has_newline && allow_newlines && strings.Contains(s, "\n\n\n") {
		s = re_reduce_newlines.ReplaceAllString(s, "\n\n")
	}
	
	//	Reduce consecutive spaces
	if strings.Contains(s, "  ") {
		s = re_reduce_spaces.ReplaceAllString(s, " ")
	}
	
	return s
}

//	UTF8 safe
func Truncate(s string, length int) string {
	if length <= 0 {
		return ""
	}
	if len(s) <= length {
		return s
	}
	for i := range s {
		if length == 0 {
			return s[:i]
		}
		length--
	}
	return s
}

func Extract_non_printable(s string) string {
	return re_printable.ReplaceAllString(s, "")
}

func Strip_non_printable(s string) string {
	return re_non_printable.ReplaceAllString(s, "")
}

func normalize(s string) string {
	return normalizer.Replace(s)
}