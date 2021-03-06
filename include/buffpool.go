/*
buffpool.go buffer pool manager

files combined from the original X86 version include:
buffpool.h
bufinit.c
getbuf.c
freebuf.c

*/

package include

import "unsafe"

const (
	// MaxBuffsInPools is the maximum number of buffer pools. renamed from 'BP_MAXN'.
	MaxBuffsInPools uint16 = 2048
	// MaxPools is the maximum buffer size in bytes. renamed from 'NBPOOLS'.
	MaxPools uint8 = 20

	// MaxBuffSize is the maximum buffer size in bytes. renamed from 'BP_MAXB'
	MaxBuffSize uint16 = 8192
	// MinBuffSize is the minmum buffer size in bytes. renamed from 'BP_MINB'.
	MinBuffSize uint16 = 8
)

// BpEntry struct is the description of a single buffer pool
type BpEntry struct {
	// BpNext point to next free buffer
	// The BpNext field must be the first field. See MakeBufPool() function.
	BpNext *BpEntry
	// BpSem is the semaphore that counts buffers currently available in the pool
	BpSem Sid32
	// BpSize is the size of buffers in this pool
	BpSize uint32
}

// BuffPoolTab is the global buffer pool table
var BuffPoolTab []BpEntry

// nbpools is the current number of allocated pools
var nbpools Bpid32

// BufInit function initialize the buff pools configuration
func BufInit() error {
	nbpools = 0
	return OK
}


// MakeBufPool function allocate memory for a buffer pool and link the buffers
// bufsize: the size of a single buffer in pool;
// numbufs: the requested number of buffers in pool;
func MakeBufPool(bufsize int32, numbufs int32) (Bpid32, error) {
	mask := Disable()
	defer Restore(mask)

	if bufsize < int32(MinBuffSize) || bufsize > int32(MaxBuffSize) || numbufs < 1 || numbufs > int32(MaxBuffsInPools) || nbpools >= Bpid32(MaxPools) {
		return NoneBpid, ErrSYSERR
	}

	// round request to a multiple of 4 bytes
	bufsize = ( (bufsize + 3) & (^3))

	// the first sizeof(Bpid32) bytes are used for saving pool id 
	// when allocating buffers from this pool.
	reqMem := numbufs * (bufsize + int32(unsafe.Sizeof(Bpid32(0))))
	buf, err := GetMem(uint32(reqMem))
	if err != OK {
		return NoneBpid, err
	}

	poolid := nbpools
	nbpools++

	bpptr := &BuffPoolTab[poolid]
	bpptr.BpNext = (*BpEntry)(buf)
	// the BpSize filed is the usable valid buffer size, it not record the sizeof(Bpid32) memory
	bpptr.BpSize = uint32(bufsize)  

	if bpptr.BpSem, err = SemCreate(numbufs); err != OK {
		// if not enough semaphore, free the allocated memory
		FreeMem(buf, uint32(reqMem))
		return NoneBpid, ErrEMPTY
	}

	// bufsize now is the dividing uint size
	bufsize += int32(unsafe.Sizeof(Bpid32(0)))
	//  links the buffers together
	// for now, the first sizeof(Bpid32) bytes are used for saving
	// the BpNext field. The BpNext field must be the FIRST positon
	// in BpEntry struct.
	for numbufs--; numbufs > 0; numbufs-- {
		bpptr = (*BpEntry)(buf)
		buf = (unsafe.Pointer)(uintptr(buf) + uintptr(bufsize))
		bpptr.BpNext = (*BpEntry)(buf)
	}
	//     /-----------------\ point to here
	//     |                 \|/
	//  [BpNext|ValidBuffSize][      |             ][  nil |             ]
	//                         <-------bufsize----->

	bpptr = (*BpEntry)(buf)
	bpptr.BpNext = nil

	return poolid, OK
}

// GetBuf function  get a buffer from a preestablished buffer pool
func GetBuf(poolid Bpid32) (unsafe.Pointer, error) {
	mask := Disable()
	defer Restore(mask)

	if poolid < 0 || poolid >= nbpools {
		return NonePointer, ErrSYSERR
	}

	bpptr := &BuffPoolTab[poolid]
	Wait(bpptr.BpSem)
	bufptr := bpptr.BpNext

	bpptr.BpNext = bufptr.BpNext

	bufPidPtr := (*Bpid32)((unsafe.Pointer(bufptr)))
	*bufPidPtr = poolid
	// bufptr            bpptr.BpNext
	// \|/                   \|/
	//  [poolid|ValidBuffSize][BpNext|             ][  nil |             ]
	//        /|\                 |                /|\
	//       retptr                \----------------/ remaining buffer pool

	retptr := unsafe.Pointer(uintptr(unsafe.Pointer(bufptr)) + unsafe.Sizeof(Bpid32(0)))

	return retptr, OK
}

// FreeBuf function free a buffer that was allocated from a poll by GetBuf
func FreeBuf(bufaddr unsafe.Pointer) error {
	mask := Disable()
	defer Restore(mask)

	// extract pool id from integer prior to buffer address
	bufaddr = unsafe.Pointer(uintptr(bufaddr) - unsafe.Sizeof(Bpid32(0)))
	poolid := *(*Bpid32)(bufaddr)
	// bufaddr(new)            bpptr.BpNext
	// \|/                   \|/
	//  [poolid|ValidBuffSize][BpNext|             ][  nil |             ]
	//        /|\                 |                /|\
	//       bufaddr(old)          \----------------/ remaining buffer pool now
	
	if poolid < 0 || poolid >= nbpools {
		return ErrSYSERR
	}

	// get address of correct pool entry in table
	bpptr := &BuffPoolTab[poolid]

	// insert buffer into buffer list
	((*BpEntry)(bufaddr)).BpNext = bpptr.BpNext
	bpptr.BpNext = (*BpEntry)(bufaddr)
	//  the remaining buffer pool now
	//  /---------------------\           
	//  |                     \|/
	//  [BpNext|ValidBuffSize][BpNext|             ][  nil |             ]
	// /|\                        |                /|\
	// bpptr.BpNext               \----------------/ 

	// signal semaphore
	Signal(bpptr.BpSem)

	return OK
}