package entities

import (
	"bytes"
	"encoding/binary"
)

type AcceptedData struct {
	ShardID int64
	Rv      *ResultVector
	Count   int
}

func (ad AcceptedData) MarshalBytes() ([]byte, error) {
	b := new(bytes.Buffer)

	if err := binary.Write(b, binary.LittleEndian, ad.ShardID); err != nil {
		return nil, err
	}
	dt, err := ad.Rv.MarshalBytes()
	if err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, int64(len(dt))); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, dt); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (ad *AcceptedData) UnmarshalBytes(data []byte) error {
	b := bytes.NewBuffer(data)

	var shardID int64
	if err := binary.Read(b, binary.LittleEndian, &shardID); err != nil {
		return err
	}
	ad.ShardID = shardID
	var rvDtL int64
	if err := binary.Read(b, binary.LittleEndian, &rvDtL); err != nil {
		return err
	}
	dt := make([]byte, rvDtL)
	if err := binary.Read(b, binary.LittleEndian, &dt); err != nil {
		return err
	}
	ad.Rv = NewResultVector()
	if err := ad.Rv.UnmarshalBytes(dt); err != nil {
		return err
	}

	return nil
}

type WorkerData struct {
	Start    int64
	End      int64
	ShardID  int64
	FileName string
}

func (wd WorkerData) MarshalBytes() ([]byte, error) {
	b := new(bytes.Buffer)

	var l int64 = int64(len(wd.FileName))
	if err := binary.Write(b, binary.LittleEndian, l); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, []byte(wd.FileName)); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, wd.Start); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, wd.End); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, wd.ShardID); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (wd *WorkerData) UnmarshalBytes(data []byte) error {
	buff := bytes.NewBuffer(data)

	var l int64
	if err := binary.Read(buff, binary.LittleEndian, &l); err != nil {
		return err
	}
	fn := make([]byte, l)
	if err := binary.Read(buff, binary.LittleEndian, &fn); err != nil {
		return err
	}
	var start int64
	if err := binary.Read(buff, binary.LittleEndian, &start); err != nil {
		return err
	}
	var end int64
	if err := binary.Read(buff, binary.LittleEndian, &end); err != nil {
		return err
	}
	var shardID int64
	if err := binary.Read(buff, binary.LittleEndian, &shardID); err != nil {
		return err
	}

	wd.FileName = string(fn)
	wd.Start = start
	wd.End = end
	wd.ShardID = shardID

	return nil
}
