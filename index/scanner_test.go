package index_test

import (
	"bytes"
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat/checksum"
	"scat/index"
)

func TestScannerEmpty(t *testing.T) {
	buf := &bytes.Buffer{}
	scan := index.NewScanner(buf)
	assert.False(t, scan.Next())
	assert.NoError(t, scan.Err())
}

func TestScanner(t *testing.T) {
	buf := &bytes.Buffer{}
	scan := index.NewScanner(buf)

	h1 := checksum.Sum([]byte("a"))
	h2 := checksum.Sum([]byte("b"))
	fmt.Fprintf(buf, "%x 123\n", h1)
	fmt.Fprintf(buf, "%x 456\n", h2)

	assert.True(t, scan.Next())
	assert.NoError(t, scan.Err())
	assert.Equal(t, 0, scan.Chunk().Num())
	assert.Equal(t, h1, scan.Chunk().Hash())
	assert.Equal(t, 123, scan.Chunk().TargetSize())

	assert.True(t, scan.Next())
	assert.NoError(t, scan.Err())
	assert.Equal(t, 1, scan.Chunk().Num())
	assert.Equal(t, h2, scan.Chunk().Hash())
	assert.Equal(t, 456, scan.Chunk().TargetSize())

	assert.False(t, scan.Next())
	assert.NoError(t, scan.Err())
}
