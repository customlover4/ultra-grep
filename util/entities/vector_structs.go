package entities

import (
	"bytes"
	"encoding/binary"
	"sync"
)

type VectorData struct {
	Value     string
	CustomIdx int
}

func (vd VectorData) MarshalBytes() ([]byte, error) {
	b := new(bytes.Buffer)

	if err := binary.Write(b, binary.LittleEndian, int64(vd.CustomIdx)); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, int64(len(vd.Value))); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, []byte(vd.Value)); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (vd *VectorData) UnmarshalBytes(data []byte) error {
	b := bytes.NewBuffer(data)

	var idx int64
	if err := binary.Read(b, binary.LittleEndian, &idx); err != nil {
		return err
	}
	var l int64
	if err := binary.Read(b, binary.LittleEndian, &l); err != nil {
		return err
	}
	s := make([]byte, l)
	if err := binary.Read(b, binary.LittleEndian, &s); err != nil {
		return err
	}

	vd.CustomIdx = int(idx)
	vd.Value = string(s)

	return nil
}

type ResultVector struct {
	v       []VectorData
	counter int64

	mu *sync.Mutex

	alreadyHave map[int]struct{}
}

func (rv ResultVector) MarshalBytes() ([]byte, error) {
	b := new(bytes.Buffer)

	if err := binary.Write(b, binary.LittleEndian, rv.counter); err != nil {
		return nil, err
	}
	var l int64 = int64(len(rv.v))
	if err := binary.Write(b, binary.LittleEndian, l); err != nil {
		return nil, err
	}

	for _, v := range rv.v {
		dt, err := v.MarshalBytes()
		if err != nil {
			return nil, err
		}

		if err := binary.Write(b, binary.LittleEndian, int64(len(dt))); err != nil {
			return nil, err
		}
		if err := binary.Write(b, binary.LittleEndian, dt); err != nil {
			return nil, err
		}
	}

	mpLen := int64(len(rv.alreadyHave))
	if err := binary.Write(b, binary.LittleEndian, mpLen); err != nil {
		return nil, err
	}
	for k := range rv.alreadyHave {
		fK := int64(k)
		if err := binary.Write(b, binary.LittleEndian, fK); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

func (rv *ResultVector) UnmarshalBytes(data []byte) error {
	b := bytes.NewBuffer(data)

	rv.mu = new(sync.Mutex)

	var counter int64
	if err := binary.Read(b, binary.LittleEndian, &counter); err != nil {
		return err
	}
	rv.counter = counter

	var l int64
	if err := binary.Read(b, binary.LittleEndian, &l); err != nil {
		return err
	}
	rv.v = make([]VectorData, l)
	for i := 0; i < int(l); i++ {
		var dtL int64
		if err := binary.Read(b, binary.LittleEndian, &dtL); err != nil {
			return err
		}

		dt := make([]byte, dtL)
		if err := binary.Read(b, binary.LittleEndian, &dt); err != nil {
			return err
		}

		var vd VectorData
		err := vd.UnmarshalBytes(dt)
		if err != nil {
			return err
		}

		rv.v[i] = vd
	}

	var mpL int64
	if err := binary.Read(b, binary.LittleEndian, &mpL); err != nil {
		return err
	}
	rv.alreadyHave = make(map[int]struct{}, mpL)
	for i := 0; i < int(mpL); i++ {
		var k int64
		if err := binary.Read(b, binary.LittleEndian, &k); err != nil {
			return err
		}

		rv.alreadyHave[int(k)] = struct{}{}
	}

	return nil
}

func NewResultVector() *ResultVector {
	return &ResultVector{
		v:           make([]VectorData, 0, 10),
		alreadyHave: make(map[int]struct{}),
		mu:          new(sync.Mutex),
	}
}
