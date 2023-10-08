package matcher

// Match compares a string with a pattern that supports '*' and '?' wildcard characters
// '*' matches any sequence of characters, including empty sequence
// '?' matches any single character
// Source: https://www.geeksforgeeks.org/wildcard-pattern-matching/
func Match(s string, p string) bool {
	var (
		sIndex = 0
		pIndex = 0

		// wIndex stores the last '*' wildcard index
		wIndex = -1
		// bIndex stores the last matched character from the text
		bIndex = -1
		// nIndex stores the next character index after the last '*' wildcard
		nIndex = -1
	)

	var (
		sLen = len(s)
		pLen = len(p)
	)

	for sIndex < sLen {
		if pIndex < pLen && (p[pIndex] == '?' || p[pIndex] == s[sIndex]) {
			// Either same character, or pattern has the '?' wildcard
			sIndex++
			pIndex++
		} else if pIndex < pLen && p[pIndex] == '*' {
			// Not the same character, but pattern has '*' wildcard
			wIndex = pIndex
			nIndex = pIndex + 1
			bIndex = sIndex

			pIndex++
		} else if wIndex == -1 {
			// Characters don't match and we didn't find any wildcard to cover this mismatch
			return false
		} else {
			// Characters don't match, but we found a wildcard to cover this mismatch
			pIndex = nIndex
			sIndex = bIndex + 1

			bIndex++
		}
	}

	// If we finished the text, but still have characters in the pattern,
	// they all must be the '*' wildcard to match
	for _, c := range p[pIndex:] {
		if c != '*' {
			return false
		}
	}

	return true
}
