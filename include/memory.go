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
	// MNext is the pointer to next free memory block
	MNext *MemBlk

	// MLength is the size of memory block including itself
	MLength uint32
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

// GetMem function allocate heap storage, returning the lowest word address.
// This function try to allocate memory from the FIRST adequate memory block
// of the free memory list. Converse strategy for GetStk().
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

// GetStk function allocates stack memory, returning highest word address.
// This function try to allocate memory from the LAST adequate memory 
// block of the free memory list. Converse strategy for GetMem().
func GetStk(nbytes uint32) (unsafe.Pointer, error) {
	mask := Disable()
	defer Restore(mask)

	if nbytes == 0 {
		return NonePointer, ErrSYSERR
	}

	nbytes = uint32(RoundMB(int32(nbytes)))
	
	prev := &freememlist
	curr := freememlist.MNext

	var fits, fitsprev *MemBlk = nil, nil 
	for curr != nil {
		if curr.MLength >= nbytes { // find the last adequate memory block
			fits = curr
			fitsprev = prev
		}

		prev = curr
		curr = curr.MNext
	}

	if fits == nil {  // no adequate memory block found
		return NonePointer, ErrEMPTY
	}

	if fits.MLength == nbytes {  // the block exactly match
		fitsprev.MNext = fits.MNext
	} else {
		fits.MLength -= nbytes

		/*
	 fitsPointer   fits(the following one)
	   \|/                 \|/
		|-------------------|----------|
		 <--fits.MLength---> <-nbytes-->		 
		*/

		fitsPointer := unsafe.Pointer(fits)
		fits = (*MemBlk)((unsafe.Pointer)(uintptr(fitsPointer) + uintptr(fits.MLength)))
	}

	freememlist.MLength -= nbytes

	// retFits point to the piece from the highest part of the selected block
	fitsPointer := unsafe.Pointer(fits)
	retFits := (unsafe.Pointer)(
		uintptr(fitsPointer) + 
	    uintptr(nbytes) - 
		unsafe.Sizeof(uint32(1)))
	// Minus sizeof(uint32) from the fitsPointer is because
	// the xinu os assumed the underlying hardware is 32 bits pattern.
	// And GetStk function is used for allocating stack memory, each push 
	// operation would consume 32 bits memory. When returned from GetStk
	// function, the 4 bytes memory pointed by retFits, as follows, is 
	// converted to *uint32 type and used, in Create() function, to save 
	// the StackMagic constants of uint32 type. 
	// lower address        retFits                    higher address
	//                       \|/
	// | byte | byte | .....  |      |      |      |      |
	//                         <-- first push consumes -->

	return retFits, OK
}

// FreeMem function free a memory block, pointed by blkaddr, returning 
// the block to the free memory list. The nbytes arg specify the size
// of block to be freed in bytes.
func FreeMem(blkaddr unsafe.Pointer, nbytes uint32) error {
	mask := Disable()
	defer Restore(mask)

	if nbytes == 0 || uintptr(blkaddr) < uintptr(minheap) || uintptr(blkaddr) > uintptr(maxheap) {
		return ErrSYSERR
	}

	nbytes = uint32(RoundMB(int32(nbytes)))
	block := (*MemBlk)(blkaddr)

	prev := &freememlist
	next := freememlist.MNext

	// the free memory list are ordered by the physical address in ascending order
	for next != nil && uintptr(unsafe.Pointer(next)) < uintptr(blkaddr) {
		prev = next
		next = next.MNext
	}

	// topPrev is the top position of previous block
	var topPrev *MemBlk
	if prev == &freememlist {
		topPrev = nil
	} else {
		topPrev = (*MemBlk)(unsafe.Pointer(uintptr(unsafe.Pointer(prev)) + uintptr(prev.MLength)))
	}

	/*
	prev             topPrev    block       block+nbytes      next        next+next.MLength
    \|/                 \|/      \|/             \|/          \|/           \|/  
	 |-------------------|++++++++|---------------|++++++++++++|-------------|
	*/

	// check if the blkaddr block overlap previous block
	overlapPrev := prev != &freememlist && uintptr(unsafe.Pointer(block)) < uintptr(unsafe.Pointer(topPrev)) 
	
	// check if the blkaddr block overlap next block
	overlapNext := next != nil && (uintptr(unsafe.Pointer(block)) + uintptr(nbytes)) > uintptr(unsafe.Pointer(next))
	
	if overlapPrev || overlapNext {  // should not overlap
		return ErrSYSERR
	}

	freememlist.MLength += nbytes

	if topPrev == block {// coalesce with previous block, blend into the previous
		prev.MLength += nbytes
		block = prev
	} else { // not coalesce, link into list as new node
		block.MNext = next
		block.MLength = nbytes
		prev.MNext = block
	}

	// check if it coalesce with next block
	if uintptr(unsafe.Pointer(next)) == uintptr(unsafe.Pointer(block)) + uintptr(block.MLength) {
		// ..., the next block is blend into it
		block.MLength += next.MLength
		block.MNext = next.MNext
	}

	return OK
}

// FreeStk function free stack memory allocated by GetStk
func FreeStk(blkaddr unsafe.Pointer, nbytes uint32) error {
	// the converse operation with the retFits in GetStk
	nbytes = uint32(RoundMB(int32(nbytes)))
	realaddr := unsafe.Pointer(uintptr(blkaddr) - uintptr(nbytes) + unsafe.Sizeof(uint32(1)))
	
	return FreeMem(realaddr, nbytes)
}