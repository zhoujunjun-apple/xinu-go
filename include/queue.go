package include

const (
	// NQENT is the default # of queue entries:
	// 1 per process plus 2 for ready list plus 2 for sleep list plus 2 per semaphore
	NQENT int = NPROC + 4 + 2*NSEM
	// EMPTY is the NULL value for qnext or qprev index
	EMPTY Qid16 = -1
	// MAXKEY is the max key that can be stored in queue
	MAXKEY int = 0x7FFFFFFF
	// MINKEY is the min key that can be stored in queue
	MINKEY int = 0x80000000
)

// Qentry struct represents a entry struction in the queue
// One per process plus two per list
type Qentry struct {
	Qkey  int32 // key on which the queue is ordered
	Qnext Qid16 // index of next process or tail
	Qprev Qid16 // index of previous process or head
}

// Queuetab array represents the table of process queues
// [0, NPROC) saves the process nodes
// [NPROC, NQENT) = 2 + 2 + 2 * NSEM, which is :
// 2: head and tail node for ready list;
// 2: head and tail node for sleep list;
// 2*NSEM: head and tail node for each semaphore;
var Queuetab [NQENT]Qentry

// QueueHead function returns the index of head node of queue q
func QueueHead(q Qid16) Qid16 {
	return q
}

// QueueTail function returns the index of tail node of queue q,
// which is right after the head node
func QueueTail(q Qid16) Qid16 {
	return q + 1
}

// FirstID function returns the index of the first process node in queue q
func FirstID(q Qid16) Qid16 {
	return Queuetab[QueueHead(q)].Qnext
}

// LastID function returns the index of the last process node in queue q
func LastID(q Qid16) Qid16 {
	return Queuetab[QueueTail(q)].Qprev
}

// IsEmpty function checks if the queue q is empty
func IsEmpty(q Qid16) bool {
	return int(FirstID(q)) >= NPROC
}

// NonEmpty function checks if the queue q is not emtpy
func NonEmpty(q Qid16) bool {
	return !IsEmpty(q)
}

// FirstKey function returns the Qkey value of the first process node in queue q
func FirstKey(q Qid16) int32 {
	return Queuetab[FirstID(q)].Qkey
}

// LastKey function returns the Qkey value of the last process node in queue q
func LastKey(q Qid16) int32 {
	return Queuetab[LastID(q)].Qkey
}

// IsBadQid function checks if queue id q is a bad.
// A valid queue id should in [NPROC, NQENT) range
func IsBadQid(q Qid16) bool {
	return (int(q) < NPROC) || (int(q) > NQENT-1)
}

// GetItem function removes a process(pid) from an arbitrary point in a queue, at which
// pid is resides before pid process calls this function
func GetItem(pid Pid32) Pid32 {
	next := Queuetab[pid].Qnext
	prev := Queuetab[pid].Qprev

	// kick out the proces pid
	Queuetab[prev].Qnext = next
	Queuetab[next].Qprev = prev

	return pid
}

// GetFirst function removes a process from the front of a queue
func GetFirst(q Qid16) (Pid32, error) {
	if IsEmpty(q) {
		return NonePid, ErrEMPTY
	}

	head := QueueHead(q)
	return GetItem(Pid32(Queuetab[head].Qnext)), OK
}

// GetLast function removes a process from the tail of a queue
func GetLast(q Qid16) (Pid32, error) {
	if IsEmpty(q) {
		return NonePid, ErrEMPTY
	}

	tail := QueueTail(q)
	return GetItem(Pid32(Queuetab[tail].Qprev)), OK
}

// Enqueue function inserts a process pid at the tail of queue q
func Enqueue(pid Pid32, q Qid16) (Pid32, error) {
	if IsBadQid(q) || IsBadPid(pid) {
		return NonePid, ErrSYSERR
	}

	tail := QueueTail(q)
	prev := Queuetab[tail].Qprev

	// insert just before tail node
	Queuetab[pid].Qnext = tail
	Queuetab[pid].Qprev = prev
	Queuetab[prev].Qnext = Qid16(pid)
	Queuetab[tail].Qprev = Qid16(pid)

	return pid, OK
}

// Dequeue function remove and return the first process on queue q
func Dequeue(q Qid16) (Pid32, error) {
	if IsBadQid(q) {
		return NonePid, ErrSYSERR
	} else if IsEmpty(q) {
		return NonePid, ErrEMPTY
	}

	pid, err := GetFirst(q)
	if err != OK {
		return NonePid, err
	}

	Queuetab[pid].Qprev = EMPTY
	Queuetab[pid].Qnext = EMPTY

	return pid, OK
}
