package main

import (
	"fmt"

	"github.com/google/uuid"
)

func generateUUIDs() {
	for i := 0; i < 10000; i++ {
		fmt.Println(uuid.New().String())
	}
}

func main() {
	generateUUIDs()
}
