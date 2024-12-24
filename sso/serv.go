package main

import (
	"fmt"
)

func main() {
	var ste = "string of bytes!!!"
	bs := []byte(ste)
	for _, v := range bs {
		fmt.Print(string(v))
	}
}
