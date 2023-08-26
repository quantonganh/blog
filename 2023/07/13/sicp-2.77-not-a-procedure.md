---
title: "SICP Exercise 2.77: expected a procedure that can be applied to arguments, given #f"
date: 2023-07-13
description:
categories:
- Programming
tags:
- sicp
- racket
- scheme
---
[2.5.1  Generic Arithmetic Operations](https://mitp-content-server.mit.edu/books/content/sectbyfn/books_pres_0/6515/sicp.zip/full-text/book/book-Z-H-18.html#%_sec_2.5.1)

> Louis Reasoner tries to evaluate the expression (magnitude z) where z is the object shown in figure [2.24](https://mitp-content-server.mit.edu/books/content/sectbyfn/books_pres_0/6515/sicp.zip/full-text/book/book-Z-H-18.html#%_fig_2.24). 
> To his surprise, instead of the answer 5 he gets an error message from `apply-generic`, saying there is no method for the operation `magnitude` on the types `(complex)`.

To simplify the `install-complex-package` procedure, I made the following changes:

```scheme
(define (install-complex-package)
  (define (make-from-real-imag x y)
		((get 'make-from-real-imag 'rectangular) x y))
	(define (tag z) (attach-tag 'complex z))
	(put 'make-from-real-imag 'complex
		(lambda (x y) (tag (make-from-real-imag x y))))

(install-complex-package)

(trace-define (make-complex-from-real-imag x y)
	((get 'make-from-real-imag 'complex) x y))

(magnitude (make-complex-from-real-imag 3 4))
```

However, upon running the code, instead of the expected error from `apply-generic`, a different error occurred:

```sh
$ racket 2.77-apply-generic-magnitude.rkt
'done
>(make-complex-from-real-imag 3 4)
application: not a procedure;
 expected a procedure that can be applied to arguments
  given: #f
```

I wondered why it returned `#f` when `'make-from-real-imag 'complex` had already been `put` into the dispatch table:

```sh
$ racket --repl
Welcome to Racket v8.9 [cs].
> (require "2.77-apply-generic-magnitude.rkt")
> (get 'make-from-real-imag 'complex)
#<procedure:...neric-magnitude.rkt:75:16>
```

To debug the issue, I utilized the [racket/trace](https://docs.racket-lang.org/reference/debugging.html) library:

```sh
$ racket --repl
Welcome to Racket v8.9 [cs].
> (require "2.77-apply-generic-magnitude.rkt")
>(install-complex-package)
> (put 'make-from-real-imag 'complex #<procedure:...neric-magnitude.rkt:75:16>)
< #<void>
<'done
'done
> (make-complex-from-real-imag 3 4)
>(make-complex-from-real-imag 3 4)
> (get 'make-from-real-imag 'complex)
< #<procedure:...neric-magnitude.rkt:75:16>
> (make-from-real-imag 3 4)
> >(get 'make-from-real-imag 'rectangular)
< <#f
application: not a procedure;
 expected a procedure that can be applied to arguments
  given: #f
 [,bt for context]
```

As a result, it became evident that when we called `(make-complex-from-real-imag 3 4)`, it was actually invoking `make-from-real-imag 3 4`,
which, in turn, called `(get 'make-from-real-imag 'rectangular)` and this returned `#f`.

To address this issue, we need to add the `install-rectangular-package` procedure mentioned in [2.4.3  Data-Directed Programming and Additivity](https://mitp-content-server.mit.edu/books/content/sectbyfn/books_pres_0/6515/sicp.zip/full-text/book/book-Z-H-17.html#%_sec_2.4.3)
to get the expected error:

```sh
>(make-complex-from-real-imag 3 4)
> (get 'make-from-real-imag 'complex)
< #<procedure:...neric-magnitude.rkt:75:16>
> (make-from-real-imag 3 4)
> >(get 'make-from-real-imag 'rectangular)
< <#<procedure:...neric-magnitude.rkt:65:16>
> (attach-tag 'rectangular '(3 . 4))
< '(rectangular 3 . 4)
>(tag '(rectangular 3 . 4))
>(attach-tag 'complex '(rectangular 3 . 4))
<'(complex rectangular 3 . 4)
>(get 'magnitude '(complex))
<#f
No method for these types: APPLY-GENERIC '(magnitude (complex))
```