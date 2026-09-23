package application

import (
	"regexp"
	"strings"
)

// globToRegex переводит glob-шаблон в регулярное выражение, привязанное к всему пути.
// Семантика: `*` не пересекает `/`, `?` — один символ кроме `/`, `**` пересекает границы
// сегментов, а `**/` дополнительно допускает ноль сегментов. Пути сравниваются целиком.
func globToRegex(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); {
		switch c := pattern[i]; c {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					b.WriteString("(?:.*/)?")
					i += 3
					continue
				}
				b.WriteString(".*")
				i += 2
				continue
			}
			b.WriteString("[^/]*")
			i++
		case '?':
			b.WriteString("[^/]")
			i++
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
			i++
		}
	}
	b.WriteString("$")
	return regexp.Compile(b.String())
}
