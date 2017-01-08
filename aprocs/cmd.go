package aprocs

import (
	"bytes"
	"os/exec"

	"scat"
)

type CmdInFunc func(scat.Chunk) (*exec.Cmd, error)

var _ Proc = CmdInFunc(nil)

func (fn CmdInFunc) Process(c scat.Chunk) <-chan Res {
	return InplaceFunc(fn.process).Process(c)
}

func (fn CmdInFunc) process(c scat.Chunk) (err error) {
	cmd, err := fn(c)
	if err != nil {
		return
	}
	cmd.Stdin = bytes.NewReader(c.Data())
	return cmd.Run()
}

func (fn CmdInFunc) Finish() error {
	return nil
}

type CmdOutFunc func(scat.Chunk) (*exec.Cmd, error)

var _ Proc = CmdOutFunc(nil)

func (fn CmdOutFunc) Process(c scat.Chunk) <-chan Res {
	return ChunkFunc(fn.process).Process(c)
}

func (fn CmdOutFunc) process(c scat.Chunk) (new scat.Chunk, err error) {
	cmd, err := fn(c)
	if err != nil {
		return
	}
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	err = cmd.Run()
	new = c.WithData(buf.Bytes())
	return
}

func (fn CmdOutFunc) Finish() error {
	return nil
}
