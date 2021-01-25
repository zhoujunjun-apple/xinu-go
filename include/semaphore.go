/*
semaphore.go semaphore manager

files combined from the original X86 version include:
semaphore.h
wait.c
signal.c
semdelete.c
semreset.c
semcreate.c

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

// NextSem is the next semaphore index to try to allocate, used by NewSem()
var NextSem Sid32 = 0

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

// NewSem function allocate an unused semaphore and return its index
func NewSem() (Sid32, error) {
	for i := 0; i < NSEM; i++ {
		sem := NextSem  // current semaphore index to try

		NextSem++
		if int(NextSem) >= NSEM {
			NextSem = 0 // round back to index zero
		}

		if SemTab[sem].SState == SFree {  // found a free semaphore entry to use
			SemTab[sem].SState = SUsed
			return sem, OK
		}
	}

	return NoneSem, ErrSYSERR
}

// SemCreate function create a new semaphore and return the ID to the caller
func SemCreate(count int32) (Sid32, error) {
	mask := Disable()
	defer Restore(mask)

	sem, err := NewSem()
	if count <= 0 || err != OK {
		return NoneSem, ErrSYSERR
	}

	SemTab[sem].SCount = count

	return sem, OK
}

// SemDelete function delete a semaphore by releasing its table entry, and free all waiting processes
func SemDelete(sem Sid32) error {
	mask := Disable()
	defer Restore(mask)

	if IsBadSem(sem) {
		return ErrSYSERR
	}

	semptr := &SemTab[sem]
	if semptr.SState == SFree { // free entry no need to delete
		return ErrEMPTY
	}

	semptr.SState = SFree

	// defer resheduling before all the 
	// waiting processes are released from this semaphore waiting queue
	ReschedCntl(DeferStart)
	for ; semptr.SCount < 0; semptr.SCount++ {
		pid, err := GetFirst(semptr.SQueue)
		if err != OK {
			return err
		}

		err = Ready(pid)
		if err != OK {
			return err
		}
	}
	ReschedCntl(DeferStop)

	return OK
}

// SemReset function reset a semaphore's count and release waiting processes
// SemReset reuses a used semaphore entry instead of allocating a new one
func SemReset(sem Sid32, count int32) error {
	mask := Disable()
	defer Restore(mask)

	if count <= 0 || IsBadSem(sem) || SemTab[sem].SState == SFree {
		return ErrSYSERR
	}

	semptr := &SemTab[sem]
	semqueue := semptr.SQueue

	// defer rescheduling before free all the waiting processes
	ReschedCntl(DeferStart)
	for pid, err := GetFirst(semqueue); err == OK; {
		e := Ready(pid)
		if e != OK {
			return e
		}
	}
	semptr.SCount = count // new count for resetted semaphore
	ReschedCntl(DeferStop)

	return OK
}