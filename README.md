# rum

[![Build Status](https://travis-ci.org/rumlang/rum.svg?branch=master)](https://travis-ci.org/rumlang/rum)
[![Go Report Card](https://goreportcard.com/badge/github.com/rumlang/rum)](https://goreportcard.com/report/github.com/rumlang/rum)
[![Documentation](https://godoc.org/github.com/rumlang/rum?status.svg)](http://godoc.org/github.com/rumlang/rum)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/rumlang/rum/LICENSE)

Free software environment for statistical computing

## Install

```
go install github.com/rumlang/rum
```

## Run

```
rum
```

or

```
go run rum.go
```

## Proposal syntax

```
(package "main"
  ; load file on this code
  (load 'lerolero) ; ./lerolero.gl

  ; import package lerolero and used methods
  (import test 'lerolero.test) ; ./lerolero/test.lg
  (print (test.Test))

  ; set lets
  (let a 1)
  (print (a))

  ; create function
  (def hi()
    'Hello)
  (print hi)

  ; create function (by lambda)
  (let area
    (lambda (r) (* 3.141592653 (* r r))))
  (print (area 10.0))

  ; create loops with for
  (for (print 'hello)) ; will loop while expression or function return false
  (for area (10.0 20.0 30.0) ; will interact on the list elements
  (for (let a 1)
    (= a 10)
      (let a (+ a 1)
      (print a)))

  ; create if
  (if (= a 10)        ; if a is equal to 10
    (print 'Hello))   ; print Hello

  (if (= a 'good)     ; if a is equal to "good"
    (print 'good)     ; print "good"
    (else             ; otherwise
      (print 'bad)))  ; print "bad"

  (if (= a 'good)     ; if a is equal to "good"
    (print 'good)     ; print "good"
    (else (= a 'bad)  ; if not and is equal to "bad"
      (print 'bad))   ; print "bad"
    (else             ; otherwise
      (print 'ugly))) ; print "ugly"

```

## Using rum as a Go package

### Install

```
go get github.com/rumlang/rum
```

### Example

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
		panic(err)
	}
}
```
