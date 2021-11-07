---
title: "SICP Exercise 2.43: Eight queens: interchange the order of the nested mappings"
date: 2021-11-02
description:
categories:
- Programming
tags:
- sicp
- scheme
---
> [Exercise 2.43](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_thm_2.43):
> Louis Reasoner is having a terrible time doing [exercise 2.42](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%25_thm_2.42). 
> His queens procedure seems to work, but it runs extremely slowly. 
> (Louis never does manage to wait long enough for it to solve even the 6× 6 case.) 
> When Louis asks Eva Lu Ator for help, she points out that he has interchanged the order of the nested mappings in the flatmap, writing it as
>
> ```racket
                 (flatmap
                     (trace-lambda (new-row)
                         (map (lambda (rest-of-queens)
                                 (adjoin-position new-row k rest-of-queens))
                              (queen-cols (- k 1))))
                     (enumerate-interval 1 board-size))
```

> Explain why this interchange makes the program run slowly. 
> Estimate how long it will take Louis's program to solve the eight-queens puzzle, 
> assuming that the program in [exercise 2.42](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%25_thm_2.42) solves the puzzle in time `T`.

The procedure in 2.42:

```racket
     (trace-define (queen-cols k)
         (if (= k 0)
             (list empty-board)
             (filter
                 (lambda (positions) (safe? k positions))
                 (flatmap
                     (lambda (rest-of-queens)
                         (map (trace-lambda (new-row)
                                 (adjoin-position new-row k rest-of-queens))
                              (enumerate-interval 1 board-size)))
                     (queen-cols (- k 1))))))
```

So, for each column `k`, `(queen-cols (- k 1))` is called only one time:

```shell
$ racket 2.42-eight-queens.rkt 4
>(queen 4)
>(queen-cols 4)
> (queen-cols 3)
> >(queen-cols 2)
> > (queen-cols 1)
> > >(queen-cols 0)
< < <'(())
< < '(((1 1)) ((2 1)) ((3 1)) ((4 1)))
< <'(((3 2) (1 1))
     ((4 2) (1 1))
     ((4 2) (2 1))
     ((1 2) (3 1))
     ((1 2) (4 1))
     ((2 2) (4 1)))
< '(((2 3) (4 2) (1 1))
    ((1 3) (4 2) (2 1))
    ((4 3) (1 2) (3 1))
    ((3 3) (1 2) (4 1)))
<'(((3 4) (1 3) (4 2) (2 1)) ((2 4) (4 3) (1 2) (3 1)))
'(((3 4) (1 3) (4 2) (2 1)) ((2 4) (4 3) (1 2) (3 1)))
```

In 2.43, since the order of the nested mappings has interchanged:

```racket
    (trace-define (queen-cols k)
        (if (= k 0)
            (list empty-board)
            (filter
                (lambda (positions) (safe? k positions))
                (flatmap
                    (trace-lambda (new-row)
                        (map (lambda (rest-of-queens)
                                (adjoin-position new-row k rest-of-queens))
                             (queen-cols (- k 1))))
                    (enumerate-interval 1 board-size)))))
```

For each column `k`, `(queen-cols (- k 1))` is called `board-size` times:

```shell
$ racket 2.43-eight-queens-interchange-order.rkt 4
>(queen 4)
>(queen-cols 4)
> (queen-cols 3)
> >(queen-cols 2)
> > (queen-cols 1)
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
< < '(((1 1)) ((2 1)) ((3 1)) ((4 1)))
> > (queen-cols 1)
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
< < '(((1 1)) ((2 1)) ((3 1)) ((4 1)))
> > (queen-cols 1)
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
< < '(((1 1)) ((2 1)) ((3 1)) ((4 1)))
> > (queen-cols 1)
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
> > >(queen-cols 0)
< < <'(())
< < '(((1 1)) ((2 1)) ((3 1)) ((4 1)))
< <'(((1 2) (3 1))
     ((1 2) (4 1))
     ((2 2) (4 1))
     ((3 2) (1 1))
     ((4 2) (1 1))
     ((4 2) (2 1)))
 ...
```

```shell
$ for size in (seq 3 6); echo -n $size": "; racket 2.43-eight-queens-interchange-order.rkt $size | grep -c 'queen-cols 0'; end
3: 27 = 3^3
4: 256 = 4^4
5: 3125 = 5^5
6: 46656 = 6^6
```

So, if the program in 2.42 solves the eight-queens puzzle in time T, 2.43's version will take roughly T<sup>7</sup> to complete.