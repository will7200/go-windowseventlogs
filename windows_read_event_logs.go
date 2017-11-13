package windowseventlogs

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING      error = syscall.Errno(errnoERROR_IO_PENDING)
	EVENTLOG_SEQUENTIAL_READ       = 0x0001
	EVENTLOG_SEEK_READ             = 0x0002
	EVENTLOG_FORWARDS_READ         = 0x0004
	EVENTLOG_BACKWARDS_READ        = 0x0008
	MAX_BUFFER_SIZE                = 0x7ffff
	MAX_DEFAULT_BUFFER_SIZE        = 0x10000
)
var (
	modadvapi32      = windows.NewLazySystemDLL("advapi32.dll")
	procOpenEventLog = modadvapi32.NewProc("OpenEventLogW")
	procReadEventLog = modadvapi32.NewProc("ReadEventLogW")
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

func openEventLog(uncServerName *uint16, sourceName *uint16) (handle windows.Handle, err error) {
	r0, _, e1 := syscall.Syscall(procOpenEventLog.Addr(), 2, uintptr(unsafe.Pointer(uncServerName)), uintptr(unsafe.Pointer(sourceName)), 0)
	handle = windows.Handle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func readEventLog(log windows.Handle, readFlags uint32, recordOffset uint32, buffer *byte, bufferSize uint32, bytesRead *uint32, minBytesToRead *uint32) (err error) {
	r0, _, e1 := syscall.Syscall9(procReadEventLog.Addr(), 7, uintptr(log), uintptr(readFlags), uintptr(recordOffset), uintptr(unsafe.Pointer(buffer)), uintptr(bufferSize), uintptr(unsafe.Pointer(bytesRead)), uintptr(unsafe.Pointer(minBytesToRead)), 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// Log provides access to the system log.
type EventLog struct {
	Handle     windows.Handle
	bufferSize uint32
	buffer     []byte
	readFlags  uint32
	minRead    uint32
}

// OpenEventLog retrieves a handle to the specified event log.
func OpenEventLog(source string) (*EventLog, error) {
	return OpenRemoteEventLog("", source)
}

// OpenRemoteEventLog does the same as Open, but on different computer host.
func OpenRemoteEventLog(host, source string) (*EventLog, error) {
	if source == "" {
		return nil, errors.New("Specify event log source")
	}
	var s *uint16
	if host != "" {
		s = syscall.StringToUTF16Ptr(host)
	}
	h, err := openEventLog(s, syscall.StringToUTF16Ptr(source))
	if err != nil {
		return nil, err
	}
	buf := make([]byte, MAX_BUFFER_SIZE+1)
	return &EventLog{Handle: h, bufferSize: uint32(MAX_DEFAULT_BUFFER_SIZE), buffer: buf, minRead: uint32(0)}, nil
}

// SetBufferSize Sets the buffer size and reallocates the buffer
func (el *EventLog) SetBufferSize(size uint32) bool {
	if size <= uint32(MAX_BUFFER_SIZE) {
		el.bufferSize = size
		buf := make([]byte, size+1)
		el.buffer = buf
		return true
	}
	return false
}

// SetReadFlags Sets the Read Flags for Next Reading
func (el *EventLog) SetReadFlags(flags int) bool {
	el.readFlags = uint32(flags)
	return true
}

// ReadEventLog Calls Windows API to read from log
func (el *EventLog) ReadEventLog(offset uint32, read uint32) {
	readEventLog(el.Handle, el.readFlags, offset, &el.buffer[0], el.bufferSize, &read, &el.minRead)
}

// Print from local buffer
func (el *EventLog) Print(offset int, read int) {
	for i := 0; i < read; i++ {
		if uint32(offset+i) > el.bufferSize {
			break
		}
		fmt.Printf("%c", el.buffer[offset+i])
	}
}
