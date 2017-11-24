# gin (Go in numbers)

[![Build Status](https://travis-ci.org/gin-lang/gin.svg?branch=master)](https://travis-ci.org/gin-lang/gin)
[![Go Report Card](https://goreportcard.com/badge/github.com/gin-lang/gin)](https://goreportcard.com/report/github.com/gin-lang/gin)
[![Documentation](https://godoc.org/github.com/gin-lang/gin?status.svg)](http://godoc.org/github.com/gin-lang/gin)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/gin-lang/gin/LICENSE)

Free software environment for statistical computing

## Install

```
go install github.com/gin-lang/gin
```

## Run

```
gin
```

or

```
go run gin.go
```

## Proposal syntax

```
(package "main"
  ; load file on this code
  (load 'lerolero) ; ./lerolero.gl

  ; import package lerolero and used methods
  (import test 'lerolero.test) ; ./lerolero/test.lg
  (print (test.Test))

  ; set vars
  (var a 1)
  (print (a))

  ; create function
  (def hi()
    'Hello)
  (print hi)

  ; create function (by lambda)
  (var area
    (lambda (r) (* 3.141592653 (* r r))))
  (print (area 10.0))

  ; create loops with for
  (for (print 'hello)) ; will loop while expression or function return false
  (for area (10.0 20.0 30.0) ; will interact on the list elements
  (for (var a 1)
    (= a 10)
      (var a (+ a 1)
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

## Using gin as a Go package

### Install

```
go get github.com/gin-lang/gin
```

### Example

```golang
package main

import (
	"bufio"
	"strings"

	"github.com/gin-lang/gin"
)

func main() {
	const input = `
(package main
  (print 'Hello)
)
`
	s := bufio.NewScanner(strings.NewReader(input))
	vm := gin.New()
	err := vm.Run(s)
	if err != nil {
		panic(err)
	}
}
```
