package procs_test

import (
	"testing"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/testutil"
	assert "github.com/stretchr/testify/require"
)

func TestSortFinish(t *testing.T) {
	sortp := &procs.Sort{}

	// 0 missing
	// 1 ok
	_, err := testutil.ReadChunks(sortp.Process(scat.NewChunk(1, nil)))
	assert.NoError(t, err)
	err = sortp.Finish()
	assert.Equal(t, procs.ErrShort, err)

	// idempotence
	err = sortp.Finish()
	assert.Equal(t, procs.ErrShort, err)

	// 0 ok
	// 1 ok
	_, err = testutil.ReadChunks(sortp.Process(scat.NewChunk(0, nil)))
	assert.NoError(t, err)
	err = sortp.Finish()
	assert.NoError(t, err)

	// idempotence
	err = sortp.Finish()
	assert.NoError(t, err)
}
