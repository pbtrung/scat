package aprocs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/aprocs"
)

func TestGroup(t *testing.T) {
	g := aprocs.NewGroup(2)

	chunks, err := readChunks(g.Process(scat.NewChunk(1, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))

	chunks, err = readChunks(g.Process(scat.NewChunk(2, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))

	chunks, err = readChunks(g.Process(scat.NewChunk(0, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))

	chunk := chunks[0]
	assert.Equal(t, 0, chunk.Num())
	grp := chunk.Meta().Get("group").([]scat.Chunk)
	assert.Equal(t, 2, len(grp))
	assert.Equal(t, 0, grp[0].Num())
	assert.Equal(t, 1, grp[1].Num())

	chunks, err = readChunks(g.Process(scat.NewChunk(3, nil)))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))

	chunk = chunks[0]
	assert.Equal(t, 1, chunk.Num())
	grp = chunk.Meta().Get("group").([]scat.Chunk)
	assert.Equal(t, 2, len(grp))
	assert.Equal(t, 2, grp[0].Num())
	assert.Equal(t, 3, grp[1].Num())
}

func TestGroupMinSize(t *testing.T) {
	assert.Panics(t, func() { aprocs.NewGroup(0) })
	assert.NotPanics(t, func() { aprocs.NewGroup(1) })
}

func TestGroupChunkNums(t *testing.T) {
	testChunkNums(t, aprocs.NewGroup(2), 6)
}

func TestGroupFinish(t *testing.T) {
	g := aprocs.NewGroup(2)

	// 0 ok
	// 1 missing
	_, err := readChunks(g.Process(scat.NewChunk(0, nil)))
	assert.NoError(t, err)
	err = g.Finish()
	assert.Equal(t, aprocs.ErrShort, err)

	// idempotence
	err = g.Finish()
	assert.Equal(t, aprocs.ErrShort, err)

	// 0 ok
	// 1 ok
	_, err = readChunks(g.Process(scat.NewChunk(1, nil)))
	assert.NoError(t, err)
	err = g.Finish()
	assert.NoError(t, err)

	// idempotence
	err = g.Finish()
	assert.NoError(t, err)
}
