### Xinu-Go

For better understanding the underlying principle of Operating System, I reimplement some of important module of Xinu OS, which is described from the book 《[Operating System Design: The Xinu Approach, Second Edition](https://www.amazon.com/Operating-System-Design-Approach-Second/dp/1498712436/ref=sr_1_1?dchild=1&keywords=xinu&qid=1611368492&sr=8-1)》, written by Douglas Comer.

> NOTE: this project is NOT a working OS but some reimplemented essential module from the original [X86 version Xinu](https://xinu.cs.purdue.edu/files/Xinu-code-Galileo.tar.gz)

The reimplemented modules include:<br>
1. process queue. Gather multiple queues inside a statically allocated array ; <br>
2. process management, including process rescheduling, rescheduling defer; <br>

Some modifications compared with the original [X86 version Xinu](https://xinu.cs.purdue.edu/files/Xinu-code-Galileo.tar.gz) : <br>
1. Header files and C source code files, which share some similarity in functionality, have been combined into a single .go file. So most .c files under the 'system' directory have combined into .go files under the 'include' directory ; <br>
    eg: The functions and struct defination from queue.h, queue.c and getitem.c have been reimplement in queue.go out of simplicity. <br>
2.  All the uniersal return constants have been redefined by error type out of consideration for the (return values, error indication) pattern in golang ; <br>
3. The first letter of most of  names, including function name, constant, global variables, customed type, struct field etc,  have been capitalized according to the naming convention of golang; <br>
    eg: enqueue -> Enqueue; isbadqid -> IsBadQid. <br>
4. 