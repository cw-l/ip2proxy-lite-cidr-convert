package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
    cmd := exec.Command("go", "build", "-o", "../../ip2proxyliteconvert", ".")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to build binary: %v\n", err)
        os.Exit(1)
    }
    os.Exit(m.Run())
}

func getPaths(t *testing.T) (string, string) {
	wd, _ := os.Getwd()
	// Since we are in cmd/ip2proxyliteconvert, root is two levels up
	root := filepath.Join(wd, "..", "..")
	bin := filepath.Join(root, "ip2proxyliteconvert")
	data := filepath.Join(root, "testdata", "samples")
	return bin, data
}

func TestHappyMatrix(t *testing.T) {
	binPath, testDataDir := getPaths(t)
	levels := []string{"px1", "px2", "px3", "px4", "px5", "px6", "px7", "px8", "px9", "px10", "px11", "px12"}
	variants := []string{"ipv4", "both"}

	for _, lv := range levels {
		for _, v := range variants {
			testName := fmt.Sprintf("%s_%s", lv, v)
			t.Run(testName, func(t *testing.T) {
				inputFile := filepath.Join(testDataDir, testName+".csv")
				if _, err := os.Stat(inputFile); os.IsNotExist(err) {
					t.Skipf("Missing %s", inputFile)
				}

				outputFile := filepath.Join(t.TempDir(), testName+".mmdb")
				
				// CHANGED: --level to --db to match your binary's flags
				cmd := exec.Command(binPath, "-in", inputFile, "-out", outputFile, "-db", lv)
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("Failed %s\nOutput: %s", testName, out)
				}
			})
		}
	}
}

func TestExceptions(t *testing.T) {
	binPath, testDataDir := getPaths(t)
	cases := []struct {
		file string
		fail bool
	}{
		{"exception_poisoned.csv", true},
		{"exception_threshold.csv", true},
		{"exception_junk.csv", true},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			input := filepath.Join(testDataDir, tc.file)
			if _, err := os.Stat(input); os.IsNotExist(err) {
				t.Skipf("Missing %s", tc.file)
			}

			// CHANGED: --level to --db here as well
			cmd := exec.Command(binPath, "-in", input, "-out", filepath.Join(t.TempDir(), "err.mmdb"), "-db", "px1")
			err := cmd.Run()
			
			if tc.fail && err == nil {
				t.Error("Should have failed (101+ errors threshold)")
			}
			if !tc.fail && err != nil {
				t.Errorf("Should have passed (under 101 errors): %v", err)
			}
		})
	}
}