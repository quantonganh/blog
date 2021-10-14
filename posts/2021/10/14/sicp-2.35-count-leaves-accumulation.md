---
title: "SICP Exercise 2.35: Counting leaves of a tree"
date: 2021-10-14
description:
categories:
- Programming
tags:
- sicp
- scheme
---
> [Exercise 2.35](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_thm_2.35): Redefine count-leaves from section [2.2.2](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_sec_2.2.2) as an accumulation:

> ```racket
(define (count-leaves t)
    (accumulate <??> <??> (map <??> <??>)))
```

The `count-leaves` procedure from section 2.2.2:

```racket
(define (count-leaves x)
    (cond ((null? x) 0)
          ((not (pair? x)) 1)
          (else (+ (count-leaves (car x))
                   (count-leaves (cdr x))))))
```

The `accumulate` procedure from section 2.2.3:

```racket
(define (accumulate op initial sequence)
    (if (null? sequence)
        initial
        (op (car sequence)
            (accumulate op initial (cdr sequence)))))
```

The first thing comes to my mind is: to count leaves of a tree, we need to flatten out that tree.
Recall the `fringe` produce from exercise [2.28](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%25_thm_2.28):

```racket
(define (fringe x)
    (cond ((null? x) null)
        ((pair? (car x))
            (append (fringe (car x)) (fringe (cdr x))))
        (else
            (cons (car x) (fringe (cdr x))))))
```

Let's see what is the output if we apply `fringe` function to each element of the input list:

```racket
(define t (cons (list 1 2) (list 3 4)))
(map fringe t)
```

```shell
$ racket 2.35-count-leaves-accumulation.rkt
>(fringe '(1 2))
> (fringe 1)
< '(1)
> (fringe '(2))
> >(fringe 2)
< <'(2)
> >(fringe '())
< <'()
< '(2)
<'(1 2)
>(fringe 3)
<'(3)
>(fringe 4)
<'(4)
'((1 2) (3) (4))
```

Now we can just compute the length of each element:

```racket
(define (count-leaves t)
    (accumulate (lambda (x y)
                    (+ (length x) y))
                0
                (map fringe t)))

(define t (cons (list 1 2) (list 3 4)))
(map fringe t)
(count-leaves t)
```

Test:

```shell
$ racket 2.35-count-leaves-accumulation.rkt
'((1 2) (3) (4))
>(accumulate #<procedure:...es-accumulation.rkt:15:16> 0 '((1 2) (3) (4)))
> (accumulate #<procedure:...es-accumulation.rkt:15:16> 0 '((3) (4)))
> >(accumulate #<procedure:...es-accumulation.rkt:15:16> 0 '((4)))
> > (accumulate #<procedure:...es-accumulation.rkt:15:16> 0 '())
< < 0
< <1
< 2
<4
4
```

After coming up with this solution, I [looked](https://billthelizard.blogspot.com/2011/04/sicp-235-counting-leaves-of-tree.html) [around](http://jots-jottings.blogspot.com/2011/10/sicp-exercise-235-counting-leaves-again.html) to see if there is a better way.
So, instead of applying `fringe` to each element of the input list (`map fringe t`), what function can be used to apply to the entire list?

```racket
(map <??> (fringe t))
```

Let's print the output of `fringe t`:

```racket
(define t (cons (list 1 2) (list 3 4)))
(fringe t)
```

```shell
$ racket 2.35-count-leaves-accumulation.rkt
>(fringe '((1 2) 3 4))
> (fringe '(1 2))
> >(fringe 1)
< <'(1)
> >(fringe '(2))
> > (fringe 2)
< < '(2)
> > (fringe '())
< < '()
< <'(2)
< '(1 2)
> (fringe '(3 4))
> >(fringe 3)
< <'(3)
> >(fringe '(4))
> > (fringe 4)
< < '(4)
> > (fringe '())
< < '()
< <'(4)
< '(3 4)
<'(1 2 3 4)
'(1 2 3 4)
```

This is a flattened list. To count leaves, we can just simply return 1 for each element:

```racket
(define (count-leaves t)
    (accumulate +
                0
                (map (lambda (x) 1) (fringe t))))
```

Test:

```shell
$ racket 2.35-count-leaves-accumulation.rkt
'(1 2 3 4)
>(accumulate #<procedure:+> 0 '(1 1 1 1))
> (accumulate #<procedure:+> 0 '(1 1 1))
> >(accumulate #<procedure:+> 0 '(1 1))
> > (accumulate #<procedure:+> 0 '(1))
> > >(accumulate #<procedure:+> 0 '())
< < <0
< < 1
< <2
< 3
<4
4
```