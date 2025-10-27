package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"grepkvorum/util/entities"
	"net"
)

func Data(conn net.Conn) (entities.WorkerData, error) {
	const op = "data"

	var l int64
	if err := binary.Read(conn, binary.LittleEndian, &l); err != nil {
		return entities.WorkerData{}, fmt.Errorf("%s: %w", op, err)
	}
	data := make([]byte, l)
	if err := binary.Read(conn, binary.LittleEndian, &data); err != nil {
		return entities.WorkerData{}, fmt.Errorf("%s: %w", op, err)
	}

	var wd entities.WorkerData
	err := wd.UnmarshalBytes(data)
	if err != nil {
		return entities.WorkerData{}, fmt.Errorf("%s: %w", op, err)
	}

	return wd, nil
}

func Answer(conn net.Conn, rv *entities.ResultVector, shardID int64) error {
	const op = "answer"

	d := entities.AcceptedData{
		ShardID: shardID,
		Rv:      rv,
	}

	data, err := d.MarshalBytes()
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

	payload := msg.Bytes()
	for len(payload) > 0 {
		n, err := conn.Write(payload)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		payload = payload[n:]
	}

	return nil
}
