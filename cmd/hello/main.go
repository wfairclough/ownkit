package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello from ownkit!")
	fmt.Printf("Args: %v\n", os.Args[1:])
}
