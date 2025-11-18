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
	errors := validateYAML(data, filename)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	fmt.Println("YAML is valid!")
}