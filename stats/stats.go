package stats

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
)

const aliveThreshold = 1 * time.Second

type Statsd struct {
	counters   map[Id]*Counter
	countersMu sync.Mutex
	nextPos    uint32
}

type Id interface{}

func New() *Statsd {
	return &Statsd{
		counters: make(map[Id]*Counter),
	}
}

func (st *Statsd) Counter(id Id) *Counter {
	st.countersMu.Lock()
	defer st.countersMu.Unlock()
	if _, ok := st.counters[id]; !ok {
		st.counters[id] = &Counter{pos: st.nextPos, start: time.Now()}
		st.nextPos++
	}
	return st.counters[id]
}

type sortedCounter struct {
	id  Id
	cnt *Counter
}

func (st *Statsd) sortedCounters() (scnts []sortedCounter) {
	st.countersMu.Lock()
	defer st.countersMu.Unlock()
	ids := make([]Id, 0, len(st.counters))
	for id := range st.counters {
		ids = append(ids, id)
	}
	pos := func(i int) uint32 {
		return st.counters[ids[i]].pos
	}
	sort.Slice(ids, func(i, j int) bool {
		return pos(i) < pos(j)
	})
	scnts = make([]sortedCounter, len(ids))
	for i, id := range ids {
		scnts[i] = sortedCounter{id: id, cnt: st.counters[id]}
	}
	return
}

func (st *Statsd) WriteTo(w io.Writer) (written int64, err error) {
	printf := func(format string, args ...interface{}) error {
		n, err := fmt.Fprintf(w, format, args...)
		written += int64(n)
		return err
	}

	// Headers
	err = printf(
		"%15s   \t%11s\t\x1b[90m%11s\x1b[0m\n",
		"PROC", "PERF", "AVG OUT",
	)
	if err != nil {
		return
	}

	// Procs
	now := time.Now()
	for _, scnt := range st.sortedCounters() {
		cnt := scnt.cnt
		out, dur := cnt.getOut()
		outRate := rateStr(out, now.Sub(cnt.start))
		ownRate := rateStr(out, dur)
		ninst := cnt.getInst()
		info := ""
		if ninst == 0 && now.Sub(cnt.last) > aliveThreshold {
			info = fmt.Sprintf("\x1b[90m%11s\x1b[0m", "stopped")
		} else {
			info = fmt.Sprintf("%9s/s\t\x1b[90m%9s/s\x1b[0m", ownRate, outRate)
		}
		err = printf("%15s x%d\t%s\n", scnt.id, ninst, info)
		if err != nil {
			return
		}
	}

	// Goroutines
	err = printf("%15s x%d\n", "(goroutines)", runtime.NumGoroutine())
	return
}

func rateStr(n uint64, d time.Duration) string {
	rate := uint64(float64(n) / d.Seconds())
	return humanize.IBytes(rate)
}