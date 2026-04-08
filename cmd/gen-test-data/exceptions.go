package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateExceptionData creates the 3 specific test cases we designed.
func GenerateExceptionData(destDir string) {
	// Case 1: Poisoned Rows (2 valid, 2 invalid)
	// Goal: Test that v1 logs errors but finishes.
	poisoned := []string{
		`"1.0.0.0","1.0.0.255","US","United States"`, // Valid
		`"999.999.999.999","XX","Invalid IP"`,        // Junk IP
		`"2.0.0.0/99","US","Invalid CIDR"`,           // Junk CIDR
		`"1.1.1.1","1.1.1.255","AU","Australia"`,     // Valid
	}
	writeLines(filepath.Join(destDir, "exception_poisoned.csv"), poisoned)

	// Case 2: Dirty File (200 total, 101 errors)
	// Goal: Trigger the 101st error threshold break.
	var dirty []string
	for i := 0; i < 99; i++ {
		dirty = append(dirty, `"1.1.1.1","1.1.1.255","US","Good Row"`)
	}
	for i := 0; i < 101; i++ {
		dirty = append(dirty, fmt.Sprintf("junk_data_row_%d,bad,bad,bad", i))
	}
	writeLines(filepath.Join(destDir, "exception_threshold.csv"), dirty)

	// Case 3: Total Junk (101 lines of random garbage)
	// Goal: Immediate failure.
	var totalJunk []string
	for i := 0; i < 101; i++ {
		totalJunk = append(totalJunk, "NOT_A_CSV_LINE_AT_ALL_JUST_RANDOM_GARBAGE_123456789")
	}
	writeLines(filepath.Join(destDir, "exception_junk.csv"), totalJunk)

	fmt.Println("Generated 3 Exception Test Files in", destDir)
}

func writeLines(path string, lines []string) {
	content := strings.Join(lines, "\n") + "\n"
	os.WriteFile(path, []byte(content), 0644)
}