# pstat
## A tool to get stats of a single process

## Build
```
go build .
```

## Example
```
$ ./pstat sleep 5

RSSMaxB   1785856
VirtMaxB  3121152
Start     1789490877
End       1789490882
ElapsedNS 5003994386
UTimeNS   0
STimeNS   0
RChar     4148
WChar     0
SyscR     8
SyscW     0
RBytes    0
WBytes    0
```

## Example with human readable output
```
$ ./pstat -h sleep 5

Resident Mem Max 1.727MiB
Virtual Mem Max  2.977MiB
Elapsed          5s 6ms 581μs 137ns
User Time        0ns
System Time      0ns
Read Char        4.051KiB
Write Char       0B
Syscall Read     8
Syscall Write    0
Read Bytes       0B
Write Bytes      0B
```
