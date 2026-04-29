package main

import (
	"fmt"
	"os"
	"strings"
	"log"

	"github.com/gabehf/koito/engine"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "reset-password" {
		if err := engine.ResetPassword(readEnvOrFile, os.Stdout, Version); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		return
	}
	if err := engine.Run(readEnvOrFile, os.Stdout, Version); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func readEnvOrFile(envName string) string {
	envContent := os.Getenv(envName)

	if envContent == "" {
		filename := os.Getenv(envName + "_FILE")

		if filename != "" {
			b, err := os.ReadFile(filename)

			if err != nil {
				log.Fatalf("Failed to load file for %s_FILE (%s): %s", envName, filename, err)
			}

			envContent = strings.TrimSpace(string(b))
		}
	}

	return envContent
}
