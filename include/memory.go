/*
memory.go memory manage module

files combined from the original X86 version include:
memory.h
getstk.c

*/

package include

import (
	"unsafe"
)

const (
	// PageSize is the bytes of a single page of memory
	PageSize int16 = 4096
)

// MemBlk struct represent the basic memory management entry
type MemBlk struct {
	// MLength is the size of memory block including itself
	MLength uint32

	// MNext is the pointer to next free memory block
	MNext *MemBlk
}

// freememlist is the head of free memory block list. renamed from 'memlist'
var freememlist MemBlk

// minheap is the start address of heap
var minheap unsafe.Pointer

// maxheap is the highest valid heap address
var maxheap unsafe.Pointer

// RoundMB function round(look up) x to the minimum memory block size, which is multiples of 8.
// eg: 25 -> 32
//     41 -> 48
//     105 -> 112
//     305 -> 312
//     1001 -> 1008
func RoundMB(x int32) int32 {
	return (7 + x) & (^7)
}

// TruncMB function round(look down) x to the maximum memory block size, which is multiples of 8.
// eg: 25 -> 24
//     41 -> 40
//     105 -> 104
//     305 -> 304
//     1001 -> 1000
func TruncMB(x int32) int32 {
	return x & (^7)
}

// GetMem function allocate heap storage, returning the lowest word address
func GetMem(nbytes uint32) (unsafe.Pointer, error) {
	mask := Disable()
	defer Restore(mask)

	if nbytes == 0 {
		return NonePointer, ErrSYSERR
	}

	// use memblk multiples
	nbytes = uint32(RoundMB(int32(nbytes)))

	prev := &freememlist
	curr := freememlist.MNext

	// search free list to allocate nbytes memory
	for curr != nil {
		if curr.MLength == nbytes { // free memory block is exactly match
			// unlink this memory block from the free memory list
			prev.MNext = curr.MNext
			freememlist.MLength -= nbytes

			return unsafe.Pointer(curr), OK

		} else if curr.MLength > nbytes { // split big block
			/* curr      leftover
			    |---------|-----------------|
			    <-nbytes->
			    <--------curr.MLength------>
			*/

			// split the first nbytes for allocating
			currPointer := unsafe.Pointer(curr)
			leftover := (*MemBlk)(unsafe.Pointer(uintptr(currPointer) + uintptr(nbytes)))
			prev.MNext = leftover

			// update the remaining block information
			leftover.MNext = curr.MNext
			leftover.MLength = curr.MLength - nbytes

			// update global memory size
			freememlist.MLength -= nbytes
			return currPointer, OK
		} else {
			prev = curr
			curr = curr.MNext
		}
	}

	// can't satisfy the memory allocating request
	return NonePointer, ErrEMPTY
}

// GetStk function allocates stack memory, returning highest word address
// TODO: complete this function
func GetStk(nbytes uint32) (*uint32, error) {
	var fakeStk uint32 = 1
	return &fakeStk, OK
}
