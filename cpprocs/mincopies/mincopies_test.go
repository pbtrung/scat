package mincopies

import (
	"errors"
	"sort"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/cpprocs"
	"secsplit/testutil"
)

func TestMinCopies(t *testing.T) {
	const min = 2

	shuffleOrig := shuffle
	defer func() {
		shuffle = shuffleOrig
	}()

	hash1 := checksum.Sum([]byte("hash1"))
	hash2 := checksum.Sum([]byte("hash2"))
	hash3 := checksum.Sum([]byte("hash3"))
	hash4 := checksum.Sum([]byte("hash4"))

	called := []string{}
	errs := map[string]error{}
	testProc := func(id string) aprocs.Proc {
		return aprocs.InplaceProcFunc(func(*ss.Chunk) error {
			called = append(called, id)
			return errs[id]
		})
	}

	copiers := []cpprocs.Copier{
		cpprocs.NewCopier("a",
			testutil.SliceLister{{Hash: hash1}},
			testProc("a"),
		),
		cpprocs.NewCopier("b",
			testutil.SliceLister{{Hash: hash1}, {Hash: hash2}},
			testProc("b"),
		),
		cpprocs.NewCopier("c",
			testutil.SliceLister{},
			testProc("c"),
		),
	}

	var mc aprocs.DynProcer
	resetMc := func() {
		var err error
		mc, err = New(min, copiers)
		assert.NoError(t, err)
	}
	resetCalled := func() {
		called = called[:0]
	}
	resetErrs := func() {
		for k := range errs {
			delete(errs, k)
		}
	}
	reset := func() {
		resetMc()
		resetCalled()
		resetErrs()
	}

	testProcsForHash := func(
		h checksum.Hash, expectedCalls []string, expectedErr error,
	) {
		c := &ss.Chunk{Hash: h}
		procs, err := mc.Procs(c)
		assert.NoError(t, err)
		chunks, err := processByAll(c, procs)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, len(chunks))
		assert.Equal(t, []*ss.Chunk{c}, chunks)
		assert.Equal(t, expectedCalls, called)
	}

	reset()
	testProcsForHash(hash1, []string{}, nil)

	reset()
	shuffle = byId
	testProcsForHash(hash2, []string{"a"}, nil)
	resetCalled()
	testProcsForHash(hash2, []string{}, nil)

	reset()
	shuffle = reverse
	testProcsForHash(hash2, []string{"c"}, nil)
	resetCalled()
	testProcsForHash(hash2, []string{}, nil)

	reset()
	shuffle = byId
	testProcsForHash(hash3, []string{"a", "b"}, nil)
	resetCalled()
	testProcsForHash(hash3, []string{}, nil)

	reset()
	shuffle = reverse
	testProcsForHash(hash3, []string{"c", "b"}, nil)
	resetCalled()
	testProcsForHash(hash3, []string{}, nil)

	// Failover: OK
	reset()
	shuffle = byId
	someErr := errors.New("some err")
	errs["a"] = someErr
	testProcsForHash(hash3, []string{"a", "c", "b"}, nil)
	resetCalled()
	testProcsForHash(hash3, []string{}, nil)
	resetCalled()
	testProcsForHash(hash4, []string{"b", "c"}, nil)

	// Failover: all KO
	reset()
	shuffle = byId
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	errs["a"] = err1
	errs["c"] = err2
	testProcsForHash(hash3, []string{"a", "c", "b"}, err2)
	resetCalled()
	testProcsForHash(hash3, []string{}, nil)
}

func TestFinish(t *testing.T) {
	copiers := []cpprocs.Copier{
		cpprocs.NewCopier(
			nil,
			testutil.SliceLister{},
			testutil.FinishErrProc{Err: nil},
		),
	}
	mc, err := New(2, copiers)
	assert.NoError(t, err)
	err = mc.Finish()
	assert.NoError(t, err)
}

func TestFinishError(t *testing.T) {
	someErr := errors.New("some err")
	copiers := []cpprocs.Copier{
		cpprocs.NewCopier(
			nil,
			testutil.SliceLister{},
			testutil.FinishErrProc{Err: someErr},
		),
	}
	mc, err := New(2, copiers)
	assert.NoError(t, err)
	err = mc.Finish()
	assert.Equal(t, someErr, err)
}

func processByAll(c *ss.Chunk, procs []aprocs.Proc) (
	all []*ss.Chunk, err error,
) {
	for _, proc := range procs {
		chunks, e := testutil.ReadChunks(proc.Process(c))
		if e != nil {
			if err == nil {
				err = e
			}
			continue
		}
		all = append(all, chunks...)
	}
	return
}

func TestShuffle(t *testing.T) {
	s := []cpprocs.Copier{
		cpprocs.NewCopier("a", nil, nil),
		cpprocs.NewCopier("b", nil, nil),
		cpprocs.NewCopier("c", nil, nil),
	}
	ids := ids(shuffle(s))
	sort.Strings(ids)
	assert.Equal(t, []string{"a", "b", "c"}, ids)
}

func byId(s []cpprocs.Copier) (res []cpprocs.Copier) {
	res = make([]cpprocs.Copier, len(s))
	copy(res, s)
	sortable := func(i int) string {
		return res[i].Id().(string)
	}
	sort.Slice(res, func(i, j int) bool {
		return sortable(i) < sortable(j)
	})
	return
}

func reverse(s []cpprocs.Copier) (res []cpprocs.Copier) {
	s = byId(s)
	n := len(s)
	res = make([]cpprocs.Copier, n)
	for i := 0; i < n; i++ {
		res[i] = s[n-i-1]
	}
	return
}

func TestReverseTest(t *testing.T) {
	s := []cpprocs.Copier{
		cpprocs.NewCopier("a", nil, nil),
		cpprocs.NewCopier("b", nil, nil),
		cpprocs.NewCopier("c", nil, nil),
	}
	assert.Equal(t, []string{"c", "b", "a"}, ids(reverse(s)))
}

func ids(s []cpprocs.Copier) (ids []string) {
	for _, c := range s {
		ids = append(ids, c.Id().(string))
	}
	return
}
