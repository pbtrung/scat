package procs

import "scat"

type Filter struct {
	Proc
	Filter func(Res) Res
}

var _ Proc = Filter{}

func (p Filter) Process(c *scat.Chunk) <-chan Res {
	ch := p.Proc.Process(c)
	out := make(chan Res)
	go func() {
		defer close(out)
		for res := range ch {
			out <- p.Filter(res)
		}
	}()
	return out
}
