/* ports.go port management

files combined from the original X86 version include:
ports.h
ptinit.c

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
	PtSeq    int32  // sequence change at creation

	PtHead *MsgNode // head of message list
	PtTail *MsgNode // tail of message list
}

// PortTab is the global port table
var PortTab [MaxPorts]PtEntry

// PtNextID is the next table entry to try in PortTab
var PtNextID int32

// ptfree is the head of global list of free message node
var ptfree *MsgNode

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
