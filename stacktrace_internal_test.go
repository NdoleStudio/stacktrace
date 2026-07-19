package stacktrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetCallSite(t *testing.T) {
	original := CleanPath
	CleanPath = nil
	t.Cleanup(func() {
		CleanPath = original
	})

	st := &stacktrace{message: "message"}
	setCallSite(st, 0, "file.go", 42, false)
	assert.Empty(t, st.file)

	setCallSite(st, 0, "file.go", 42, true)
	assert.Equal(t, "file.go", st.file)
	assert.Equal(t, 42, st.line)
	assert.Empty(t, st.function)
	assert.Equal(t, "message\n --- at file.go:42 ---", formatFull(st))
}
