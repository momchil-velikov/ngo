package main

var a, b = f.X, f.Y

type S struct {
	X [len(b)]int
}

type P [1]S

type Q [1]int

type F struct {
	X P
	Y Q
}

var f F
