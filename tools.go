package pjson5

import (
	"strings"
	"unicode"
)

func skipWhiteSpace(s string, pos int) (int, bool) {
	containsLineBreaker := false
	for pos < len(s) {
		isLB := isLineBreaker(s[pos])
		containsLineBreaker = containsLineBreaker || isLB
		if !isLB && !isWhitespaceNLB(s[pos]) {
			break
		}
		pos++
	}
	return pos, containsLineBreaker
}

func skipLineWhiteSpace(s string, pos int) int {
	for pos < len(s) && isWhitespaceNLB(s[pos]) {
		pos++
	}
	return pos
}

func isWhitespaceNLB(c byte) bool {
	return c == ' ' || c == '\t'
}

func isLineBreaker(c byte) bool {
	return c == '\r' || c == '\n'
}

func trimStringPart(raw string, endPos int, l int) string {
	startPos := endPos - l
	if startPos < 0 {
		startPos = 0
	}
	return raw[startPos:endPos]
}

// findEndOfNumber 用于找到字符串中有效数字的结尾位置
func findEndOfNumber(s string) int {
	sl := len(s)

	if sl == 0 {
		return 0
	}
	// 检查是否为正负无穷或 NaN
	if sl >= 8 && strings.EqualFold(s[:8], "INFINITY") {
		return 8
	}
	if s[0] == '-' && sl >= 9 && strings.EqualFold(s[:9], "-INFINITY") {
		return 9
	}
	if sl >= 3 && strings.EqualFold(s[:3], "NAN") {
		return 3
	}

	// 用于标记是否已经出现过小数点
	hasDot := false
	// 用于标记是否已经出现过 e 或 E（科学计数法）
	hasE := false
	// 用于标记 e 或 E 后面是否有符号
	hasESign := false
	// 用于标记是否是十六进制
	isHex := false
	// 用于标记是否是八进制
	isOctal := false

	// 检查是否为十六进制或八进制
	if len(s) >= 2 && s[0] == '0' {
		if s[1] == 'x' || s[1] == 'X' {
			isHex = true
		} else if s[1] == 'o' || s[1] == 'O' {
			isOctal = true
		}
	}

	// 遍历字符串中的每个字符
	for i, r := range s {
		switch {
		// 如果是十六进制
		case isHex:
			if i == 0 || i == 1 {
				continue
			}
			if !unicode.IsDigit(r) && !((r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return i
			}
		// 如果是八进制
		case isOctal:
			if i == 0 || i == 1 {
				continue
			}
			if r < '0' || r > '7' {
				return i
			}
		// 如果是数字字符（0 - 9）
		case unicode.IsDigit(r):
			continue
		// 如果是小数点
		case r == '.':
			// 如果已经出现过小数点或者已经出现过 e 或 E，则数字结束
			if hasDot || hasE {
				return i
			}
			// 标记已经出现过小数点
			hasDot = true
		// 如果是 e 或 E
		case r == 'e' || r == 'E':
			// 如果已经出现过 e 或 E，则数字结束
			if hasE {
				return i
			}
			// 标记已经出现过 e 或 E
			hasE = true
			// 重置 e 后面是否有符号的标记
			hasESign = false
		// 如果是正负号
		case r == '+' || r == '-':
			// 如果不是在字符串开头或者 e 或 E 后面，则数字结束
			if i != 0 && !hasE || hasESign {
				return i
			}
			// 如果是在 e 或 E 后面，标记已经出现过符号
			if hasE {
				hasESign = true
			}
		default:
			// 遇到其他字符，数字结束
			return i
		}
	}
	// 遍历完整个字符串，返回字符串的长度
	return len(s)
}
