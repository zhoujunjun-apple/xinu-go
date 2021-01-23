/* context switching for X86 */

        .text
        .global ctxsw

/* ctxsw: the call if ctxsw(&old_sp, &new_sp) */
/* old_sp and new_sp is the PrStkPtr field in ProcEnt struct */
ctxsw:
        /* save the old process's context */
        pushl %ebp  /* push current ebp onto stack */
        movl %esp,%ebp /* record current SP in ebp */
        pushfl /* push flags register value onto the stack */
        pushal /* push general register value onto stack */

        /* last thing before switch: save the old esp */
        /* &old_sp is saved at position 8(%ebp) */
        movl 8(%ebp),%eax /* save &old_sp to eax  */
        /*now, eax has the value &old_sp */
        movl %esp,(%eax) /* save current (old soon) process's SP to PrStkPtr field */

        /* &new_sp is saved at position 12(%ebp) */
        movl 12(%ebp),%eax /* save &new_sp to eax */
        /* the next instruction switches from the old process's stack to the new process's stack */
        movl (%eax),%esp /* pop up new process's sp from PrStkPtr field to esp */

        /* restore the new process's context */
        popal  /* restore general registers */
        movl 4(%esp),%ebp
        popfl
        add $4,%esp
        ret  /* return to the new process according to the return address on stack */
