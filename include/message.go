/*
message.go inter-process communication of message

files combined from the original X86 version include:
send.c
receive.c
recvclr.c

*/

package include

// Send function pass a message to process and start recepient if waiting
func Send(pid Pid32, msg Umsg32) error {
	mask := Disable()
	defer Restore(mask)

	if IsBadPid(pid) {
		return ErrSYSERR
	}

	prptr := &Proctab[pid]
	if prptr.PrState == PrFree || prptr.PrHasMsg {
		// if there is a previous message to be received, do not overwrite it
		return ErrSYSERR
	}

	// save the msg to process pid and notify it by set the PrHasMsg field
	prptr.PrMsg = msg
	prptr.PrHasMsg = true

	if prptr.PrState == PrRecv {
		// if process pid is in PrRecv state, make it ready
		Ready(pid)
	} else if prptr.PrState == PrRecTime {
		// the process pid is in PrRecTime state,
		// remove the process from the sleep queue and make it ready.
		Unsleep(pid)
		Ready(pid)
	}

	return OK
}

// Receive function wait for message and return the message to the caller
func Receive() Umsg32 {
	mask := Disable()
	defer Restore(mask)

	prptr := &Proctab[CurrPid]
	if prptr.PrHasMsg == false {
		// no message available now, waiting for it
		prptr.PrState = PrRecv
		// give chance to another process to run
		Resched()
		// when returned from Resched(), it means another process
		// must have send message to it by calling Send() function.
	}

	msg := prptr.PrMsg     // retrieve message and save it on stack
	prptr.PrHasMsg = false // reset message nofity flag

	// DO NOT: return prptr.PrMsg
	// because after restoring the interrupt, another interrupt could occur,
	// and may overwrite the prptr.PrMsg filed since PrHasMsg has been reset.
	return msg
}

// RecvClr function clear incoming message and return message if one message is
// waitting to be retrieved. It not block when there is no message available.
func RecvClr() (Umsg32, error) {
	mask := Disable()
	defer Restore(mask)

	prptr := &Proctab[CurrPid]
	if prptr.PrHasMsg == false {
		// no rescheduling when no message available
		return NoneMsg, ErrEMPTY
	}

	msg := prptr.PrMsg
	prptr.PrHasMsg = false

	return msg, OK
}
