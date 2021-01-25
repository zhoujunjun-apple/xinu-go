/* intutils.go utils function about interrupt manager

files combined from the original X86 version include:
intutils.c

 */

package include

import "fmt"

// Restore function restore(roll back) the interrupte state to im
func Restore(im IntMask) {
	fmt.Printf("restore interrupt mask to %v\n", im)
}

// Disable function disable interrupt and return the previous state
func Disable() IntMask {
	var oldIm IntMask = 0 // fake interrupt mask
	fmt.Printf("disable interrupt, previous mask is %v\n", oldIm)

	return oldIm
}