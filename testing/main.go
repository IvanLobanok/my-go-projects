package main

import "fmt"

func main() {
	a := make(map[string]int)
	for i := 1; i < 6; i++ {
		var b string
		_, err := fmt.Scanln(&b)
		if err != nil {
		}

		a[b]++
	}
	fmt.Println(a)
}
