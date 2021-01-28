### Xinu-Go

For better understanding the underlying principle of Operating System, I reimplement some of important module of Xinu OS, which is described from the book 《[Operating System Design: The Xinu Approach, Second Edition](https://www.amazon.com/Operating-System-Design-Approach-Second/dp/1498712436/ref=sr_1_1?dchild=1&keywords=xinu&qid=1611368492&sr=8-1)》, written by Douglas Comer.


> NOTE: this project is NOT a working OS but some reimplemented essential module from the original [X86 version Xinu](https://xinu.cs.purdue.edu/files/Xinu-code-Galileo.tar.gz)


The reimplemented modules include:<br>
1. process queue. Gather multiple queues inside a statically allocated array in queue.go file; <br>
2. process management, including process rescheduling, rescheduling defer, process suspend, resume, create function in resched.go and process.go files; <br>
3. semaphore management, including create, delete and reset semaphore. Waiting on a semaphore and signal the arrival of release of semaphore. All in semaphore.go file; <br>
4. Lower-level IPC of message. Including the message send and receive, in file message.go; <br>
5. process preemption and time-delay function, implemented in separate clock.go file; <br>
6. High-level message passing with ports. It supports message queuing, synchronously sending messages to a port, synchronously receiving messages from a port. It very much like the golang's channel. ^_^; <br> 
7. Basic memory management, including allocation and free of heap and stack memory at oppositon direction, all in memory.go file; <br>
8. Buffer pool management, including allocating and freeing of buffer from pool, which has limited memory. Buffer pool is one of the memory partition mechanism that split free memory into independent subsets. Thus, the system can guarantee that excessive requests will not lead to global deprivation.<br>


Some modifications compared with the original [X86 version Xinu](https://xinu.cs.purdue.edu/files/Xinu-code-Galileo.tar.gz) : <br>
1. Header files and C source code files, which share some similarity in functionality, have been combined into a single .go file. So most .c files under the 'system' directory have combined into .go files under the 'include' directory ; <br>
    eg: The functions and struct defination from queue.h, queue.c and getitem.c have been reimplement in queue.go out of simplicity. <br>
2.  All the uniersal return constants have been redefined by error type out of consideration for the (return values, error indication) pattern in golang ; <br>
3. The first letter of most of  names, including function name, constant, global variables, customed type, struct field etc,  have been capitalized according to the naming convention of golang; <br>
    eg: enqueue -> Enqueue; isbadqid -> IsBadQid. <br>
4. 