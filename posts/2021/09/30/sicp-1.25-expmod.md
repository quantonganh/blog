---
title: "SICP Exercise 1.25: A simpler expmod?"
date: 2021-09-30
description:
categories:
- Programming
tags:
- sicp
- scheme
---
> [Exercise 1.25](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-11.html#%_sec_1.2.6): Alyssa P. Hacker complains that we went to a lot of extra work in writing `expmod`.
> After all, she says, since we already know how to compute exponentials, we could have simply written:

> ```racket
(define (expmod base exp m)
    (remainder (fast-expt base exp) m))
```

> Is she correct? Would this procedure serve as well for our fast prime tester? Explain.

First, look at the original algorithm:

```racket
 (define (expmod base exp m)
     (cond ((= exp 0) 1)
         ((even? exp)
             (remainder
                 (square (expmod base (/ exp 2) m))
                 m))
         (else
             (remainder
                 (* base (expmod base (- exp 1) m))
                 m))))
```

Trying to run both:

```shell
$ racket 1.25-expmod.scm
>(expmod 2 10 10)
> (expmod 2 5 10)
> >(expmod 2 4 10)
> > (expmod 2 2 10)
> > >(expmod 2 1 10)
> > > (expmod 2 0 10)
< < < 1
< < <2
< < 4
< <6
< 2
<4
4
```

```shell
$ racket 1.25-expmod-fast-expt.scm
>(expmod 2 10 10)
<4
4
```

Looks like she is correct.

But there is a problem: instead of breaking down the problem into smaller numbers, she tries to fully compute `expt` before computing `remainder`.

Let's see what happens with a large number:

```shell
$ gtime -v racket 1.25-expmod.scm
>(expmod 2 100000000 100000000)
> (expmod 2 50000000 100000000)
> >(expmod 2 25000000 100000000)
> > (expmod 2 12500000 100000000)
> > >(expmod 2 6250000 100000000)
> > > (expmod 2 3125000 100000000)
> > > >(expmod 2 1562500 100000000)
> > > > (expmod 2 781250 100000000)
> > > > >(expmod 2 390625 100000000)
> > > > > (expmod 2 390624 100000000)
> > > >[10] (expmod 2 195312 100000000)
> > > >[11] (expmod 2 97656 100000000)
> > > >[12] (expmod 2 48828 100000000)
> > > >[13] (expmod 2 24414 100000000)
> > > >[14] (expmod 2 12207 100000000)
> > > >[15] (expmod 2 12206 100000000)
> > > >[16] (expmod 2 6103 100000000)
> > > >[17] (expmod 2 6102 100000000)
> > > >[18] (expmod 2 3051 100000000)
> > > >[19] (expmod 2 3050 100000000)
> > > >[20] (expmod 2 1525 100000000)
> > > >[21] (expmod 2 1524 100000000)
> > > >[22] (expmod 2 762 100000000)
> > > >[23] (expmod 2 381 100000000)
> > > >[24] (expmod 2 380 100000000)
> > > >[25] (expmod 2 190 100000000)
> > > >[26] (expmod 2 95 100000000)
> > > >[27] (expmod 2 94 100000000)
> > > >[28] (expmod 2 47 100000000)
> > > >[29] (expmod 2 46 100000000)
> > > >[30] (expmod 2 23 100000000)
> > > >[31] (expmod 2 22 100000000)
> > > >[32] (expmod 2 11 100000000)
> > > >[33] (expmod 2 10 100000000)
> > > >[34] (expmod 2 5 100000000)
> > > >[35] (expmod 2 4 100000000)
> > > >[36] (expmod 2 2 100000000)
> > > >[37] (expmod 2 1 100000000)
> > > >[38] (expmod 2 0 100000000)
< < < <[38] 1
< < < <[37] 2
< < < <[36] 4
< < < <[35] 16
< < < <[34] 32
< < < <[33] 1024
< < < <[32] 2048
< < < <[31] 4194304
< < < <[30] 8388608
< < < <[29] 44177664
< < < <[28] 88355328
< < < <[27] 85987584
< < < <[26] 71975168
< < < <[25] 8628224
< < < <[24] 49394176
< < < <[23] 98788352
< < < <[22] 90875904
< < < <[21] 27817216
< < < <[20] 55634432
< < < <[19] 23962624
< < < <[18] 47925248
< < < <[17] 95861504
< < < <[16] 91723008
< < < <[15] 96568064
< < < <[14] 93136128
< < < <[13] 38832384
< < < <[12] 47123456
< < < <[11] 5383936
< < < <[10] 66852096
< < < < < 39593216
< < < < <79186432
< < < < 12890624
< < < <87109376
< < < 87109376
< < <87109376
< < 87109376
< <87109376
< 87109376
<87109376
87109376
	Command being timed: "racket 1.25-expmod.scm"
	User time (seconds): 0.26
	System time (seconds): 0.07
	Percent of CPU this job got: 95%
	Elapsed (wall clock) time (h:mm:ss or m:ss): 0:00.35
	Average shared text size (kbytes): 0
	Average unshared data size (kbytes): 0
	Average stack size (kbytes): 0
	Average total size (kbytes): 0
	Maximum resident set size (kbytes): 87988
	Average resident set size (kbytes): 0
	Major (requiring I/O) page faults: 0
	Minor (reclaiming a frame) page faults: 27796
	Voluntary context switches: 0
	Involuntary context switches: 385
	Swaps: 0
	File system inputs: 0
	File system outputs: 0
	Socket messages sent: 0
	Socket messages received: 0
	Signals delivered: 0
	Page size (bytes): 4096
	Exit status: 0
```

```shell
$ gtime -v racket 1.25-expmod-fast-expt.scm
>(expmod 2 100000000 100000000)
<87109376
87109376
	Command being timed: "racket 1.25-expmod-fast-expt.scm"
	User time (seconds): 0.37
	System time (seconds): 0.12
	Percent of CPU this job got: 96%
	Elapsed (wall clock) time (h:mm:ss or m:ss): 0:00.51
	Average shared text size (kbytes): 0
	Average unshared data size (kbytes): 0
	Average stack size (kbytes): 0
	Average total size (kbytes): 0
	Maximum resident set size (kbytes): 151672
	Average resident set size (kbytes): 0
	Major (requiring I/O) page faults: 0
	Minor (reclaiming a frame) page faults: 56633
	Voluntary context switches: 0
	Involuntary context switches: 439
	Swaps: 0
	File system inputs: 0
	File system outputs: 0
	Socket messages sent: 0
	Socket messages received: 0
	Signals delivered: 0
	Page size (bytes): 4096
	Exit status: 0
```

So, her version:

- requires more memory
- slower

Trying with a larger number, it took... > 2 minutes and 10 GB to run:

```shell
$ gtime -v racket 1.25-expmod-fast-expt.scm
>(expmod 2 100000000000 100000000000)
<81787109376
81787109376
	Command being timed: "racket 1.25-expmod-fast-expt.scm"
	User time (seconds): 122.83
	System time (seconds): 99.90
	Percent of CPU this job got: 86%
	Elapsed (wall clock) time (h:mm:ss or m:ss): 4:17.10
	Average shared text size (kbytes): 0
	Average unshared data size (kbytes): 0
	Average stack size (kbytes): 0
	Average total size (kbytes): 0
	Maximum resident set size (kbytes): 10491416
	Average resident set size (kbytes): 0
	Major (requiring I/O) page faults: 0
	Minor (reclaiming a frame) page faults: 60640485
	Voluntary context switches: 45
	Involuntary context switches: 598751
	Swaps: 0
	File system inputs: 0
	File system outputs: 0
	Socket messages sent: 0
	Socket messages received: 0
	Signals delivered: 0
	Page size (bytes): 4096
	Exit status: 0
```

References:

- https://codology.net/post/sicp-solution-exercise-1-25/