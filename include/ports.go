/* ports.go port management

files combined from the original X86 version include:
ports.h
ptinit.c
ptcreate.c
ptsend.c
ptrecv.c

*/

package include

import (
	"unsafe"
)

const (
	// MaxPorts is the maximum number of ports. renamed from NPORTS
	MaxPorts int = 30
	// MaxMsgs is the total messages in system. renamed from PT_MSGS
	MaxMsgs int = 100

	// port state

	// PtStateFree : port is free
	PtStateFree uint16 = 1
	// PtStateLimbo : port is being deleted/reset
	PtStateLimbo uint16 = 2
	// PtStateAlloc : port is allocated
	PtStateAlloc uint16 = 3
)

// MsgNode struct is a node on list of messages
type MsgNode struct {
	PtMsg  Umsg32   // a one-word message
	PtNext *MsgNode // pointer to next node on list
}

// PtEntry struct is the entry in port table
type PtEntry struct {
	PtSsem Sid32 // sender semaphore
	PtRsem Sid32 // receiver semaphore

	PtState  uint16 // port state: free, limbo, alloc
	PtMaxCnt uint16 // max messages to be queued
	PtSeq    int32  // sequence change at creation. much like the sequence number of TCP

	PtHead *MsgNode // head of message list
	PtTail *MsgNode // tail of message list
}

// PortTab is the global port table
var PortTab [MaxPorts]PtEntry

// PtNextID is the next table entry to try in PortTab
var PtNextID int

// ptfree is the head of global list of free message node
var ptfree *MsgNode

// IsBadPort function check if portid is bad
func IsBadPort(portid int32) bool {
	return portid < 0 || portid >= int32(MaxPorts)
}

// PtInit function initialize all ports and initialize the free message list
func PtInit(maxmsgs int32) error {
	// TODO: ptfree = GetMem()
	err := OK
	if err != OK {
		return err
	}

	// allocate port entry starting from index 0
	PtNextID = 0

	// initialize all port table entries to free
	for i := 0; i < MaxPorts; i++ {
		PortTab[i].PtState = PtStateFree
		PortTab[i].PtSeq = 0
	}

	// create a free list of message nodes linked together
	curr, next := ptfree, ptfree
	for maxmsgs--; maxmsgs > 0; curr = next {
		// ++next
		nextPtr := unsafe.Pointer(next)
		next = (*MsgNode)(unsafe.Pointer(uintptr(nextPtr) + 1))

		curr.PtNext = next
	}

	// set the pointer of tail node to nil
	curr.PtNext = nil

	return OK
}

// PtCreate function create a port that allows 'count' outstanding messages
func PtCreate(count uint16) (int, error) {
	mask := Disable()
	defer Restore(mask)

	if count <= 0 {
		return -1, ErrSYSERR
	}

	for i := 0; i < MaxPorts; i++ {
		ptnum := PtNextID // try allocate this port entry

		// update PtNextID for next try
		PtNextID++
		if PtNextID >= MaxPorts {
			PtNextID = 0
		}

		ptptr := &PortTab[ptnum]
		if ptptr.PtState == PtStateFree { // free port entry can allocate
			ptptr.PtState = PtStateAlloc // update state

			ptptr.PtRsem, _ = SemCreate(0)            // cannot receive message right now
			ptptr.PtSsem, _ = SemCreate(int32(count)) // can send message count times

			ptptr.PtHead = nil
			ptptr.PtTail = nil

			ptptr.PtSeq++
			ptptr.PtMaxCnt = count

			return ptnum, OK
		}
	}

	return -1, ErrEMPTY
}

// PtSend function send a message to a port by adding it to the tail of queue
func PtSend(portid int32, msg Umsg32) error {
	mask := Disable()
	defer Restore(mask)

	if IsBadPort(portid) {
		return ErrSYSERR
	}

	ptptr := &PortTab[portid]
	if ptptr.PtState != PtStateAlloc {
		// message can only send to allocated port
		return ErrSYSERR
	}

	seq := ptptr.PtSeq // record the orignal sequence
	if Wait(ptptr.PtSsem) != OK || ptptr.PtState != PtStateAlloc || ptptr.PtSeq != seq {
		// because Wait() could cause current process into waitting queue,
		// the portid-th entry in port table could be deleted or reseted by another process.
		// so we need to recheck its state and sequence
		return ErrSYSERR
	}

	if ptfree == nil {
		// there is no more free message node
		return ErrEMPTY
	}

	// obtain node from free list by unlinking
	msgNode := ptfree
	ptfree = msgNode.PtNext

	// copy the input msg into it
	msgNode.PtNext = nil
	msgNode.PtMsg = msg

	// link into queue for the portid port entry
	tailNode := ptptr.PtTail
	if tailNode == nil {
		ptptr.PtHead = msgNode
		ptptr.PtTail = msgNode
	} else {
		tailNode.PtNext = msgNode
		ptptr.PtTail = msgNode
	}

	// let the reveiver know that there is msg avilable
	Signal(ptptr.PtRsem)

	return OK
}
