package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: vaultview [flags] <vault-path>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	vaultPath, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalf("invalid vault path: %v", err)
	}

	info, err := os.Stat(vaultPath)
	if err != nil || !info.IsDir() {
		log.Fatalf("vault path %q is not a valid directory", vaultPath)
	}

	log.Printf("serving vault %s on %s", vaultPath, *addr)
	if err := serve(vaultPath, *addr); err != nil {
		log.Fatal(err)
	}
}
