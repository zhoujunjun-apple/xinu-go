/*
process.go defines process's operating functions and constants

files combined from the original X86 version include:
process.h
process.c

*/

package include

// process state constants
const (
	PrFree   uint16 = 0 // process table entry is unused
	PrCurr   uint16 = 1 // process is currently running
	PrReady  uint16 = 2 // process is on ready queue
	PrRecv   uint16 = 3 // process waiting for message
	PrSleep  uint16 = 4 // process is sleeping
	PrSusp   uint16 = 5 // process is suspended
	PrWait   uint16 = 6 // process is on semaphore queue
	PrRectim uint16 = 7 // process is receiving with timeout
)

// miscellaneous
const (
	PNMLen     uint16 = 16         // length of process's name
	NULLProc   uint16 = 0          // ID of the null process
	NDesc      uint16 = 5          // number of device descriptors a process can have open. must be odd to make procent 4N bytes
	StackMagic        = 0x0A0AAAA9 // makeer for the top of a process stack(used to help detect overflow)
)

// InitRet is the address to which process returns
var InitRet = UserRet // TODO: get function address

// process initialization constants
const (
	InitStk  uint   = 65536 // initial process stack size
	InitPrio uint16 = 20    // initial process priority
)

var (
	// PrCount is the number of currently active processes
	PrCount int32
	// CurrPid is the process id of currently executing process
	CurrPid Pid32
)

// ProcEnt struct is the entry in the process table.
// Must be multiple of 32 bits, but WHY?
type ProcEnt struct {
	PrState uint16 // process state
	PrPrio  Pri16  // process priority

	PrStkPtr  *byte  // saved stack pointer
	PrStkBase *byte  // base of run time stack
	PrStkLen  uint32 // stack length in bytes

	PrName   [PNMLen]byte // process name
	PrSem    Sid32        // semaphore on which process waits
	PrParent Pid32        // ID of the creating process

	PrMsg    Umsg32 // message sent to this process
	PrHasMsg bool   // true if msg is valid

	PrDesc [NDesc]int16 // device descriptors for process
}

// Proctab is the process table
var Proctab []ProcEnt

// IsBadPid function checks if pid is valid or not assuming interrupts are disabled
// True: pid is invalid;
// False: pid is valid;
func IsBadPid(pid Pid32) bool {
	return pid < 0 || int(pid) >= NPROC || Proctab[pid].PrState == PrFree
}
