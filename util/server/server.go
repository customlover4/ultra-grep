package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	datareader "grepkvorum/util/dataReader"
	"grepkvorum/util/entities"
	"io"
	"net"
)

func ParseData(v net.Conn) (entities.AcceptedData, error) {
	const op = "parseData"

	defer func() {
		_ = v.Close()
	}()

	var d entities.AcceptedData

	var l int64
	if err := binary.Read(v, binary.LittleEndian, &l); err != nil {
		return d, fmt.Errorf("%s: %w", op, err)
	}
	data := make([]byte, l)
	if _, err := io.ReadFull(v, data); err != nil {
		return d, fmt.Errorf("%s: %w", op, err)
	}

	err := d.UnmarshalBytes(data)
	if err != nil {
		return d, fmt.Errorf("%s: %w", op, err)
	}
	return d, nil
}

func SendShards(repls int, fpath string, shards []datareader.Shard, conns []net.Conn) error {
	const op = "sendShards"

	connID := 0
	for i, v := range shards {
		wd := entities.WorkerData{
			FileName: fpath,
			ShardID:  int64(i),
			Start:    v.Start,
			End:      v.End,
		}
		data, err := wd.MarshalBytes()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		msg := new(bytes.Buffer)
		if err := binary.Write(msg, binary.LittleEndian, int64(len(data))); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err := binary.Write(msg, binary.LittleEndian, data); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		end := connID + repls
		for ; connID < end; connID++ {
			_, err := conns[connID].Write(msg.Bytes())
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	return nil
}
