package main

import (
	"fmt"

	"github.com/watzon/goshot/pkg/syntax"
)

func main() {
	config := syntax.DefaultConfig()
	width := config.GetMonospaceWidth(120)
	fmt.Printf("Width needed for 120 characters: %dpx\n", width)
}
