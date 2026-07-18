package stacktrace_test

import (
	"path/filepath"
	"testing"

	"github.com/NdoleStudio/stacktrace"
)

func fixturePath(name string) string {
	return filepath.Join("github.com", "NdoleStudio", "stacktrace", name)
}

func useFixturePaths(t *testing.T) {
	t.Helper()

	original := stacktrace.CleanPath
	stacktrace.CleanPath = func(path string) string {
		return fixturePath(filepath.Base(path))
	}
	t.Cleanup(func() {
		stacktrace.CleanPath = original
	})
}
