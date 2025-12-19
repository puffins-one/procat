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
	var excludeFlag string

	flag.BoolVar(&clipboardFlag, "clipboard", false, "Copy output to the system clipboard.")
	flag.BoolVar(&clipboardFlag, "c", false, "Alias for --clipboard.")

	flag.StringVar(&excludeFlag, "exclude", "", "Comma-separated list of file extensions to exclude (e.g. md,jpg).")
	flag.StringVar(&excludeFlag, "x", "", "Alias for --exclude.")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintln(os.Stderr, "Usage: procat [--clipboard | -c] [--exclude | -x md,jpg] <project_directory> [output_file]")
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

	var excludedExts []string
	if excludeFlag != "" {
		parts := strings.Split(excludeFlag, ",")
		for _, part := range parts {
			ext := strings.TrimSpace(part)
			if ext == "" {
				continue
			}
			ext = strings.ToLower(ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			excludedExts = append(excludedExts, ext)
		}
	}

	output, err := processProject(projectDir, outputFile, excludedExts)
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

func processProject(projectDir, outputFilename string, excludedExts []string) (string, error) {
	var builder strings.Builder

	ignore, err := gitignore.NewRepository(projectDir)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("error reading .gitignore: %w", err)
	}

	var catIgnore gitignore.GitIgnore
	catIgnorePath := filepath.Join(projectDir, ".catignore")
	if _, err := os.Stat(catIgnorePath); err == nil {
		catIgnore, _ = gitignore.NewFromFile(catIgnorePath)
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == projectDir {
			return nil
		}

		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if relPath == outputFilename {
			return nil
		}

		if ignore != nil {
			if match := ignore.Match(path); match != nil && match.Ignore() {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if catIgnore != nil {
			if match := catIgnore.Match(path); match != nil && match.Ignore() {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

		if len(excludedExts) > 0 {
			fileExt := strings.ToLower(filepath.Ext(path))
			for _, excluded := range excludedExts {
				if fileExt == excluded {
					return nil
				}
			}
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
