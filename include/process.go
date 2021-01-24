/*
process.go defines process's operating functions and constants

files combined from the original X86 version include:
process.h
process.c
ready.c

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

// Ready function set process state to indicate ready and add to ready list, then rescheduling
func Ready(pid Pid32) error {
	if IsBadPid(pid) {
		return ErrSYSERR
	}

	prptr := &Proctab[pid]
	prptr.PrState = PrReady
	Insert(pid, ReadyList, int32(prptr.PrPrio))
	Resched()

	return OK
}

// Resume function unsuspend a process, making it ready, return its previous priority
func Resume(pid Pid32) (Pri16, error) {
	mask := Disable()   // close interrupt
	defer Restore(mask) // make sure interrupt mask is restored before return

	if IsBadPid(pid) {
		return NonePri, ErrSYSERR
	}

	prpter := &Proctab[pid]
	if prpter.PrState != PrSusp { // resume only valid for SUSPEND state
		return NonePri, ErrSYSERR
	}

	// record priority on current stack since Ready could cause a rescheduling
	prio := prpter.PrPrio
	Ready(pid)

	return prio, OK
}

// Suspend function suspend a process, placing it in hibernation(休眠)
func Suspend(pid Pid32) (Pri16, error) {
	mask := Disable()
	defer Restore(mask)

	// the null process cannot be suspended
	if IsBadPid(pid) || pid == Pid32(NULLProc) {
		return NonePri, ErrSYSERR
	}

	// only suspend a process that in current or ready
	prptr := &Proctab[pid]
	if prptr.PrState != PrCurr && prptr.PrState != PrReady {
		return NonePri, ErrSYSERR
	}

	// in ready
	if prptr.PrState == PrReady {
		GetItem(pid)           // remove it from the ready list
		prptr.PrState = PrSusp // update its state to SUSPEND
	} else { // in current
		prptr.PrState = PrSusp // update its state to SUSPEND
		Resched()              // rescheduling another process to execute
	}

	// when it resume execution, it return from the Resched, and start from here
	prio := prptr.PrPrio
	return prio, OK
}
