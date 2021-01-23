/*
resched.go is about process reschedule

files combined from the original X86 version include:
resched.h
resched.c
ready.c
*/

package include

import (
	"fmt"
)

const (
	// DeferStart means start deferred rescehduling
	DeferStart uint8 = 1
	// DeferStop means stop deferred rescehduling
	DeferStop uint8 = 2
)

// Defer struct collects items related to deferred rescheduling
type Defer struct {
	NDefers uint32 // number of outstanding defers
	Attempt bool   // was resched called during the deferral period
}

// def is used for recording reschedule defer
var def Defer

// ReadyList is the head index of ready process list
var ReadyList Qid16

// Resched function try to reschedule a new process runnin
// if you don't want current process remains eligible,
// you should change the state to other state before call Resched()
func Resched() {
	if def.NDefers > 0 { // reschedule is defered by os
		def.Attempt = true // let os know that a rescheduling attempt is made
		return
	}

	// ptold point to process table entry for the current (old soon) process
	ptold := &Proctab[CurrPid]

	// the current process remains eligible
	if ptold.PrState == PrCurr {
		if int32(ptold.PrPrio) > FirstKey(ReadyList) {
			// no ready process have larger priority, current process continue runs con
			return
		}

		// another ready process have higher priority than current process, switch to it
		ptold.PrState = PrReady
		// insert current process back to the priority list
		Insert(CurrPid, ReadyList, int32(ptold.PrPrio))
		// but current process still runs until called ctxsw
	}

	// extract the process of highest priority from the ready list
	CurrPid, _ = Dequeue(ReadyList)
	ptnew := &Proctab[CurrPid]
	ptnew.PrState = PrCurr // update it's state to PrCurr
	Preempt = QUANTUM      // reset the preempt counter for the new process

	ctxsw(ptold.PrStkPtr, ptnew.PrStkPtr) // switch context from old process to new process
	return
}

// ctxsw function wraps the ctxsw written with Assembly language in ctxsw.S
func ctxsw(oldsp, newsp *byte) {
	// TODO: link ctxsw.S with golang
	fmt.Printf("context swithched from %v to %v\n", oldsp, newsp)
}

// ReschedCntl function control whether rescheduling is defered or allowed
func ReschedCntl(d uint8) error {
	if d == DeferStart { // start defer rescheduling
		if def.NDefers == 0 { // the first time to defer rescheduling
			def.NDefers++
			def.Attempt = false // reset the rescheduling attempt mark
		} else {
			def.NDefers++ // record the total number of defer
		}
		return OK

	} else if d == DeferStop { // stop defer rescheduling
		if def.NDefers <= 0 { // something must going wrong
			return ErrSYSERR
		}

		def.NDefers-- // decrease the total number of defer by one

		// if there is no defer left and
		// rescheduling attempt is made at lease once during the defer period,
		// then start rescheduling
		if def.NDefers == 0 && def.Attempt {
			Resched()
		}
		return OK
	} else {
		return ErrSYSERR
	}
}
