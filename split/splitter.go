package split

import (
	"fmt"
	"io"

	"github.com/restic/chunker"

	"scat"
)

const (
	defaultMinSize = chunker.MinSize
	defaultMaxSize = chunker.MaxSize
	minMinSize     = 512 * 1024 // chunker.chunkerBufSize
)

type splitter struct {
	chunker *chunker.Chunker
	buf     []byte
	num     int // int for use as slice index
	chunk   scat.Chunk
	err     error
}

func NewSplitter(r io.Reader) scat.ChunkIter {
	return NewSplitterSize(r, defaultMinSize, defaultMaxSize)
}

func NewSplitterSize(r io.Reader, minSize, maxSize uint) scat.ChunkIter {
	if minSize < minMinSize {
		panic(fmt.Sprintf("min size must be >= %d bytes", minMinSize))
	}
	chunker := chunker.New(r, chunker.Pol(0x3DA3358B4DC173))
	chunker.MinSize = minSize
	chunker.MaxSize = maxSize
	return &splitter{
		chunker: chunker,
		buf:     make([]byte, maxSize),
	}
}

func (s *splitter) Next() bool {
	c, err := s.chunker.Next(s.buf)
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		s.err = err
		return false
	}
	data := make([]byte, len(c.Data))
	copy(data, c.Data)
	s.chunk = scat.NewChunk(s.num, data)
	s.num++
	// Check for overflow: uint resets to 0, int resets to -minInt
	if s.num <= 0 {
		panic("overflow")
	}
	return true
}

func (s *splitter) Chunk() scat.Chunk {
	return s.chunk
}

func (s *splitter) Err() error {
	return s.err
}
