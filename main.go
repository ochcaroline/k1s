package main

import (
	"flag"
	"fmt"
	"os"

	"k1s/internal/ui"
)

func main() {
	var namespace string
	var context string

	flag.StringVar(&namespace, "n", "", "namespace (default: from kubeconfig)")
	flag.StringVar(&context, "context", "", "kubeconfig context (default: current)")
	flag.Parse()

	a, err := ui.New(namespace, context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
