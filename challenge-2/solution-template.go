package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Read input from standard input
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := scanner.Text()

		// Call the ReverseString function
		output := ReverseString(input)

		// Print the result
		fmt.Println(output)
	}
}

// ReverseString returns the reversed string of s.
func ReverseString(s string) string {
	sByte := make([]byte, len(s))

	for i := 0; i < len(s); i++ {
		sByte[len(s)-i-1] = s[i]
	}

	return string(sByte)
}
