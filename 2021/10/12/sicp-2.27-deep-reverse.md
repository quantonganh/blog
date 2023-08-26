---
title: "SICP Exercise 2.27: Reversing nested lists"
date: 2021-10-12
description:
categories:
- Programming
tags:
- sicp
- scheme
---
> [Exercise 2.27](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_thm_2.27): Modify your reverse procedure of exercise [2.18](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_thm_2.18) to produce a deep-reverse procedure that takes a list as argument and returns as its value the list with its elements reversed and with all sublists deep-reversed as well. For example,

> ```racket
(define x (list (list 1 2) (list 3 4)))

x
((1 2) (3 4))

(reverse x)
((3 4) (1 2))

(deep-reverse x)
((4 3) (2 1))
```

First, look at my `reverse` procedure:

```racket
#lang racket/base
(require racket/trace)

(define (reverse items)
    (iter items null))

(define (iter remaining result)
    (trace iter)
    (if (null? remaining)
        result
        (iter (cdr remaining) (cons (car remaining) result))))

(trace reverse)
(reverse (list (list 1 2) (list 3 4)))
```

```shell
$ racket 2.18-reverse-list.rkt
>(reverse '((1 2) (3 4)))
>(iter '((3 4)) '((1 2)))
>(iter '() '((3 4) (1 2)))
<'((3 4) (1 2))
'((3 4) (1 2))
```

The `reverse` procedure just reverses the order of the top-level list.

To implement `deep-reverse`, we also need to reverse the order of the nested lists:

- Same as `reverse`, return `result` if the `remaining` is null
- Check the first item of the remaining:
  - If it is a pair: call `deep-reverse`
  - If not: do the same as before

Here's the code:

```racket
#lang racket/base
(require racket/trace)

(define (deep-reverse items)
    (iter items null))

(define (iter remaining result)
    (trace iter)
    (cond ((null? remaining) result)
        ((pair? (car remaining))
            (iter (cdr remaining) (cons (deep-reverse (car remaining)) result)))
        (else
            (iter (cdr remaining) (cons (car remaining) result)))))

(trace deep-reverse)
(deep-reverse (list (list 1 2) (list 3 4)))
```

Test:

```shell
$ racket 2.27-deep-reverse.rkt
>(deep-reverse '((1 2) (3 4)))
> (deep-reverse '(1 2))
> (iter '(1 2) '())
> (iter '(2) '(1))
> (iter '() '(2 1))
< '(2 1)
>(iter '((3 4)) '((2 1)))
> (deep-reverse '(3 4))
> (iter '(3 4) '())
> (iter '(4) '(3))
> (iter '() '(4 3))
< '(4 3)
>(iter '() '((4 3) (2 1)))
<'((4 3) (2 1))
'((4 3) (2 1))
```