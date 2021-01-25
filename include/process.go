/*
process.go defines process's operating functions and constants

files combined from the original X86 version include:
process.h
process.c
ready.c
resume.c
suspend.c
create.c
kill.c
getpid.c
getprio.c
chprio.c

*/

package include

// process state constants
const (
	PrFree    uint16 = 0 // process table entry is unused
	PrCurr    uint16 = 1 // process is currently running
	PrReady   uint16 = 2 // process is on ready queue
	PrRecv    uint16 = 3 // process waiting for message
	PrSleep   uint16 = 4 // process is sleeping
	PrSusp    uint16 = 5 // process is suspended
	PrWait    uint16 = 6 // process is on semaphore queue
	PrRecTime uint16 = 7 // process is receiving with timeout
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

	PrStkPtr  *uint32 // saved stack pointer
	PrStkBase *uint32 // base of run time stack
	PrStkLen  uint32  // stack length in bytes

	PrName   [PNMLen]byte // process name
	PrSem    Sid32        // semaphore on which process waits
	PrParent Pid32        // ID of the creating process

	PrMsg    Umsg32 // message sent to this process
	PrHasMsg bool   // true if msg is valid

	PrDesc [NDesc]int16 // device descriptors for process
}

// Proctab is the process table
var Proctab []ProcEnt

// nextpid represet the position in table to try or one beyond end of table
// used by NewPid(). Initialization value is 1 because process #0 is the null process
var nextpid Pid32 = 1

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

// GetPid function return the current process's id
func GetPid() Pid32 {
	return CurrPid
}

// GetPrio function return the current process's priority
func GetPrio(pid Pid32) (Pri16, error) {
	mask := Disable()
	defer Restore(mask)

	if IsBadPid(pid) {
		return NonePri, ErrSYSERR
	}

	prio := Proctab[pid].PrPrio

	return prio, OK
}

// ChPrio function change the scheduling priority of a process(pid) to np, returning the old priority
func ChPrio(pid Pid32, np Pri16) (Pri16, error) {
	mask := Disable()
	defer Restore(mask)

	if IsBadPid(pid) {
		return NonePri, ErrSYSERR
	}

	prptr := &Proctab[pid]
	op := prptr.PrPrio
	prptr.PrPrio = np

	return op, OK
}

// Create function create a process to start running a function on X86
// The basic idea is that Create() builds an image of the process as if
// it had been stopped while running. Specifically, Create() forms a saved
// environment on the process's stack as if the specified function had been called.
// funcAddr: address of the function at which the process should start execution;
// ssize: stack size in words;
// priority: process priority > 0;
// name: name for debugging;
// nargs: number of args that follow. (I think those args are used by function at funcAddr)
// When this function is called, the stack(NOT the one allocated for new process)
// as follows (according to the C calling convention):
//
// | (stack bottom) |
// | arg #nargs     | -----
// | .....          |   |  those args will be copy from
// | arg #2         |   |  this stack to the new process's stack
// | arg #1         | -----
// | nargs          |
// | name           |
// | priority       |
// | ssize          |
// | funcAddr       |
// | ret addr       | <- pushed by call Create()
//
// When the new process's stack is formed, the stack look as follows:
//
// | (stack bottom) |
// | STACKMAGIC     |
// | arg #nargs     |-----
// | .....          |  | those args copied from the previous stack
// | arg #2         |  |
// | arg #1         |-----
// | INITRET        | <- called when new process(started at funcAddr) exit. =UserRet()
// | funcAddr       |
// | ebp            |
// | 0x0020         |
// | 0(eax)         |
// | 0(ecx)         |
// | 0(edx)         |
// | 0(ebx)         |
// | 0(esp)         |
// | ebp            |
// | 0(esi)         |
// | 0(edi)         |
// |                | <- pointed by PrStkPtr field in new process struct
//
func Create(funcAddr *byte, ssize uint32, priority Pri16, name [PNMLen]byte, nargs uint32) (Pid32, error) {
	mask := Disable()
	defer Restore(mask)

	if ssize < MINSTK {
		ssize = MINSTK
	}

	ssize = uint32(RoundMB(int32(ssize)))

	// allocate pid and stack memory for new process
	pid, pidErr := NewPid()
	saddr, saddrErr := GetStk(ssize)

	if priority < 1 || pidErr != OK || saddrErr != OK {
		return NonePid, ErrSYSERR
	}

	// pid is allocated, stack memory is allocated

	PrCount++
	prptr := &Proctab[pid]

	// initialize process table entry for new process pid
	prptr.PrState = PrSusp
	prptr.PrPrio = priority
	prptr.PrStkBase = saddr
	prptr.PrStkLen = ssize
	prptr.PrName = name
	prptr.PrSem = -1
	prptr.PrParent = GetPid()
	prptr.PrHasMsg = false

	prptr.PrDesc[0] = CONSOLE // stdin
	prptr.PrDesc[1] = CONSOLE // stdout
	prptr.PrDesc[2] = CONSOLE // stderr

	// initialize stack as if the new process was called
	*saddr = StackMagic

	// push arguments
	// TODO: copy args from current stack to the new process's stack
	// TODO: push on return address: INITRET

	// the following entries on the stack must match what  ctxsw
	// expects a saved process state to contain: ret address,
	// ebp, interrupt mask, flags, registers, and an old SP
	// TODO: how to operate pointer on golang!!

	return pid, OK
}

// NewPid function returns a new free process ID
func NewPid() (Pid32, error) {
	var pid Pid32

	for i := 0; i < NPROC; i++ {
		nextpid %= Pid32(NPROC)
		if Proctab[nextpid].PrState == PrFree {
			pid = nextpid
			nextpid++
			return pid, OK
		}
		nextpid++
	}

	return NonePid, ErrSYSERR
}

// Kill function kill a process and remove it from the system
func Kill(pid Pid32) error {
	// TODO: implement it
	return OK
}
