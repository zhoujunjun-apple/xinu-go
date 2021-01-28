package include

import (
	"fmt"
	"math"
	"unsafe"
)

// Qid16 is the process queue ID
type Qid16 int16

// Pri16 is the process priority
type Pri16 int16

// Pid32 is the process ID
type Pid32 int32

// Sid32 is the semaphore ID
type Sid32 int32

// Umsg32 is the message type
type Umsg32 uint32

// IntMask is the interrupt mask type
type IntMask uint32

// Bpid32 is the buffer pool id type
type Bpid32 int32

// NonePid represent the universal invalid process id
const NonePid Pid32 = -1

// NoneQid represent the universal invalid queue id
const NoneQid Qid16 = -1

// NonePri represent the universal invalid priority value
const NonePri Pri16 = -11111

// NoneSem represent the universal invalid semaphore id
const NoneSem Sid32 = -1

// TimeoutMsg represent the TimeOut message
const TimeoutMsg Umsg32 = math.MaxUint32 - 1

// NoneMsg represent the universal invalid message content
const NoneMsg Umsg32 = math.MaxUint32

// NoneBpid represent the universal invalid buffer pool id
const NoneBpid Bpid32 = -1

// None is the null address value
const None uintptr = 0

// NonePointer represent a fake null pointer
var NonePointer unsafe.Pointer = unsafe.Pointer(None)

// QUANTUM is the time slice in milliseconds
const QUANTUM uint8 = 2

// MINSTK is the minimum stack size in bytes
const MINSTK uint32 = 400

/* Universal return constants */
var (
	// OK: system call ok
	OK error = nil
	// SYSERR : system call failed
	ErrSYSERR error = fmt.Errorf("SYSERR")
	// EOF : End-of-file (usually from read)
	ErrEOF error = fmt.Errorf("EOF")
	// TIMEOUT : system call timed out
	ErrTIMEOUT error = fmt.Errorf("TIMEOUT")
	// ErrEMPTY is the error that caused by invalid operation on empty queue
	ErrEMPTY error = fmt.Errorf("EMPTY")
)
