/*
semaphore.go semaphore manager

files combined from the original X86 version include:
semaphore.h
wait.c
signal.c

*/

package include

// SFree state: semaphore table entry is available
const SFree uint8 = 0

// SUsed state: semaphore table entry is used
const SUsed uint8 = 1

// SEntry struct is the semaphore table entry
type SEntry struct {
	// SState identify whether entry is SFree or SUsed
	SState uint8

	// SCount is the count for the semaphore
	// Positive SCount means that wait() can be called SCount more times before any process blocks;
	// Negative SCount means that the semaphore queue contains SCount waiting processes.
	SCount int32

	// queue id of processes that are waiting on the semaphore
	SQueue Qid16
}

// SemTab is the semaphore table
var SemTab []SEntry

// IsBadSem function checks if semaphore id is bad
func IsBadSem(s Sid32) bool {
	return s < 0 || int(s) >= NSEM
}

// Wait function cause current process to wait on a semaphore
func Wait(sem Sid32) error {
	mask := Disable()
	defer Restore(mask)

	if IsBadSem(sem) {
		return ErrSYSERR
	}

	semptr := &SemTab[sem]
	if semptr.SState == SFree {
		return ErrSYSERR
	}

	semptr.SCount--

	if semptr.SCount < 0 {  // semaphore is not enough, current process must wait
		prptr := &Proctab[CurrPid]
		prptr.PrState = PrWait
		prptr.PrSem = sem

		// enqueue current process at the corresponding semaphore queue
		Enqueue(CurrPid, semptr.SQueue)
		// rescheduling another process to run
		Resched()
		// resume running when returned from Resched() after another process signal it
	}

	// semaphore is enough or signaled from another process
	return OK
}

// Signal function signal a semaphore, releasing a process if one is waiting
func Signal(sem Sid32) error {
	mask := Disable()
	defer Restore(mask)

	if IsBadSem(sem) {
		return ErrSYSERR
	}

	semptr := &SemTab[sem]
	if semptr.SState == SFree {
		return ErrSYSERR
	}

	oldCount := semptr.SCount
	semptr.SCount++

	if oldCount < 0 { // oldCount processes are waiting on this semaphore
		// need to release a waiting process from the semaphore waiting queue
		p, _ := Dequeue(semptr.SQueue)
		Ready(p)  // could cause a rescheduling
	}

	return OK
}