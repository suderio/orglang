package lexer

func isSignedNumber(s string) bool {
	if len(s) < 2 {
		return false
	}
	// Must start with + or -
	if s[0] != '+' && s[0] != '-' {
		return false
	}
	// Rest must be digits or digits.digits
	hasDot := false
	for i := 1; i < len(s); i++ {
		if s[i] == '.' {
			if hasDot {
				return false // double dot
			}
			hasDot = true
		} else if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	// Cannot end with dot (e.g., "-1.")? Spec says DECIMAL identifier is number.number?
	// Lexer usually handles "1." as INT then DOT.
	// If we return true here, valid flow.
	return true
}
