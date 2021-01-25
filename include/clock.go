/* clock.go clock manager, including timed delay and preemption functionality.

####################################################################
Efficient management of delay with a delta list:
input delay(ms): 1, 2, 3
output delta list: 1 -> 1 -> 1

input delay(ms): 1, 1, 3, 4, 4, 9
output delta list: 1 -> 0 -> 2 -> 1 -> 0 -> 5

In summary, the key of each other process on a delta list specifies
the number of clock ticks the process must delay BEYOND the PRECEDING
process on the list.
####################################################################

files combined from the original X86 version include:
clock.h
insertd.c
unsleep.c
sleep.c

*/

package include

import "math"

// MAXSECONDS is the max seconds per 32-bit msec
const MAXSECONDS uint32 = math.MaxUint32 / 1000

// Preempt is the preemption counter
var Preempt uint8

// sleepq is the queue id of the global sleep queue
var sleepq Qid16

// InsertDelta function insert a process in delta list using delay as the key
// pid: process id of to be inserted;
// q: the id of delta queue, which is actually the 'sleepq' variable;
// key: time delay from 'now' in millisecond, or in clock tick;
func InsertDelta(pid Pid32, q Qid16, key int32) error {
	if IsBadPid(pid) || IsBadQid(q) {
		return ErrSYSERR
	}

	prev := QueueHead(q)
	next := Queuetab[prev].Qnext

	// search the delta list until the tail node reached or
	// a right position is find for insert  
	for next != QueueTail(q) && Queuetab[next].Qkey <= key {
		// time delay if all the previous nodes awaken
		key -= Queuetab[next].Qkey
		prev = next
		next = Queuetab[next].Qnext
	}

	// insert new node between prev and next nodes
	Queuetab[pid].Qnext = next
	Queuetab[pid].Qprev = prev
	Queuetab[prev].Qnext = Qid16(pid)
	Queuetab[next].Qprev = Qid16(pid)

	if next != QueueTail(q) {
		// substract the extra delay that the new node introduced
		Queuetab[next].Qkey -= key
	}

	return OK
}

// Unsleep function removes a process from the sleep queue.
func Unsleep(pid Pid32) error {
	// TODO
	return OK
}

// Sleep function delay the calling process 'delay' seconds
func Sleep(delay uint32) error {
	if delay > MAXSECONDS {
		return ErrSYSERR
	}

	err := Sleepms(delay*1000)
	return err
}

// Sleepms function delay the calling process 'delayms' milliseconds
func Sleepms(delayms uint32) error {
	mask := Disable()
	defer Restore(mask)

	// if delay 0 ms, then try to yield the CPU
	if delayms == 0 {
		Yield()
		return OK
	}

	// put the calling process into delta list
	if err := InsertDelta(CurrPid, sleepq, int32(delayms)); err != OK {
		return ErrSYSERR
	}

	// update its state and rescheduling
	Proctab[CurrPid].PrState = PrSleep
	Resched()

	return OK
}