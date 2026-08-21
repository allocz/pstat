package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

func wrap(err error) error {
	if err == nil {
		return nil
	}

	skip := 1
	pc, _, line, ok := runtime.Caller(skip)
	if !ok {
		return fmt.Errorf("wrap: error calling runtime.Caller")
	}

	f := runtime.FuncForPC(pc)

	return fmt.Errorf("%s:%d: %w", f.Name(), line, err)
}

func WrapMessage(msgFmt string, args ...any) error {
	skip := 1
	pc, _, line, ok := runtime.Caller(skip)
	if !ok {
		return fmt.Errorf("wrap: error calling runtime.Caller")
	}

	f := runtime.FuncForPC(pc)

	return fmt.Errorf("%s:%d: %w", f.Name(), line,
		fmt.Errorf(msgFmt, args...))
}

var _ = procPidStatGen()

func procPidStatGen() string {
	const (
		typeU64 = "uint64"
		typeStr = "string"
	)
	genTable := []struct {
		name  string
		ftype string
	}{
		// reference: https://www.man7.org/linux/man-pages//man5/proc_pid_stat.5.html
		{"Pid", typeU64},
		{"Comm", typeStr},
		{"State", typeStr},
		{"Ppid", typeU64},
		{"Pgrp", typeU64},
		{"Session", typeU64},
		{"TtyNr", typeU64},
		{"Tpgid", typeU64},
		{"Flags", typeU64},
		{"Minflt", typeU64},
		{"Cminflt", typeU64},
		{"Majflt", typeU64},
		{"Cmajflt", typeU64},
		{"Utime", typeU64},
		{"Stime", typeU64},
		{"Cutime", typeU64},
		{"Cstime", typeU64},
		{"Priority", typeU64},
		{"Nice", typeU64},
		{"NumThreads", typeU64},
		{"Itrealvalue", typeU64},
		{"Starttime", typeU64},
		{"Vsize", typeU64},
		{"Rss", typeU64},
		{"Rsslim", typeU64},
		{"Startcode", typeU64},
		{"Endcode", typeU64},
		{"Startstack", typeU64},
		{"Kstkesp", typeU64},
		{"Kstkeip", typeU64},
		{"Signal", typeU64},
		{"Blocked", typeU64},
		{"Sigignore", typeU64},
		{"Sigcatch", typeU64},
		{"Wchan", typeU64},
		{"Nswap", typeU64},
		{"Cnswap", typeU64},
		{"ExitSignal", typeU64},
		{"Processor", typeU64},
		{"RtPriority", typeU64},
		{"Policy", typeU64},
		{"DelayacctBlkioTicks", typeU64},
		{"GuestTime", typeU64},
		{"CguestTime", typeU64},
		{"StartData", typeU64},
		{"EndData", typeU64},
		{"StartBrk", typeU64},
		{"ArgStart", typeU64},
		{"ArgEnd", typeU64},
		{"EnvStart", typeU64},
		{"EnvEnd", typeU64},
		{"ExitCode", typeU64},
	}
	var buf bytes.Buffer
	// gen type
	buf.WriteString("type procPidStat struct {")
	for _, v := range genTable {
		buf.WriteString("\n\t")
		buf.WriteString(v.name)
		buf.WriteString(" ")
		buf.WriteString(v.ftype)
	}
	buf.WriteString("\n}")
	// gen parser
	buf.WriteString(
		"\n\nfunc (p *procPidStat) Parse(items [][]byte) error {",
	)
	buf.WriteString("\n\tvar err error")
	fmt.Fprintf(&buf, "\n\tif len(items) != %d {", len(genTable))
	buf.WriteString("\n\t\treturn WrapMessage(\"bad stat items length\")")
	buf.WriteString("\n\t}")
	for i, v := range genTable {
		switch v.ftype {
		case typeU64:
			fmt.Fprintf(
				&buf,
				"\n\tp.%s, err = strconv.ParseUint(string(items[%d]), 10, 64)"+
					"\n\tif err != nil {"+
					"\n\t\treturn WrapMessage(\"%s: %%w\", err)"+
					"\n\t}",
				v.name,
				i,
				v.name,
			)
		case typeStr:
			fmt.Fprintf(
				&buf,
				"\n\tp.%s = string(items[%d])",
				v.name,
				i,
			)
		}
	}
	buf.WriteString("\n\treturn nil\n}")
	return buf.String()
}

type procPidStat struct {
	Pid                 uint64
	Comm                string
	State               string
	Ppid                uint64
	Pgrp                uint64
	Session             uint64
	TtyNr               uint64
	Tpgid               uint64
	Flags               uint64
	Minflt              uint64
	Cminflt             uint64
	Majflt              uint64
	Cmajflt             uint64
	Utime               uint64
	Stime               uint64
	Cutime              uint64
	Cstime              uint64
	Priority            uint64
	Nice                uint64
	NumThreads          uint64
	Itrealvalue         uint64
	Starttime           uint64
	Vsize               uint64
	Rss                 uint64
	Rsslim              uint64
	Startcode           uint64
	Endcode             uint64
	Startstack          uint64
	Kstkesp             uint64
	Kstkeip             uint64
	Signal              uint64
	Blocked             uint64
	Sigignore           uint64
	Sigcatch            uint64
	Wchan               uint64
	Nswap               uint64
	Cnswap              uint64
	ExitSignal          uint64
	Processor           uint64
	RtPriority          uint64
	Policy              uint64
	DelayacctBlkioTicks uint64
	GuestTime           uint64
	CguestTime          uint64
	StartData           uint64
	EndData             uint64
	StartBrk            uint64
	ArgStart            uint64
	ArgEnd              uint64
	EnvStart            uint64
	EnvEnd              uint64
	ExitCode            uint64
}

func (p *procPidStat) Parse(items [][]byte) error {
	var err error
	if len(items) != 52 {
		return WrapMessage("bad stat items length")
	}
	p.Pid, err = strconv.ParseUint(string(items[0]), 10, 64)
	if err != nil {
		return WrapMessage("Pid: %w", err)
	}
	p.Comm = string(items[1])
	p.State = string(items[2])
	p.Ppid, err = strconv.ParseUint(string(items[3]), 10, 64)
	if err != nil {
		return WrapMessage("Ppid: %w", err)
	}
	p.Pgrp, err = strconv.ParseUint(string(items[4]), 10, 64)
	if err != nil {
		return WrapMessage("Pgrp: %w", err)
	}
	p.Session, err = strconv.ParseUint(string(items[5]), 10, 64)
	if err != nil {
		return WrapMessage("Session: %w", err)
	}
	p.TtyNr, err = strconv.ParseUint(string(items[6]), 10, 64)
	if err != nil {
		return WrapMessage("TtyNr: %w", err)
	}
	p.Tpgid, err = strconv.ParseUint(string(items[7]), 10, 64)
	if err != nil {
		return WrapMessage("Tpgid: %w", err)
	}
	p.Flags, err = strconv.ParseUint(string(items[8]), 10, 64)
	if err != nil {
		return WrapMessage("Flags: %w", err)
	}
	p.Minflt, err = strconv.ParseUint(string(items[9]), 10, 64)
	if err != nil {
		return WrapMessage("Minflt: %w", err)
	}
	p.Cminflt, err = strconv.ParseUint(string(items[10]), 10, 64)
	if err != nil {
		return WrapMessage("Cminflt: %w", err)
	}
	p.Majflt, err = strconv.ParseUint(string(items[11]), 10, 64)
	if err != nil {
		return WrapMessage("Majflt: %w", err)
	}
	p.Cmajflt, err = strconv.ParseUint(string(items[12]), 10, 64)
	if err != nil {
		return WrapMessage("Cmajflt: %w", err)
	}
	p.Utime, err = strconv.ParseUint(string(items[13]), 10, 64)
	if err != nil {
		return WrapMessage("Utime: %w", err)
	}
	p.Stime, err = strconv.ParseUint(string(items[14]), 10, 64)
	if err != nil {
		return WrapMessage("Stime: %w", err)
	}
	p.Cutime, err = strconv.ParseUint(string(items[15]), 10, 64)
	if err != nil {
		return WrapMessage("Cutime: %w", err)
	}
	p.Cstime, err = strconv.ParseUint(string(items[16]), 10, 64)
	if err != nil {
		return WrapMessage("Cstime: %w", err)
	}
	p.Priority, err = strconv.ParseUint(string(items[17]), 10, 64)
	if err != nil {
		return WrapMessage("Priority: %w", err)
	}
	p.Nice, err = strconv.ParseUint(string(items[18]), 10, 64)
	if err != nil {
		return WrapMessage("Nice: %w", err)
	}
	p.NumThreads, err = strconv.ParseUint(string(items[19]), 10, 64)
	if err != nil {
		return WrapMessage("NumThreads: %w", err)
	}
	p.Itrealvalue, err = strconv.ParseUint(string(items[20]), 10, 64)
	if err != nil {
		return WrapMessage("Itrealvalue: %w", err)
	}
	p.Starttime, err = strconv.ParseUint(string(items[21]), 10, 64)
	if err != nil {
		return WrapMessage("Starttime: %w", err)
	}
	p.Vsize, err = strconv.ParseUint(string(items[22]), 10, 64)
	if err != nil {
		return WrapMessage("Vsize: %w", err)
	}
	p.Rss, err = strconv.ParseUint(string(items[23]), 10, 64)
	if err != nil {
		return WrapMessage("Rss: %w", err)
	}
	p.Rsslim, err = strconv.ParseUint(string(items[24]), 10, 64)
	if err != nil {
		return WrapMessage("Rsslim: %w", err)
	}
	p.Startcode, err = strconv.ParseUint(string(items[25]), 10, 64)
	if err != nil {
		return WrapMessage("Startcode: %w", err)
	}
	p.Endcode, err = strconv.ParseUint(string(items[26]), 10, 64)
	if err != nil {
		return WrapMessage("Endcode: %w", err)
	}
	p.Startstack, err = strconv.ParseUint(string(items[27]), 10, 64)
	if err != nil {
		return WrapMessage("Startstack: %w", err)
	}
	p.Kstkesp, err = strconv.ParseUint(string(items[28]), 10, 64)
	if err != nil {
		return WrapMessage("Kstkesp: %w", err)
	}
	p.Kstkeip, err = strconv.ParseUint(string(items[29]), 10, 64)
	if err != nil {
		return WrapMessage("Kstkeip: %w", err)
	}
	p.Signal, err = strconv.ParseUint(string(items[30]), 10, 64)
	if err != nil {
		return WrapMessage("Signal: %w", err)
	}
	p.Blocked, err = strconv.ParseUint(string(items[31]), 10, 64)
	if err != nil {
		return WrapMessage("Blocked: %w", err)
	}
	p.Sigignore, err = strconv.ParseUint(string(items[32]), 10, 64)
	if err != nil {
		return WrapMessage("Sigignore: %w", err)
	}
	p.Sigcatch, err = strconv.ParseUint(string(items[33]), 10, 64)
	if err != nil {
		return WrapMessage("Sigcatch: %w", err)
	}
	p.Wchan, err = strconv.ParseUint(string(items[34]), 10, 64)
	if err != nil {
		return WrapMessage("Wchan: %w", err)
	}
	p.Nswap, err = strconv.ParseUint(string(items[35]), 10, 64)
	if err != nil {
		return WrapMessage("Nswap: %w", err)
	}
	p.Cnswap, err = strconv.ParseUint(string(items[36]), 10, 64)
	if err != nil {
		return WrapMessage("Cnswap: %w", err)
	}
	p.ExitSignal, err = strconv.ParseUint(string(items[37]), 10, 64)
	if err != nil {
		return WrapMessage("ExitSignal: %w", err)
	}
	p.Processor, err = strconv.ParseUint(string(items[38]), 10, 64)
	if err != nil {
		return WrapMessage("Processor: %w", err)
	}
	p.RtPriority, err = strconv.ParseUint(string(items[39]), 10, 64)
	if err != nil {
		return WrapMessage("RtPriority: %w", err)
	}
	p.Policy, err = strconv.ParseUint(string(items[40]), 10, 64)
	if err != nil {
		return WrapMessage("Policy: %w", err)
	}
	p.DelayacctBlkioTicks, err = strconv.ParseUint(string(items[41]), 10, 64)
	if err != nil {
		return WrapMessage("DelayacctBlkioTicks: %w", err)
	}
	p.GuestTime, err = strconv.ParseUint(string(items[42]), 10, 64)
	if err != nil {
		return WrapMessage("GuestTime: %w", err)
	}
	p.CguestTime, err = strconv.ParseUint(string(items[43]), 10, 64)
	if err != nil {
		return WrapMessage("CguestTime: %w", err)
	}
	p.StartData, err = strconv.ParseUint(string(items[44]), 10, 64)
	if err != nil {
		return WrapMessage("StartData: %w", err)
	}
	p.EndData, err = strconv.ParseUint(string(items[45]), 10, 64)
	if err != nil {
		return WrapMessage("EndData: %w", err)
	}
	p.StartBrk, err = strconv.ParseUint(string(items[46]), 10, 64)
	if err != nil {
		return WrapMessage("StartBrk: %w", err)
	}
	p.ArgStart, err = strconv.ParseUint(string(items[47]), 10, 64)
	if err != nil {
		return WrapMessage("ArgStart: %w", err)
	}
	p.ArgEnd, err = strconv.ParseUint(string(items[48]), 10, 64)
	if err != nil {
		return WrapMessage("ArgEnd: %w", err)
	}
	p.EnvStart, err = strconv.ParseUint(string(items[49]), 10, 64)
	if err != nil {
		return WrapMessage("EnvStart: %w", err)
	}
	p.EnvEnd, err = strconv.ParseUint(string(items[50]), 10, 64)
	if err != nil {
		return WrapMessage("EnvEnd: %w", err)
	}
	p.ExitCode, err = strconv.ParseUint(string(items[51]), 10, 64)
	if err != nil {
		return WrapMessage("ExitCode: %w", err)
	}
	return nil
}

func runCmd(cmd string) ([]byte, error) {
	var so bytes.Buffer
	var stderr bytes.Buffer
	c := exec.Command("bash", "-c", cmd)
	c.Stdout = &so
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	}
	return so.Bytes(), nil
}

func (p *procPidStat) Update(pid int) error {
	b, err := runCmd(fmt.Sprintf("cat /proc/%d/stat", pid))
	if err != nil {
		return err
	}
	b = bytes.Trim(b, " ")
	b = bytes.Trim(b, "\n")
	items := bytes.Split(b, []byte(" "))
	var p2 procPidStat
	err = p2.Parse(items)
	if err != nil {
		return wrap(err)
	}
	*p = p2
	return nil
}

type procPidIo struct {
	rchar                 uint64
	wchar                 uint64
	syscr                 uint64
	syscw                 uint64
	read_bytes            uint64
	write_bytes           uint64
	cancelled_write_bytes uint64
}

func (i *procPidIo) Update(pid int) error {
	const ioStatNRow = 7
	const ioStatNCol = 2
	var r procPidIo
	expKeys := [][]byte{
		[]byte("rchar:"),
		[]byte("wchar:"),
		[]byte("syscr:"),
		[]byte("syscw:"),
		[]byte("read_bytes:"),
		[]byte("write_bytes:"),
		[]byte("cancelled_write_bytes:"),
	}
	o, cerr := runCmd(fmt.Sprintf("cat /proc/%d/io", pid))
	if cerr != nil {
		return wrap(cerr)
	}
	ob := bytes.Split(o, []byte("\n"))
	if len(ob) > 1 {
		ob = ob[:len(ob)-1]
	}
	if len(ob) != ioStatNRow {
		return WrapMessage("unexpected row count")
	}
	for i, o := range ob {
		kv := bytes.Split(o, []byte(" "))
		if len(kv) != ioStatNCol {
			return WrapMessage("unexpected column count")
		}
		if !bytes.Equal(expKeys[i], kv[0]) {
			return WrapMessage("unexpected column key")
		}
		v, err := strconv.ParseUint(string(kv[1]), 10, 64)
		if err != nil {
			return wrap(err)
		}
		switch i {
		case 0:
			r.rchar = v
		case 1:
			r.wchar = v
		case 2:
			r.syscr = v
		case 3:
			r.syscw = v
		case 4:
			r.read_bytes = v
		case 5:
			r.write_bytes = v
		case 6:
			r.cancelled_write_bytes = v
		}
	}
	*i = r
	return nil
}

func (i *procPidIo) String() string {
	return fmt.Sprintf(
		"rchar %d wchar %d syscr %d syscw %d read_bytes %d"+
			" write_bytes %d cacelled_write_bytes %d",
		i.rchar,
		i.wchar,
		i.syscr,
		i.syscw,
		i.read_bytes,
		i.write_bytes,
		i.cancelled_write_bytes,
	)
}

func pageSize() (uint64, error) {
	so, cerr := runCmd("getconf PAGE_SIZE")
	if cerr != nil {
		return 0, fmt.Errorf("pSize: %w", cerr)
	}
	so = bytes.Trim(so, "\n")
	v, err := strconv.ParseUint(string(so), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("pSize: %w", cerr)
	}
	return v, nil
}

func clockTick() (uint64, error) {
	so, cerr := runCmd("getconf CLK_TCK")
	if cerr != nil {
		return 0, fmt.Errorf("clockTick: %w", cerr)
	}
	so = bytes.Trim(so, "\n")
	tick, err := strconv.ParseUint(string(so), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("clockTick: %w", cerr)
	}
	return tick, nil
}
