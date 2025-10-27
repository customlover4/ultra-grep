package entities

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

const (
	Delimeter = "--"
)

func (rv *ResultVector) Add(v string, idx int, justPrint bool, numered bool) {
	if justPrint {
		if numered {
			fmt.Printf("%d: %s\n", idx, v)
		} else {
			fmt.Println(v)
		}
		return
	}

	rv.mu.Lock()
	defer rv.mu.Unlock()
	if len(rv.v) == 0 { // if this is first el we will just add this via append
		rv.v = append(
			rv.v, VectorData{v, idx},
		)
		return
	}

	if _, ok := rv.alreadyHave[idx]; ok {
		return // if alredy have this el in array, we will just out from func
	}

	rv.v = append(
		rv.v, VectorData{v, idx},
	)

	// set this idx to added in our result array
	rv.alreadyHave[idx] = struct{}{}
}

func (rv *ResultVector) Len() int {
	return len(rv.v)
}

func (rv *ResultVector) AddCounter() {
	rv.counter++
}

func (rv *ResultVector) GetCounter() int64 {
	return rv.counter
}

func (rv *ResultVector) PrettyPrint(outputNumber bool) {
	b := bufio.NewWriterSize(os.Stdout, 1<<20)
	defer b.Flush()
	for i, v := range rv.v {
		if (i > 0) && (v.CustomIdx-rv.v[i-1].CustomIdx > 1) {
			b.WriteString(Delimeter)
			b.WriteByte('\n')
		}

		if outputNumber && v.Value != Delimeter {
			b.WriteString(fmt.Sprintf("%d: ", v.CustomIdx))
		}
		b.WriteString(v.Value)
		b.WriteByte('\n')
	}
}

func (rv *ResultVector) Get(i int) VectorData {
	if i >= rv.Len() {
		return VectorData{}
	} else if i < 0 {
		return VectorData{}
	}

	return rv.v[i]
}

func (rv *ResultVector) Result(outputNumber bool) []string {
	res := make([]string, 0, len(rv.v))
	for _, v := range rv.v {
		var s string
		if outputNumber && v.Value != Delimeter {
			s = fmt.Sprintf("%d: %s", v.CustomIdx, v.Value)
		} else {
			s = v.Value
		}
		res = append(res, s)
	}
	return res
}

func Cmp(a *ResultVector, b *ResultVector) bool {
	if len(a.v) != len(b.v) {
		return false
	}
	for i := 0; i < len(a.v); i++ {
		if a.v[i] != b.v[i] {
			return false
		}
	}
	return true
}

func MergeVectors(vectors ...*ResultVector) *ResultVector {
	if len(vectors) == 0 {
		return NewResultVector()
	} else if len(vectors) == 1 {
		return vectors[0]
	}

	sort.Slice(vectors, func(i, j int) bool {
		if len(vectors[i].v) == 0 {
			return false
		}
		if len(vectors[j].v) == 0 {
			return true
		}
		return vectors[i].v[0].CustomIdx < vectors[j].v[0].CustomIdx
	})

	r := NewResultVector()

	for i := 0; i < len(vectors); i++ {
		if len(vectors[i].v) == 0 {
			continue
		}

		for j := 0; j < len(vectors[i].v); j++ {
			r.Add(vectors[i].v[j].Value, vectors[i].v[j].CustomIdx, false, false)
			r.AddCounter()
		}
	}

	return r
}
