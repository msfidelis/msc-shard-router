package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run generate_uuids.go <quantidade>")
		fmt.Println("Exemplo: go run generate_uuids.go 1000000")
		os.Exit(1)
	}

	var count int
	_, err := fmt.Sscanf(os.Args[1], "%d", &count)
	if err != nil {
		fmt.Printf("Erro ao converter quantidade: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Gerando %d UUIDs...\n", count)

	for i := 0; i < count; i++ {
		fmt.Println(uuid.New().String())
	}

	fmt.Fprintf(os.Stderr, "✅ %d UUIDs gerados com sucesso\n", count)
}
