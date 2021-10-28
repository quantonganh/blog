---
title: "SICP Exercise 2.41: Triple sum"
date: 2021-10-18
description:
categories:
- Programming
tags:
- sicp
- scheme
---
> [Exercise 2.41](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_thm_2.41): Write a procedure to find all ordered triples of distinct positive integers `i`, `j`, and `k` less than or equal to a given integer `n` that sum to a given integer `s`.

`unique-triples` can be written easily base on `unique-pairs` in [2.40](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_thm_2.40):

```racket
(define (unique-triples n)
    (flatmap
        (lambda (i)
            (flatmap
                (lambda (j)
                    (map (lambda (k) (list i j k))
                         (enumerate-interval 1 (- j 1))))
                (enumerate-interval 1 (- i 1))))
        (enumerate-interval 1 n)))
```

Define `make-triple` to convert a triple to a list:

```racket
(define (make-triple triple)
    (list (car triple) (cadr triple) (car (cddr triple))))
```

`equal-given-sum?` is used to check if sum of a triple is equal to a given integer `s`:

```racket
(define (equal-given-sum? triple s)
    (= s (+ (car triple) (cadr triple) (car (cddr triple)))))
```

and the final procedure:

```racket
(define (triple-sum n s)
    (map make-triple (filter equal-given-sum? (unique-triples n))))
```

Test:

```shell
$ racket 2.41-triple-sum.rkt
filter: contract violation
  expected: (any/c . -> . any/c)
  given: #<procedure:equal-given-sum?>
  context...:
   /Users/quanta/github.com/quantonganh/sicp-exercises/2.41-triple-sum.rkt:28:0
   body of "/Users/quanta/github.com/quantonganh/sicp-exercises/2.41-triple-sum.rkt"
```

Take a look at the [filter](https://docs.racket-lang.org/reference/pairs.html#%28def._%28%28lib._racket%2Fprivate%2Flist..rkt%29._filter%29%29) syntax:

```racket
(filter pred lst) → list?

  pred : procedure?
  lst : list?
```

> Returns a list with the elements of `lst` for which `pred` produces a true value. The `pred` procedure is applied to each element from first to last.

So, the reason is the predicate only takes one parameter while the `equal-given-sum?` takes 2 parameters.

We can fix that by using `lambda`:

```racket
(define (triple-sum n s)
    (map make-triple (filter (lambda (triple)
                                (equal-given-sum? triple s))
                             (unique-triples n))))
```

Verify:

```shell
$ racket 2.41-triple-sum.rkt
>(triple-sum 4 8)
> (accumulate #<procedure:append> '() '())
< '()
> (accumulate #<procedure:append> '() '(()))
> >(accumulate #<procedure:append> '() '())
< <'()
< '()
> (accumulate #<procedure:append> '() '(() ((3 2 1))))
> >(accumulate #<procedure:append> '() '(((3 2 1))))
> > (accumulate #<procedure:append> '() '())
< < '()
< <'((3 2 1))
< '((3 2 1))
> (accumulate #<procedure:append> '() '(() ((4 2 1)) ((4 3 1) (4 3 2))))
> >(accumulate #<procedure:append> '() '(((4 2 1)) ((4 3 1) (4 3 2))))
> > (accumulate #<procedure:append> '() '(((4 3 1) (4 3 2))))
> > >(accumulate #<procedure:append> '() '())
< < <'()
< < '((4 3 1) (4 3 2))
< <'((4 2 1) (4 3 1) (4 3 2))
< '((4 2 1) (4 3 1) (4 3 2))
> (accumulate
   #<procedure:append>
   '()
   '(() () ((3 2 1)) ((4 2 1) (4 3 1) (4 3 2))))
> >(accumulate
    #<procedure:append>
    '()
    '(() ((3 2 1)) ((4 2 1) (4 3 1) (4 3 2))))
> > (accumulate #<procedure:append> '() '(((3 2 1)) ((4 2 1) (4 3 1) (4 3 2))))
> > >(accumulate #<procedure:append> '() '(((4 2 1) (4 3 1) (4 3 2))))
> > > (accumulate #<procedure:append> '() '())
< < < '()
< < <'((4 2 1) (4 3 1) (4 3 2))
< < '((3 2 1) (4 2 1) (4 3 1) (4 3 2))
< <'((3 2 1) (4 2 1) (4 3 1) (4 3 2))
< '((3 2 1) (4 2 1) (4 3 1) (4 3 2))
<'((4 3 1))
'((4 3 1))
```