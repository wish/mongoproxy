package integrationtest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSTests(t *testing.T) {
	// walk directory for tests
	// Find all tests
	err := filepath.Walk("jstests", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		// Tests are only .js files
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".js") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			cmd := exec.Command("mongosh", "localhost:27016", path)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("%s\n", stdoutStderr)
				t.Fatal(err)
			}
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
