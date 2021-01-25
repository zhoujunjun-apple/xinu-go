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
func IsBadSem(s int) bool {
	return s < 0 || s >= NSEM
}
