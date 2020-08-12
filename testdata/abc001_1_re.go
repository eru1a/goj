package main

import "fmt"

func main() {
	var h, w int
	fmt.Scan(&h, &w)
	var a [1]int
	a[100000000] = 1
	fmt.Println(h - w)
}
