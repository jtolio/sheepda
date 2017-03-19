# sheepda

a memoizing pure lambda calculus interpreter with only one builtin:

 * BYTE_PRINT - takes a church-encoded numeral, turns it into a byte
    internally, and writes it to the configured output stream

grammar:

```
<expr> ::= <variable>
         | `λ`<variable>`.`<expr>
         | `(`<expr> <expr>`)`
```

two forms of syntax sugar are added.

before the first real expression, if

```
<variable> `=` <expr>
<rest>
```

is encountered, it is turned into

```
`(λ`<variable>`.`<rest> <expr>`)`
```

second, if more than one argument is encountered, it is assumed to be curried.
for example, instead of chaining arguments like this:

```
((((f a1) a2) a3) a4)
```

you can instead use the form:

```
(f a1 a2 a3 a4)
```

check out interview-probs for how some abstraction towers can be built up for
solving common whiteboard problems in pure lambda calculus.
