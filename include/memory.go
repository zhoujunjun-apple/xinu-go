/*
memory.go memory manage module

files combined from the original X86 version include:
memory.h
getstk.c

*/

package include

// RoundMB function round(look up) x to the minimum memory block size, which is multiples of 8
// eg: 25 -> 32
//     41 -> 48
//     105 -> 112
//     305 -> 312
//     1001 -> 1008
func RoundMB(x int32) int32 {
	return (7 + x) & (^7)
}

// TruncMB function round(look down) x to the maximum memory block size, which is multiples of 8
// eg: 25 -> 24
//     41 -> 40
//     105 -> 104
//     305 -> 304
//     1001 -> 1000
func TruncMB(x int32) int32 {
	return x & (^7)
}

// GetStk function allocates stack memory, returning highest word address
// TODO: complete this function
func GetStk(nbytes uint32) (*uint32, error) {
	var fakeStk uint32 = 1
	return &fakeStk, OK
}
