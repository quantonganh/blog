---
title: "SICP Exercise 1.16: Iterative Exponentiation"
date: 2021-09-29
description:
categories:
- Programming
tags:
- sicp
- scheme
- recursion
- tail-recursion
- iteration
---
I am reading [SICP](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book.html).

[Section 1.2.4](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-11.html#%_sec_1.2.4) talks about the problem of computing the exponential of a given number.

The authors start with a recursive procedure:

```racket
#lang sicp

(define (expt b n)
    (if (= n 0)
        1
        (* b (expt b (- n 1)))))
```
This requires O(n) steps and O(n) space.

then an iterative procedure:

```racket
#lang sicp

(define (expt b n)
    (expt-iter b n 1))

(define (expt-iter b counter product)
    (if (= counter 0)
        product
        (expt-iter b
                   (- counter 1)
                   (* b product))))
```
This version requires O(n) steps and O(1) space.

and finally he suggests a procedure which requires fewer steps by using successive squaring:

```racket
#lang sicp

(define (fast-expt b n)
    (cond ((= n 0) 1)
        ((even? n) (square (fast-expt b (/ n 2))))
        (else (* b (fast-expt b (- n 1))))))
```

This only requires O(log n) steps.

[Exercise 1.16](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-11.html#%_thm_1.16):

> Design a procedure that evolves an iterative exponentiation process that uses successive squaring and uses a logarithmic number of steps, as does `fast-expt`.
> (Hint: Using the observation that (b<sup>n/2</sup>)<sup>2</sup> = (b<sup>2</sup>)<sup>n/2</sup>, keep, along with the exponent `n` and the base `b`, an additional state variable `a`, and define the state transformation in such a way that the product ab<sup>n</sup> is unchanged from state to state.
> At the beginning of the process `a` is taken to be 1, and the answer is given by the value of `a` at the end of the process.
> In general, the technique of defining an invariant quantity that remains unchanged from state to state is a powerful way to think about the design of iterative algorithms.)

Here's my solution:

```racket
#lang sicp

(define (fast-expt b n)
    (fast-expt-iter b n 1))

(define (fast-expt-iter b counter product)
    (cond ((= counter 0)
            product)
        ((even? counter)
            (fast-expt-iter (square b)
                            (/ counter 2)
                            product))
        (else
            (fast-expt-iter b
                            (- counter 1)
                            (* b product)))))
```

Let's see what is the difference between recursive and iterative version?

Enable [debugging](https://docs.racket-lang.org/reference/debugging.html) and run the recursive version:

```shell
$ racket 1.16-fast-expt-recursive.scm 
>(fast-expt 2 10)
> (fast-expt 2 5)
> >(fast-expt 2 4)
> > (fast-expt 2 2)
> > >(fast-expt 2 1)
> > > (fast-expt 2 0)
< < < 1
< < <2
< < 4
< <16
< 32
<1024
1024
```

then the iterative version:

```shell
$ racket 1.16-fast-expt-iterative.scm
>(fast-expt 2 10)
>(fast-expt-iter 2 10 1)
>(fast-expt-iter 4 5 1)
>(fast-expt-iter 4 4 4)
>(fast-expt-iter 16 2 4)
>(fast-expt-iter 256 1 4)
>(fast-expt-iter 256 0 1024)
<1024
1024
```

The first thing you can easily spot is the recursive version requires more space.

The reason is after calling `fast-expt` recursively, Scheme must remember to either square or multiply the result with `b` depends on `n` is even or odd:

- (fast-expt 2 10) = (fast-expt 2 5)<sup>2</sup>
- (fast-expt 2 5) = (fast-expt 2 4) * 2
- ...

In the iterative version, Scheme doesn't have to remember anything: after it is done, it just returns the `product`.

Moreover, the iterative version also saves time:

```shell
$ time racket 1.16-fast-expt-recursive.scm

________________________________________________________
Executed in  540.91 millis    fish           external
   usr time  407.01 millis    0.25 millis  406.77 millis
   sys time  119.54 millis    1.84 millis  117.70 millis
```

```shell
$ time racket 1.16-fast-expt-iterative.scm 

________________________________________________________
Executed in  518.87 millis    fish           external
   usr time  389.04 millis    0.12 millis  388.92 millis
   sys time  116.15 millis    1.09 millis  115.06 millis
```

References:

- http://www.owlnet.rice.edu/~comp210/96spring/Labs/lab09.html
- https://blog.moertel.com/posts/2013-05-11-recursive-to-iterative.html