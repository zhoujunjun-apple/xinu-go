package include

import "fmt"

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

// NonePid represent the universal invalid process id
const NonePid Pid32 = -1

// NoneQid represent the universal invalid queue id
const NoneQid Qid16 = -1

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
