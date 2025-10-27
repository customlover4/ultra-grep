package datareader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

var (
	ErrNotFindFile        = errors.New("can't find file via this path")
	ErrWrongColumnForSort = errors.New("wrong column for sort")
)

func GetScanner(reader io.Reader) (*bufio.Scanner, error) {
	return bufio.NewScanner(reader), nil
}

func ParseLogs(sc *bufio.Scanner) ([]string, error) {
	logs := make([]string, 0, 50)
	for sc.Scan() {
		logs = append(logs, sc.Text())
	}

	return logs, nil
}

type Shard struct {
	Start int64
	End   int64
}

func SplitFileByLines(filename string, n int) ([]Shard, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()

	if n <= 0 {
		return nil, os.ErrInvalid
	}
	if n == 1 {
		return []Shard{{Start: 0, End: fileSize}}, nil
	}

	chunkSize := fileSize / int64(n)
	shards := make([]Shard, 0, n)

	for i := 0; i < n; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize
		if i == n-1 {
			end = fileSize
		}

		if i > 0 {
			_, err := file.Seek(start, io.SeekStart)
			if err != nil {
				return nil, err
			}

			buf := make([]byte, 1)
			for {
				_, err := file.Read(buf)
				if err == io.EOF {
					start = fileSize
					break
				}
				if err != nil {
					return nil, err
				}
				if buf[0] == '\n' {
					start, err = file.Seek(0, io.SeekCurrent)
					if err != nil {
						return nil, err
					}
					break
				}
			}
		}

		if i < n-1 {
			_, err := file.Seek(end, io.SeekStart)
			if err != nil {
				return nil, err
			}

			reader := bufio.NewReader(file)
			for {
				b, err := reader.ReadByte()
				if err == io.EOF {
					end = fileSize
					break
				}
				if err != nil {
					return nil, err
				}
				end++
				if b == '\n' {
					break
				}
			}
		}

		shards = append(shards, Shard{Start: start, End: end})
	}

	return shards, nil
}

func MoveStdinToFile() string {
	sc := bufio.NewScanner(os.Stdin)

	fName := fmt.Sprintf(
		"tmpFileOsStdin_%d", rand.Intn(int(time.Now().UnixMicro())),
	)
	f, err := os.Create(fName)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = f.Close()
	}()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	for sc.Scan() {
		t := sc.Text()
		_, err = writer.WriteString(t)
		if err != nil {
			_ = writer.Flush()
			panic(err)
		}
		_, err = writer.WriteRune('\n')
		if err != nil {
			_ = writer.Flush()
			panic(err)
		}
	}

	return fName
}
