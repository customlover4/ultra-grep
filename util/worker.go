package util

import (
	"fmt"
	"grepkvorum/util/client"
	"io"
	"net"
	"os"
	"time"
)

func worker(a Args, lock chan struct{}, pattern, coordinator string) {
	conn, err := net.Dial("tcp", coordinator)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	conn.SetReadDeadline(time.Now().Add(time.Minute * 10))
	conn.SetWriteDeadline(time.Now().Add(time.Minute * 10))

	wd, err := client.Data(conn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	f, err := os.Open(wd.FileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	f.Seek(wd.Start, io.SeekStart)
	reader := io.LimitReader(f, wd.End-wd.Start)

	a.InStream = reader

	rv, err := Do(a, false, pattern)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	lock <- struct{}{}
	err = client.Answer(conn, rv, wd.ShardID)
	<-lock
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func workers(n int, a Args, pattern, coordinator string) {
	lock := make(chan struct{}, 4)
	for i := 0; i < n; i++ {
		go worker(a, lock, pattern, coordinator)
	}
}
