package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type stats struct {
	RSSMaxB   uint64
	VirtMaxB  uint64
	Start     uint64
	End       uint64
	ElapsedNS uint64
	UTimeNS   uint64
	STimeNS   uint64

	RChar  uint64
	WChar  uint64
	SyscR  uint64
	SyscW  uint64
	RBytes uint64
	WBytes uint64
}

func (s *stats) String() string {
	format := `RSSMaxB   %d
VirtMaxB  %d
Start     %d
End       %d
ElapsedNS %d
UTimeNS   %d
STimeNS   %d
RChar     %d
WChar     %d
SyscR     %d
SyscW     %d
RBytes    %d
WBytes    %d`

	return fmt.Sprintf(
		format,
		s.RSSMaxB,
		s.VirtMaxB,
		s.Start,
		s.End,
		s.ElapsedNS,
		s.UTimeNS,
		s.STimeNS,
		s.RChar,
		s.WChar,
		s.SyscR,
		s.SyscW,
		s.RBytes,
		s.WBytes,
	)
}

func (s *stats) HumanReadableString() string {
	format := `Resident Mem Max %s
Virtual Mem Max  %s
Elapsed          %s
User Time        %s
System Time      %s
Read Char        %s
Write Char       %s
Syscall Read     %s
Syscall Write    %s
Read Bytes       %s
Write Bytes      %s`

	return fmt.Sprintf(
		format,
		hrSize(s.RSSMaxB),
		hrSize(s.VirtMaxB),
		hrElapsed(s.ElapsedNS),
		hrElapsed(s.UTimeNS),
		hrElapsed(s.STimeNS),
		hrSize(s.RChar),
		hrSize(s.WChar),
		hrNumber(s.SyscR),
		hrNumber(s.SyscW),
		hrSize(s.RBytes),
		hrSize(s.WBytes),
	)
}

func pstat(ctx context.Context, cmdAndArgs []string) stats {
	cmd := exec.CommandContext(ctx, cmdAndArgs[0], cmdAndArgs[1:]...)
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
	var statsData stats
	start := time.Now()
	statsData.Start = uint64(start.Unix())
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
				time.Second*5)
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

		statsData.RSSMaxB = max(statsData.RSSMaxB, pidStat.Rss)
		statsData.VirtMaxB = max(statsData.VirtMaxB, pidStat.Vsize)
		statsData.UTimeNS = pidStat.Utime
		statsData.STimeNS = pidStat.Stime
		statsData.RChar = ioStat.rchar
		statsData.WChar = ioStat.wchar
		statsData.SyscR = ioStat.syscr
		statsData.SyscW = ioStat.syscw
		statsData.RBytes = ioStat.read_bytes
		statsData.WBytes = ioStat.write_bytes
	}
	end := time.Now()
	statsData.ElapsedNS = uint64(end.Sub(start))
	statsData.End = uint64(end.Unix())
	statsData.RSSMaxB *= pageSize
	statsData.UTimeNS *= clockTickMul
	statsData.STimeNS *= clockTickMul

	if loopErr != nil {
		fmt.Fprintln(os.Stderr, wrap(err))
	}

	return statsData
}

func hrSize(byteSize uint64) string {
	outNum := float64(byteSize)
	outSuffix := "B"
	showDecimals := false

	const (
		tebibyte = 1 << 40
		gibibyte = 1 << 30
		mebibyte = 1 << 20
		kibibyte = 1 << 10
	)

	switch {
	case byteSize >= tebibyte:
		outNum = float64(byteSize) / float64(tebibyte)
		outSuffix = "TiB"
		showDecimals = true

	case byteSize >= gibibyte:
		outNum = float64(byteSize) / float64(gibibyte)
		outSuffix = "GiB"
		showDecimals = true

	case byteSize >= mebibyte:
		outNum = float64(byteSize) / float64(1<<20)
		outSuffix = "MiB"
		showDecimals = true

	case byteSize >= (1 << 10):
		outNum = float64(byteSize) / float64(1<<10)
		outSuffix = "KiB"
		showDecimals = true
	}

	if showDecimals {
		return fmt.Sprintf("%.3f%s", outNum, outSuffix)
	}

	return fmt.Sprintf("%d%s", uint64(outNum), outSuffix)
}

func hrElapsed(ns uint64) string {
	var days, hours, minutes, seconds, milliseconds, microseconds,
		nanoseconds uint64

	dayNs := uint64(time.Hour * 24)
	days = ns / dayNs
	ns -= dayNs * days

	hours = ns / uint64(time.Hour)
	ns -= uint64(time.Hour) * hours

	minutes = ns / uint64(time.Minute)
	ns -= uint64(time.Minute) * minutes

	seconds = ns / uint64(time.Second)
	ns -= uint64(time.Second) * seconds

	milliseconds = ns / uint64(time.Millisecond)
	ns -= uint64(time.Millisecond) * milliseconds

	microseconds = ns / uint64(time.Microsecond)
	ns -= uint64(time.Microsecond) * microseconds

	nanoseconds = ns

	switch {
	case days != 0:
		return fmt.Sprintf(
			"%dD %dH %dm %ds",
			days,
			hours,
			minutes,
			seconds,
		)

	case hours != 0:
		return fmt.Sprintf(
			"%dH %dm %ds %dms",
			hours,
			minutes,
			seconds,
			milliseconds,
		)

	case minutes != 0:
		return fmt.Sprintf(
			"%dm %ds %dms %dμs",
			minutes,
			seconds,
			milliseconds,
			microseconds,
		)

	case seconds != 0:
		return fmt.Sprintf(
			"%ds %dms %dμs %dns",
			seconds,
			milliseconds,
			microseconds,
			nanoseconds,
		)

	case milliseconds != 0:
		return fmt.Sprintf(
			"%dms %dμs %dns",
			milliseconds,
			microseconds,
			nanoseconds,
		)

	case microseconds != 0:
		return fmt.Sprintf(
			"%dμs %dns",
			microseconds,
			nanoseconds,
		)

	default:
		return fmt.Sprintf(
			"%dns",
			nanoseconds,
		)
	}
}

func hrNumber(num uint64) string {
	outNum := float64(num)
	outSuffix := ""
	showDecimals := false

	const (
		trillion = 1_000_000_000_000
		billion  = 1_000_000_000
		million  = 1_000_000
		thousand = 1_000
		hundred  = 100
	)

	switch {
	case num >= trillion:
		outNum = float64(num) / float64(trillion)
		showDecimals = true
		outSuffix = "T"

	case num >= billion:
		outNum = float64(num) / float64(billion)
		showDecimals = true
		outSuffix = "B"

	case num >= million:
		outNum = float64(num) / float64(million)
		showDecimals = true
		outSuffix = "M"

	case num >= thousand:
		outNum = float64(num) / float64(thousand)
		showDecimals = true
		outSuffix = "K"
	}

	if showDecimals {
		return fmt.Sprintf("%.3f%s", outNum, outSuffix)
	}

	return fmt.Sprintf("%d%s", uint64(outNum), outSuffix)
}

func main() {
	fs := flag.NewFlagSet("pstat", flag.ExitOnError)
	var humanReadable bool
	fs.BoolVar(
		&humanReadable,
		"h",
		false,
		"output in human readable format",
	)
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "expected command to execute")
		os.Exit(1)
	}
	err := fs.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}
	args := fs.Args()

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "expected command to execute")
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	statsData := pstat(ctx, args[:])

	if !humanReadable {
		fmt.Printf("\n%s\n", &statsData)
		return
	}
	fmt.Printf("\n%s\n", statsData.HumanReadableString())
}
