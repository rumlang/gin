# Rum Language - LISP dialect

![Github Actions](https://github.com/rumlang/rum/actions/workflows/tests.yml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/rumlang/rum)](https://goreportcard.com/report/github.com/rumlang/rum)
[![Documentation](https://godoc.org/github.com/rumlang/rum?status.svg)](http://godoc.org/github.com/rumlang/rum)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/rumlang/rum/LICENSE)

Functional language, easily extensible and possible (Lua features with LISP dialect and functional) to be embarked on software Go!

## History

Idealized in GopherCon Brasil 2017 (reason of the language being written in Go), had the purpose of being virtual machine of CLISP (development for fun and the founder team enjoys functional programming), after seeing that there was already some parser of CLISP written in Go we had a initiative to make a language with syntax like CLISP but with some own paradigms (such as asynchronous processing using goroutine underneath, thus joining what we have best in the Go).

### Why Rum?

As the language was born in Canasvieiras (Florianópolis - Brazil) neighborhood in the seaside frequented by tourists having the pirate boat as a tourist attraction, we decided to give the name of the typical beverage of pirates for the language.

**[Why another lisp?](https://github.com/rumlang/rum/issues/104)**

## Install

```sh
go install github.com/rumlang/rum
```

## Run

```sh
./rum
```

or

```sh
go run rum.go
```

## Example

```clojure
(package "main"
  (import
    (str "strings")
    csv)

  ; Canonical example
  (println "Hello, World!")

  ; Create a function, run it and print the result.
  (let area
    (lambda (r) (* 3.141592653 (* r r))))
  (println (area 10.0))

  (def area(r)
    (* 3.141592653 (* r r)))
  (println (area 100.0))

  ;; use strings package with alias str
  (println (str.contains "rumlang" "rum"))
  ;; use csv package by example
  (println "csv read all:" (. (csv.new-reader (str.new-reader "1,2,3,4")) read-all))
)
```


### Using rum as a Go package

```golang
package main

import (
  "bufio"
  "strings"

  "github.com/rumlang/rum"
)

func main() {
  const input = `
(package main
  (print 'Hello)
)
`
  s := bufio.NewScanner(strings.NewReader(input))
  vm := rum.New()
  err := vm.Run(s)
  if err != nil {
    fmt.Println(err)
  }
}
```
