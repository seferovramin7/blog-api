package main

import (
	"fmt"
	"unicode"
)

func numDecodings(s string) int {
	if !isValidInput(s) {
		return 0
	}

	n := len(s)
	if n == 0 {
		return 0
	}

	// Ways to decode using two steps ahead and one step ahead.
	waysTwoStepsAhead := 1
	waysOneStepAhead := 1

	for i := n - 1; i >= 0; i-- {
		currentWays := 0
		if s[i] != '0' {
			currentWays = waysOneStepAhead // Single-digit decode
			if i+1 < n && isValidTwoDigit(s[i], s[i+1]) {
				currentWays += waysTwoStepsAhead // Two-digit decode
			}
		}
		// Slide the window forward
		waysTwoStepsAhead, waysOneStepAhead = waysOneStepAhead, currentWays
	}

	return waysOneStepAhead
}

func isValidInput(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func isValidTwoDigit(c1, c2 byte) bool {
	twoDigitValue := (c1-'0')*10 + (c2 - '0')
	return twoDigitValue >= 10 && twoDigitValue <= 26
}

func runTests() {
	testCases := []struct {
		input    string
		expected int
	}{
		// Standard cases
		{"12", 2},         // "AB", "L"
		{"226", 3},        // "BZ", "VF", "BBF"
		{"0", 0},          // No valid decoding
		{"10", 1},         // "J"
		{"27", 1},         // "BG"
		{"11106", 2},      // "AAJF", "KJF"
		{"1234567890", 0}, // No valid decoding

		// Edge cases
		{"", 0},                       // Empty string
		{"111111", 13},                // Dense decoding
		{"000", 0},                    // Only zeros
		{"30", 0},                     // Invalid two-digit case
		{"1111111111111111111", 6765}, // Large input
		{"1a2", 0},                    // Invalid characters
	}

	for _, tc := range testCases {
		output := numDecodings(tc.input)
		if output != tc.expected {
			fmt.Printf("❌ Test Failed: Input: '%s', Expected: %d, Got: %d\n",
				tc.input, tc.expected, output)
		} else {
			fmt.Printf("✅ Test Passed: Input: '%s', Output: %d\n", tc.input, output)
		}
	}
}

func main() {
	runTests()
}
