package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "expected command to execute")
	}
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[1], os.Args[2:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, wrap(err))
		os.Exit(1)
	}
	cmdDone := make(chan struct{})
	cmdErr := make(chan error)
	go func() {
		err := cmd.Wait()
		if err != nil {
			cmdErr <- err
		}
		close(cmdDone)
	}()

	pageSize, err := pageSize()
	if err != nil {
		panic(err)
	}
	var clockTickMul uint64
	{
		clockTicks, err := clockTick()
		if err != nil {
			panic(err)
		}
		clockTickMul = 1_000_000_000 / clockTicks
	}

	pid := cmd.Process.Pid
	var pidStat procPidStat
	var ioStat procPidIo
	var stats struct{
		RSSMaxB uint64
		VirtMaxB uint64
		Start uint64
		End uint64
		UTimeNS uint64
		STimeNS uint64

		RChar uint64
		WChar uint64
		SyscR uint64
		SyscW uint64
		RBytes uint64
		WBytes uint64
	}
	stats.Start = uint64(time.Now().Unix())
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var loopErr error
	loop:
	for range ticker.C {
		select {
		case err := <-cmdErr:
			loopErr = wrap(err)
			break loop
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second * 5)
			defer cancel()
			select {
			case <-ctx.Done():
				fmt.Fprintln(os.Stderr, "stop wait timeout")
				break loop
			case <-cmdDone:
				break loop
			}
		default:
		}

		err := pidStat.Update(pid)
		if err != nil {
			break
		}

		err = ioStat.Update(pid)
		if err != nil {
			break
		}

		stats.RSSMaxB = max(stats.RSSMaxB, pidStat.Rss)
		stats.VirtMaxB = max(stats.VirtMaxB, pidStat.Vsize)
		stats.UTimeNS = pidStat.Utime
		stats.STimeNS = pidStat.Stime
		stats.RChar = ioStat.rchar
		stats.WChar = ioStat.wchar
		stats.SyscR = ioStat.syscr
		stats.SyscW = ioStat.syscw
		stats.RBytes = ioStat.read_bytes
		stats.WBytes = ioStat.write_bytes
	}
	stats.End = uint64(time.Now().Unix())
	stats.RSSMaxB *= pageSize
	stats.UTimeNS *= clockTickMul
	stats.STimeNS *= clockTickMul

	if loopErr != nil {
		fmt.Fprintln(os.Stderr, wrap(err))
	}

	fmt.Printf("%+v\n", stats)
}
