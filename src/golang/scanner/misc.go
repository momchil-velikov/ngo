package scanner

import "unicode"

func isWhitespace(ch rune) bool {
	switch ch {
	case SPACE, TAB, NL, CR:
		return true
	default:
		return false
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' ||
		'A' <= ch && ch <= 'Z' ||
		ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

func isDecDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isOctDigit(ch rune) bool {
	return '0' <= ch && ch <= '7'
}

func isHexDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}

func hexValue(ch rune) uint {
	switch {
	case '0' <= ch && ch <= '9':
		return uint(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return 10 + uint(ch-'a')
	case 'A' <= ch && ch <= 'F':
		return 10 + uint(ch-'A')
	default:
		return 0 // cannot happen
	}
}
