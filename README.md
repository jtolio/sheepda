# sheepda

A memoizing pure
[lambda calculus](https://en.wikipedia.org/wiki/Lambda_calculus) interpreter
with only one builtin:

 * `PRINT_BYTE` - takes a
    [church-encoded](https://en.wikipedia.org/wiki/Church_encoding) numeral
    and writes the corresponding byte to the configured output stream.

### Usage

```
cd $(mktemp -d)
GOPATH=$(pwd) go get github.com/jtolds/sheepda
cat src/github.com/jtolds/sheepda/interview-probs/{prelude,hello-world}.shp |
    bin/sheepda output
```

```
Usage: bin/sheepda [-a] <parsed|output|result>
  -a	if provided, skip assignments when pretty-printing in parsed mode
```

### Examples

Check out
[interview-probs](https://github.com/jtolds/sheepda/tree/master/interview-probs)
for how some abstraction towers can be built up for solving common whiteboard
problems in pure lambda calculus.

Here is Fizzbuzz in pure lambda calculus with only `PRINT_BYTE` added. Much of
the definition is building up supporting structures like lists and numbers and
logic.

```
(λU.(λY.(λvoid.(λ0.(λsucc.(λ+.(λ*.(λ1.(λ2.(λ3.(λ4.(λ5.(λ6.(λ7.(λ8.(λ9.(λ10.
(λnum.(λtrue.(λfalse.(λif.(λnot.(λand.(λor.(λmake-pair.(λpair-first.
(λpair-second.(λzero?.(λpred.(λ-.(λeq?.(λ/.(λ%.(λnil.(λnil?.(λcons.(λcar.
(λcdr.(λdo2.(λdo3.(λdo4.(λfor.(λprint-byte.(λprint-list.(λprint-newline.
(λzero-byte.(λitoa.(λfizzmsg.(λbuzzmsg.(λfizzbuzzmsg.(λfizzbuzz.(fizzbuzz
(((num 1) 0) 1)) λn.((for n) λi.((do2 (((if (zero? ((% i) 3))) λ_.(((if (zero?
((% i) 5))) λ_.(print-list fizzbuzzmsg)) λ_.(print-list fizzmsg))) λ_.(((if
(zero? ((% i) 5))) λ_.(print-list buzzmsg)) λ_.(print-list (itoa i)))))
(print-newline nil)))) ((cons (((num 0) 7) 0)) ((cons (((num 1) 0) 5)) ((cons
(((num 1) 2) 2)) ((cons (((num 1) 2) 2)) ((cons (((num 0) 9) 8)) ((cons (((num
1) 1) 7)) ((cons (((num 1) 2) 2)) ((cons (((num 1) 2) 2)) nil))))))))) ((cons
(((num 0) 6) 6)) ((cons (((num 1) 1) 7)) ((cons (((num 1) 2) 2)) ((cons (((num
1) 2) 2)) nil))))) ((cons (((num 0) 7) 0)) ((cons (((num 1) 0) 5)) ((cons
(((num 1) 2) 2)) ((cons (((num 1) 2) 2)) nil))))) λn.(((Y λrecurse.λn.λresult.
(((if (zero? n)) λ_.(((if (nil? result)) λ_.((cons zero-byte) nil)) λ_.result))
λ_.((recurse ((/ n) 10)) ((cons ((+ zero-byte) ((% n) 10))) result)))) n) nil))
(((num 0) 4) 8)) λ_.(print-byte (((num 0) 1) 0))) (Y λrecurse.λl.(((if (nil?
l)) λ_.void) λ_.((do2 (print-byte (car l))) (recurse (cdr l)))))) PRINT_BYTE)
λn.λf.((((Y λrecurse.λremaining.λcurrent.λf.(((if (zero? remaining)) λ_.void)
λ_.((do2 (f current)) (((recurse (pred remaining)) (succ current)) f)))) n) 0)
f)) λa.do3) λa.do2) λa.λb.b) λl.(pair-second (pair-second l))) λl.(pair-first
(pair-second l))) λe.λl.((make-pair true) ((make-pair e) l))) λl.(not
(pair-first l))) ((make-pair false) void)) λm.λn.((- m) ((* ((/ m) n)) n))) (Y
λ/.λm.λn.(((if ((eq? m) n)) λ_.1) λ_.(((if (zero? ((- m) n))) λ_.0) λ_.((+ 1)
((/ ((- m) n)) n)))))) λm.λn.((and (zero? ((- m) n))) (zero? ((- n) m)))) λm.
λn.((n pred) m)) λn.(((λn.λf.λx.(pair-second ((n λp.((make-pair (f (pair-first
p))) (pair-first p))) ((make-pair x) x))) n) succ) 0)) λn.((n λ_.false) true))
λp.(p false)) λp.(p true)) λx.λy.λt.((t x) y)) λa.λb.((a true) b)) λa.λb.((a b)
false)) λp.λt.λf.((p f) t)) λp.λa.λb.(((p a) b) void)) λt.λf.f) λt.λf.t) λa.λb.
λc.((+ ((+ ((* ((* 10) 10)) a)) ((* 10) b))) c)) (succ 9)) (succ 8)) (succ 7))
(succ 6)) (succ 5)) (succ 4)) (succ 3)) (succ 2)) (succ 1)) (succ 0)) λm.λn.λx.
(m (n x))) λm.λn.λf.λx.((((m succ) n) f) x)) λn.λf.λx.(f ((n f) x))) λf.λx.x)
λx.(U U)) (U λh.λf.(f λx.(((h h) f) x)))) λf.(f f))
```

Formatted and without dependent subproblems:

```
fizzbuzz = λn.
  (for n λi.
    (do2
      (if (zero? (% i 3))
          λ_. (if (zero? (% i 5))
                  λ_. (print-list fizzbuzzmsg)
                  λ_. (print-list fizzmsg))
          λ_. (if (zero? (% i 5))
                  λ_. (print-list buzzmsg)
                  λ_. (print-list (itoa i))))
      (print-newline nil)))
```

### Grammar

```
<expr> ::= <variable>
         | `λ` <variable> `.` <expr>
         | `(` <expr> <expr> `)`
```

Note that the backslash character `\` can be used instead of the lambda
character `λ`.

### Syntax sugar

Two forms of syntax sugar are added:

#### Assignments

Every valid lambda calculus program consists of exactly one expression, but
this doesn't lend itself to easy construction. So, before the main expression,
if

```
<variable> `=` <expr>
<rest>
```

is encountered, it is turned into a function call application to define the
variable, like so:

```
`(` `λ` <variable> `.` <rest> <expr> `)`
```

Example:

```
true = λx.λy.x

(true a b)
```

is turned into

```
(λtrue.(true a b) λx.λy.x)
```

Every valid program is therefore a list of zero or more assignments followed
by a single expression.

#### Currying

If more than one argument is encountered, it is assumed to be
[curried](https://en.wikipedia.org/wiki/Currying).
For example, instead of chaining arguments like this:

```
((((f a1) a2) a3) a4)
```

you can instead use the form:

```
(f a1 a2 a3 a4)
```

### License

Copyright (C) 2017 JT Olds. See LICENSE for copying information.
