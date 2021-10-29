---
title: "SICP Exercise 2.42: Eight queens puzzle"
date: 2021-10-29
description:
categories:
- Programming
tags:
- sicp
- scheme
---
> [Exercise 2.42](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_thm_2.42):
> The "eight-queens puzzle" asks how to place eight queens on a chessboard so that no queen is in check from any other (i.e., no two queens are in the same row, column, or diagonal).
> 
> One way to solve the puzzle is to work across the board, placing a queen in each column.
> Once we have placed `k - 1` queens, we must place the `k`<sup>th</sup> queen in a position where it does not check any of the queens already on the board.
> We can formulate this approach recursively: Assume that we have already generated the sequence of all possible ways to place `k - 1` queens in the first `k - 1` columns of the board.
> For each of these ways, generate an extended set of positions by placing a queen in each row of the `k`<sup>th</sup> column.
> Now filter these, keeping only the positions for which the queen in the `k`<sup>th</sup> column is safe with respect to the other queens.
> This produces the sequence of all ways to place `k` queens in the first `k` columns. By continuing this process, we will produce not only one solution, but all solutions to the puzzle.

> We implement this solution as a procedure `queens`, which returns a sequence of all solutions to the problem of placing `n` queens on an `n×n` chessboard.
> Queens has an internal procedure `queen-cols` that returns the sequence of all ways to place queens in the first `k` columns of the board.

> ```racket
(define (queen board-size)
    (define (queen-cols k)
        (if (= k 0)
            (list empty-board)
            (filter
                (lambda (positions) (safe? k positions))
                (flatmap
                    (lambda (rest-of-queens)
                        (map (lambda (new-row)
                                (adjoin-position new-row k rest-of-queens))
                             (enumerate-interval 1 board-size)))
                    (queen-cols (- k 1))))))
    (queen-cols board-size))
```

> In this procedure `rest-of-queens` is a way to place `k - 1` queens in the first `k - 1` columns,
> and `new-row` is a proposed row in which to place the queen for the `k`<sup>th</sup> column.
> Complete the program by implementing the representation for sets of board positions,
> including the procedure `adjoin-position`, which adjoins a new row-column position to a set of positions, 
> and `empty-board`, which represents an empty set of positions. 
> You must also write the procedure `safe?`, which determines for a set of positions, whether the queen in the `k`<sup>th</sup> column is safe with respect to the others. 
> (Note that we need only check whether the new queen is safe -- the other queens are already guaranteed safe with respect to each other.)

I will try to implement the simpler procedures first.

`empty-board` represents an empty set of positions:

```racket
(define empty-board '())
```

`adjoin-position` adjoins a new row-column position to a set of positions:

```racket
(define (adjoin-position row col rest-of-queens)
    (cons (list row col) rest-of-queens))
```

`attack?` checks if two queens are in the same row, column, or diagonal:

```racket
(define (attack? q1 q2)
    (or (= (row q1) (row q2))
        (= (abs (- (row q1) (row q2)))
           (abs (- (col q1) (col q2))))))

(define (row position)
    (car position))

(define (col position)
    (cadr position))
```

(*No need to check if they are in the same column*)

The most difficult procedure is `safe?`: which determines for a set of positions, whether the queen in the `k`<sup>th</sup> column is safe with respect to the others.

The steps should be something like this:

- get the queen in the `k`<sup>th</sup> column
- get the first element of the `rest-of-queens`
- check if two queens are under attack
  - if yes, return `#f` (false)
  - else: remove the first element from `rest-of-queens` and do the same with the remaining elements

Here's the code:

```racket
(define (safe? k positions)
    (if (= (length positions) 1)
        #t
        (if (attack? (k-queen positions) (first-queen positions))
            #f
            (safe? k (remove (first-queen positions) positions)))))

(define (k-queen positions)
    (car positions))

(define (first-queen positions)
    (cadr positions))
```

The `remove` procedure is similar to the one at the end of [Nested Mappings](https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-15.html#%_sec_Temp_193) section but uses `equals?` to compare list:

```racket
(trace-define (remove item sequence)
    (filter (lambda (x) (not (equal? x item)))
            sequence))
```

The completed program:

```racket
(trace-define (queen board-size)
    (define (queen-cols k)
        (if (= k 0)
            (list empty-board)
            (filter
                (lambda (positions) (safe? k positions))
                (flatmap
                    (lambda (rest-of-queens)
                        (map (lambda (new-row)
                                (adjoin-position new-row k rest-of-queens))
                             (enumerate-interval 1 board-size)))
                    (queen-cols (- k 1))))))
    (trace queen-cols)
    (queen-cols board-size))

(define empty-board '())

(define (adjoin-position row col rest-of-queens)
    (cons (list row col) rest-of-queens))

(define (safe? k positions)
    (if (= (length positions) 1)
        #t
        (if (attack? (k-queen positions) (first-queen positions))
            #f
            (safe? k (remove (first-queen positions) positions)))))

(define (k-queen positions)
    (car positions))

(define (first-queen positions)
    (cadr positions))

(define (attack? q1 q2)
    (or (= (row q1) (row q2))
        (= (abs (- (row q1) (row q2)))
           (abs (- (col q1) (col q2))))))

(define (row position)
    (car position))

(define (col position)
    (cadr position))

(define (remove item sequence)
    (filter (lambda (x) (not (equal? x item)))
            sequence))
```

Test:

```shell
$ racket 2.42-eight-queens.rkt 
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

```
|---|---|---|---|
|   |   | X |   |
|---|---|---|---|
| X |   |   |   |
|---|---|---|---|
|   |   |   | X |
|---|---|---|---|
|   | X |   |   |
|---|---|---|---|
```

```
|---|---|---|---|
|   | X |   |   |
|---|---|---|---|
|   |   |   | X |
|---|---|---|---|
| X |   |   |   |
|---|---|---|---|
|   |   | X |   |
|---|---|---|---|
```