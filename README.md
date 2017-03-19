# sheepda

A memoizing pure
[lambda calculus](https://en.wikipedia.org/wiki/Lambda_calculus) interpreter
with only one builtin:

 * `BYTE_PRINT` - takes a
    [church-encoded](https://en.wikipedia.org/wiki/Church_encoding) numeral
    and writes the corresponding byte to the configured output stream.

### Examples

Check out
[interview-probs](https://github.com/jtolds/sheepda/tree/master/interview-probs)
for how some abstraction towers can be built up for solving common whiteboard
problems in pure lambda calculus.

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

Before the first real expression, if

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
