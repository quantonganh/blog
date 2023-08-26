---
title: "SICP Exercise 1.14: orders of growth of count-change"
date: 2021-09-28
description:
categories:
- Programming
tags:
- sicp
- scheme
---

> [Exercise 1.14](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-11.html#%_thm_1.14):
> Draw the tree illustrating the process generated by the `count-change` procedure of section [1.2.2](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-11.html#%_sec_1.2.2) in making change for 11 cents.
> What are the orders of growth of the space and number of steps used by this process as the amount to be changed increases?

The `count-change` procedure:

```racket
(trace-define (count-change amount)
    (cc amount 5))

(trace-define (cc amount kinds-of-coins)
    (cond ((= amount 0) 1)
        ((or (< amount 0) (= kinds-of-coins 0)) 0)
        (else (+ (cc amount
                     (- kinds-of-coins 1))
                 (cc (- amount
                        (first-denomination kinds-of-coins))
                     kinds-of-coins)))))

(define (first-denomination kinds-of-coins)
    (cond ((= kinds-of-coins 1) 1)
          ((= kinds-of-coins 2) 5)
          ((= kinds-of-coins 3) 10)
          ((= kinds-of-coins 4) 25)
          ((= kinds-of-coins 5) 50)))
```

In making changes for 11 cents, this procedure can be illustrated by a tree:

```
                                                      (11, 5)
                                                  _______|_______
                                                 |               |
                                             (11, 4)        (-39, 5)
                                         _______|_______         |
                                        |               |        |
                                    (11, 3)         (-14, 4)     0
                          ____________|____________     |
                         |                         |    |   
                      (11, 2)                    (1, 3) 0
                  ______|______                _____|_____
                 |             |              |           |
              (11, 1)       (6, 2)         (1, 2)      (-9, 3)
              __|__          __|__         __|__           |
             |     |        |     |       |     |          |
         (11, 0) (10, 1) (6, 1) (1, 2) (1, 1) (-4, 2)      0
            |      |       |     __|__   |__     |
            |      |       |    |     |     |    |
            0    (9, 1)    1 (1, 1) (-4, 2) 1    0
                   |           |      |
                   |           |      |
                   1           1      0
```

Verify by running the above procedure in [debug](https://docs.racket-lang.org/reference/debugging.html) mode:

```shell
$ racket 1.14-count-change.rkt 11
>(count-change 11)
>(cc 11 5)
> (cc 11 4)
> >(cc 11 3)
> > (cc 11 2)
> > >(cc 11 1)
> > > (cc 11 0)
< < < 0
> > > (cc 10 1)
> > > >(cc 10 0)
< < < <0
> > > >(cc 9 1)
> > > > (cc 9 0)
< < < < 0
> > > > (cc 8 1)
> > > > >(cc 8 0)
< < < < <0
> > > > >(cc 7 1)
> > > > > (cc 7 0)
< < < < < 0
> > > > > (cc 6 1)
> > > >[10] (cc 6 0)
< < < <[10] 0
> > > >[10] (cc 5 1)
> > > >[11] (cc 5 0)
< < < <[11] 0
> > > >[11] (cc 4 1)
> > > >[12] (cc 4 0)
< < < <[12] 0
> > > >[12] (cc 3 1)
> > > >[13] (cc 3 0)
< < < <[13] 0
> > > >[13] (cc 2 1)
> > > >[14] (cc 2 0)
< < < <[14] 0
> > > >[14] (cc 1 1)
> > > >[15] (cc 1 0)
< < < <[15] 0
> > > >[15] (cc 0 1)
< < < <[15] 1
< < < <[14] 1
< < < <[13] 1
< < < <[12] 1
< < < <[11] 1
< < < <[10] 1
< < < < < 1
< < < < <1
< < < < 1
< < < <1
< < < 1
< < <1
> > >(cc 6 2)
> > > (cc 6 1)
> > > >(cc 6 0)
< < < <0
> > > >(cc 5 1)
> > > > (cc 5 0)
< < < < 0
> > > > (cc 4 1)
> > > > >(cc 4 0)
< < < < <0
> > > > >(cc 3 1)
> > > > > (cc 3 0)
< < < < < 0
> > > > > (cc 2 1)
> > > >[10] (cc 2 0)
< < < <[10] 0
> > > >[10] (cc 1 1)
> > > >[11] (cc 1 0)
< < < <[11] 0
> > > >[11] (cc 0 1)
< < < <[11] 1
< < < <[10] 1
< < < < < 1
< < < < <1
< < < < 1
< < < <1
< < < 1
> > > (cc 1 2)
> > > >(cc 1 1)
> > > > (cc 1 0)
< < < < 0
> > > > (cc 0 1)
< < < < 1
< < < <1
> > > >(cc -4 2)
< < < <0
< < < 1
< < <2
< < 3
> > (cc 1 3)
> > >(cc 1 2)
> > > (cc 1 1)
> > > >(cc 1 0)
< < < <0
> > > >(cc 0 1)
< < < <1
< < < 1
> > > (cc -4 2)
< < < 0
< < <1
> > >(cc -9 3)
< < <0
< < 1
< <4
> >(cc -14 4)
< <0
< 4
> (cc -39 5)
< 0
<4
4
```

#### Orders of growth

From section [1.2.2](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-11.html#%_sec_1.2.2):

> In general, the number of steps required by a tree-recursive process will be proportional to the number of nodes in the tree,
> while the space required will be proportional to the maximum depth of the tree.

##### Space

The maximum depth of the tree will always be the branch that represents the case of all pennies (`cc n 1`):

```shell
> > >(cc 11 1)
> > > (cc 11 0)
< < < 0
> > > (cc 10 1)
> > > >(cc 10 0)
< < < <0
> > > >(cc 9 1)
> > > > (cc 9 0)
< < < < 0
> > > > (cc 8 1)
> > > > >(cc 8 0)
< < < < <0
> > > > >(cc 7 1)
> > > > > (cc 7 0)
< < < < < 0
> > > > > (cc 6 1)
> > > >[10] (cc 6 0)
< < < <[10] 0
> > > >[10] (cc 5 1)
> > > >[11] (cc 5 0)
< < < <[11] 0
> > > >[11] (cc 4 1)
> > > >[12] (cc 4 0)
< < < <[12] 0
> > > >[12] (cc 3 1)
> > > >[13] (cc 3 0)
< < < <[13] 0
> > > >[13] (cc 2 1)
> > > >[14] (cc 2 0)
< < < <[14] 0
> > > >[14] (cc 1 1)
> > > >[15] (cc 1 0)
< < < <[15] 0
> > > >[15] (cc 0 1)
< < < <[15] 1
< < < <[14] 1
< < < <[13] 1
< < < <[12] 1
< < < <[11] 1
< < < <[10] 1
< < < < < 1
< < < < <1
< < < < 1
< < < <1
< < < 1
< < <1
```

So, the orders of growth of the space will be `O(n)`.

##### Number of steps

I cannot solve this part. So I googled for the solution and [this one](http://wiki.drewhess.com/wiki/SICP_exercise_1.14) helped me understand how to calculate time complexity of this procedure.
I will note it down for the future reference.

Look at the code:

```racket
    (cond ((= amount 0) 1)
        ((or (< amount 0) (= kinds-of-coins 0)) 0)
        (else (+ (cc amount
                     (- kinds-of-coins 1))
                 (cc (- amount
                        (first-denomination kinds-of-coins))
                     kinds-of-coins)))))
```

`(cond ((= amount 0) 1)` performs only one operation `O(1)`. If using only pennies, then:

```
T(n, 1) = 1 + T(n, 0) + T(n-1, 1)
T(n, 1) = 1 + 1 + T(n-1, 1)

T(n - (n-1), 1) = 1 + 1 + T(0, 1) = 1 + 1 + 1 = 3
T(n - (n-2), 1) = 1 + 1 + T(n - (n-1), 1) = 1 + 1 + 3 = 5
T(n - (n-3), 1) = 1 + 1 + T(n - (n-2), 1) = 1 + 1 + 5 = 7
T(n - (n-4), 1) = 1 + 1 + T(n - (n-3), 1) = 1 + 1 + 7 = 9
...

T(n, 1) = 2n + 1
```

Verify this by running `(cc 11 1)` in debug mode and counting the number of calls of `cc`:

```shell
$ racket 1.14-count-change.rkt 11
>(count-change 11)
>(cc 11 1)
> (cc 11 0)
< 0
> (cc 10 1)
> >(cc 10 0)
< <0
> >(cc 9 1)
> > (cc 9 0)
< < 0
> > (cc 8 1)
> > >(cc 8 0)
< < <0
> > >(cc 7 1)
> > > (cc 7 0)
< < < 0
> > > (cc 6 1)
> > > >(cc 6 0)
< < < <0
> > > >(cc 5 1)
> > > > (cc 5 0)
< < < < 0
> > > > (cc 4 1)
> > > > >(cc 4 0)
< < < < <0
> > > > >(cc 3 1)
> > > > > (cc 3 0)
< < < < < 0
> > > > > (cc 2 1)
> > > >[10] (cc 2 0)
< < < <[10] 0
> > > >[10] (cc 1 1)
> > > >[11] (cc 1 0)
< < < <[11] 0
> > > >[11] (cc 0 1)
< < < <[11] 1
< < < <[10] 1
< < < < < 1
< < < < <1
< < < < 1
< < < <1
< < < 1
< < <1
< < 1
< <1
< 1
<1
1

$ for amount in (seq 1 11); echo -n "T($amount, 1): "; racket 1.14-count-change.rkt $amount | grep -c 'cc'; end
T(1, 1): 3
T(2, 1): 5
T(3, 1): 7
T(4, 1): 9
T(5, 1): 11
T(6, 1): 13
T(7, 1): 15
T(8, 1): 17
T(9, 1): 19
T(10, 1): 21
T(11, 1): 23
```

Let's go one step further with 2 kind of coins:

```
T(n, 2) = 1 + T(n, 1) + T(n-5, 2)
T(n, 1) = 2n + 1
T(n-5, 2) = 1 + T(n-5, 1) + T(n-10, 2) = 1 + (2n + 1) + T(n-10, 2)
T(n-10, 2) = 1 + (2n + 1) + (2n + 1) + T(n-15, 2)
...

T(n, 2) = (n/5) * 2n + 1
```

When k = 3:

```
T(n, 3) = 1 + T(n, 2) + T(n-10, 3)
T(n, 2) = (n/5) * 2n + 1
T(n-10, 3) = 1 + T(n-10, 2) + T(n-20, 3) = 1 + ((n/5) * 2n + 1) + T(n-20, 3)
...

T(n, 3) = (n/10) * ((n/5) * 2n + 1)
```

k = 4:

```
T(n, 4) = (n/25) * T(n, 3)
```

k = 5:

```
T(n, 5) = (n/50) * T(n, 4)
```

So, the time complexity of the `count-change` procedure with 5 kinds of coins is: O(n<sup>5</sup>).