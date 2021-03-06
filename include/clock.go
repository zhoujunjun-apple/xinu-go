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
wakeup.c
sleepms.c
recvtime.c
clkhandler.c

*/

package include

import "math"

// MAXSECONDS is the max seconds per 32-bit msec
const MAXSECONDS uint32 = math.MaxUint32 / 1000

// Preempt is the preemption counter
var Preempt uint8 = QUANTUM

// sleepq is the queue id of the global sleep queue, initialized with NewQueue()
var sleepq Qid16

// count1000 represent milliseconds since last clock tick
var count1000 uint32 = 0

// clktime represent seconds since boot
var clktime uint32

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

// Unsleep function removes a process from the sleep queue prematurely.
func Unsleep(pid Pid32) error {
	mask := Disable()
	defer Restore(mask)

	if IsBadPid(pid) {
		return ErrSYSERR
	}

	prptr := &Proctab[pid]
	if prptr.PrState != PrSleep && prptr.PrState != PrRecTime {
		// candidate process must on the sleep queue
		return ErrSYSERR
	}

	pidNext := Queuetab[pid].Qnext
	if int(pidNext) < NPROC { // make sure pidNext not head or tail node
		// add the extra delay because process pid's sleep not actually end
		Queuetab[pidNext].Qkey += Queuetab[pid].Qkey
	}

	// get pid out of delta list
	GetItem(pid)

	return OK
}

// Wakeup function called by clock interrupt handler to awaken processes.
// It is different with Unsleep() because Wakeup() only called when it IS
// the time to awaken processes since they have sleeped JUST ENOUGH time
func Wakeup() {
	ReschedCntl(DeferStart)
	defer ReschedCntl(DeferStop)

	// if delta list is: 0 -> 0 -> 0 -> 2 -> ...,
	// then first three process's sleep period has end.
	for NonEmpty(sleepq) && FirstKey(sleepq) <= 0 {
		pid, err := Dequeue(sleepq)
		if err != OK {
			return
		}

		err = Ready(pid)
		if err != OK {
			return
		}
	}

	return

}

// Sleep function delay the calling process 'delay' seconds
func Sleep(delay uint32) error {
	if delay > MAXSECONDS {
		return ErrSYSERR
	}

	err := Sleepms(delay * 1000)
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

// RecvTime function wait specified time to receive a message and return
// maxWait: ticks to wait before time out
func RecvTime(maxWait int32) (Umsg32, error) {
	if maxWait < 0 {
		return NoneMsg, ErrSYSERR
	}

	mask := Disable()
	defer Restore(mask)

	prptr := &Proctab[CurrPid]
	if prptr.PrHasMsg == false { // no message available yet
		// sleep maxWait time waiting for message
		err := InsertDelta(CurrPid, sleepq, maxWait)
		if err != OK {
			return NoneMsg, ErrSYSERR
		}

		prptr.PrState = PrRecTime
		Resched()
		// message arrived or sleep time out
	}

	// either message available or timer expired
	msg := TimeoutMsg
	if prptr.PrHasMsg { // message available
		msg = prptr.PrMsg
		prptr.PrHasMsg = false
	}

	return msg, OK
}

// ClkHandler is the hign level clock interrupt handler
// NOTE: the clock tick is configured interrupt every 1 millisecond
func ClkHandler() {
	count1000++            // increament the ms counter, and see if a second has passed
	if count1000 >= 1000 { // a second has passed
		clktime++     // increment seconds count
		count1000 = 0 // reset the local ms counter for the next second
	}

	// handle sleeping processes if any exist
	if !IsEmpty(sleepq) {
		qid := FirstID(sleepq)
		Queuetab[qid].Qkey-- // 1 ms passed

		// the count reaches zero, at least one process need be awaken
		if Queuetab[qid].Qkey <= 0 {
			Wakeup()
		}
	}

	// decrement the preemption counter, and reschedule when
	// remaining time reaches zero (time slice for current process is expired)
	Preempt--
	if Preempt <= 0 { // give change to another process to run
		Preempt = QUANTUM
		Resched()
	}
}
