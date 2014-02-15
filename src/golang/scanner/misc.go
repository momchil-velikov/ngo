package scanner

import "unicode"

func is_whitespace(ch rune) bool {
    switch ch {
    case SPACE, TAB, NL, CR:
        return true
    default:
        return false
    }
}

func is_letter(ch rune) bool {
    return ('a' <= ch && ch <= 'z' ||
        'A' <= ch && ch <= 'Z' ||
        ch == '_' || ch >= 0x80 && unicode.IsLetter(ch))
}

func is_digit(ch rune) bool {
    return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

func is_dec_digit(ch rune) bool {
    return '0' <= ch && ch <= '9'
}

func is_oct_digit(ch rune) bool {
    return '0' <= ch && ch <= '7'
}

func is_hex_digit(ch rune) bool {
    return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}

func hex_value(ch rune) uint {
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
