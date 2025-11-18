package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: yamlvalid <path-to-yaml-file>")
		os.Exit(1)
	}

	filename := os.Args[1]
	
	// Чтение файла
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Валидация YAML
	if err := validateYAML(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("YAML is valid!")
}