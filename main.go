package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/denormal/go-gitignore"
)

func main() {
	var clipboardFlag bool
	flag.BoolVar(&clipboardFlag, "clipboard", false, "Copy output to the system clipboard.")
	flag.BoolVar(&clipboardFlag, "c", false, "Alias for --clipboard.")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintln(os.Stderr, "Usage: procat [--clipboard | -c] <project_directory> [output_file]")
		os.Exit(1)
	}

	projectDir := args[0]
	var outputFile string
	if len(args) == 2 {
		outputFile = args[1]
	}

	if clipboardFlag && outputFile != "" {
		fmt.Fprintln(os.Stderr, "Warning: --clipboard flag is set; output file argument will be ignored.")
		outputFile = ""
	}

	output, err := processProject(projectDir, outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing project: %v\n", err)
		os.Exit(1)
	}

	if clipboardFlag {
		if err := clipboard.WriteAll(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to copy to clipboard: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project content copied to clipboard!")
	} else if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(output)
	}
}

func processProject(projectDir, outputFilename string) (string, error) {
	var builder strings.Builder

	ignore, err := gitignore.NewRepository(projectDir)
	if err != nil && !os.IsNotExist(err) {
		// A non-existent .gitignore is fine
		return "", fmt.Errorf("error reading .gitignore: %w", err)
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory
		if path == projectDir {
			return nil
		}

		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return err
		}

		// Skip .git directory and output file
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if relPath == outputFilename {
			return nil
		}

		// Gitignore matching
		if ignore != nil {
			if match := ignore.Match(path); match != nil && match.Ignore() {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping unreadable file %s: %v\n", path, err)
			return nil
		}

		for _, b := range content {
			if b == 0 {
				fmt.Fprintf(&builder, "// Skipping binary file: %s\n\n", relPath)
				return nil
			}
		}

		fmt.Fprintf(&builder, "// Start %s\n", relPath)
		builder.Write(content)
		fmt.Fprintf(&builder, "\n// End %s\n\n", relPath)

		return nil
	}

	if err := filepath.Walk(projectDir, walkFn); err != nil {
		return "", fmt.Errorf("error walking path %q: %w", projectDir, err)
	}

	return builder.String(), nil
}