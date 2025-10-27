package util

import (
	"errors"
	"fmt"
	datareader "grepkvorum/util/dataReader"
	"grepkvorum/util/entities"
	"grepkvorum/util/server"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

func CoordinatorProcess(
	n, repls, port int, fpath, pattern string, a Args,
	shards []datareader.Shard,
) error {
	const op = "main process"

	// Set zero's for context flags because we can't process it in shards
	a.AfterMatch = 0
	a.BeforeMatch = 0
	a.AroundMatch = 0
	a.NumberForStringsFlag = false

	connN := n * repls
	coordinatorAddr := ":" + strconv.Itoa(port)

	coordinator, err := net.Listen("tcp", coordinatorAddr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		_ = coordinator.Close()
	}()

	workers(connN, a, pattern, coordinatorAddr)

	conns, err := getConns(coordinator, connN)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for _, v := range conns {
		defer func() {
			_ = v.Close()
		}()
	}

	err = server.SendShards(repls, fpath, shards, conns)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	accepted, err := accept(repls, shards, conns)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	output(accepted, a)

	return nil
}

func output(accepted []entities.AcceptedData, a Args) {
	sort.Slice(accepted, func(i int, j int) bool {
		return accepted[i].ShardID < accepted[j].ShardID
	})

	prevShardLastIDX := -1
	counter := 0
	for i, v := range accepted {
		if v.Rv.Len() == 0 {
			continue
		}

		if a.CountStrings {
			counter += int(v.Rv.GetCounter())
			continue
		}

		if i != 0 && prevShardLastIDX != -1 && v.Rv.Get(0).CustomIdx-prevShardLastIDX > 1 {
			fmt.Println(entities.Delimeter)
		}
		v.Rv.PrettyPrint(a.NumberForStringsFlag)
		prevShardLastIDX = v.Rv.Get(v.Rv.Len() - 1).CustomIdx
	}

	if a.CountStrings {
		fmt.Println("Total lines:", counter)
	}
}

func accept(repls int, shards []datareader.Shard, conns []net.Conn) ([]entities.AcceptedData, error) {
	const op = "accept"

	qourum := repls/2 + 1
	accepted := make([]entities.AcceptedData, len(shards))

	mu := new(sync.Mutex)

	wg := new(sync.WaitGroup)
	wg.Add(len(conns))

	err := make(chan error, len(conns))
	for _, v := range conns {
		go func(wg *sync.WaitGroup, mu *sync.Mutex, e chan error) {
			defer wg.Done()

			d, err := server.ParseData(v)
			if err != nil {
				e <- err
				return
			}

			mu.Lock()
			defer mu.Unlock()
			switch accepted[d.ShardID].Count {
			case 0:
				accepted[d.ShardID] = d
				accepted[d.ShardID].Count++
				return
			case qourum:
				return
			}

			if entities.Cmp(accepted[d.ShardID].Rv, d.Rv) {
				accepted[d.ShardID].Count++
			}
		}(wg, mu, err)
	}
	wg.Wait()

	close(err)
	errs := make([]error, 0, len(conns))
	for v := range err {
		errs = append(errs, v)
	}
	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return accepted, nil
}

func getConns(coordinator net.Listener, n int) ([]net.Conn, error) {
	const op = "getConns"
	conns := make([]net.Conn, 0, n)
	for {
		select {
		case <-time.After(5 * time.Second):
			return nil, fmt.Errorf("%s: %w", op, errors.New("timeout"))
		default:
			if len(conns) == n {
				return conns, nil
			}
			conn, err := coordinator.Accept()
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			conns = append(conns, conn)
		}
	}
}
