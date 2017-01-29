package procs

import (
	"io"
)

func NewJoin(w io.Writer) Proc {
	return NewBacklog(1, Chain{NewSort(), NewWriterTo(w)})
}
