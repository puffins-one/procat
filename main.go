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
	var includeFile string
	var forceFlag bool

	flag.BoolVar(&clipboardFlag, "clipboard", false, "Copy output to the system clipboard.")
	flag.BoolVar(&clipboardFlag, "c", false, "Alias for --clipboard.")

	flag.StringVar(&excludeFlag, "exclude", "", "Comma-separated list of file extensions to exclude (e.g. md,jpg).")
	flag.StringVar(&excludeFlag, "x", "", "Alias for --exclude.")

	flag.StringVar(&includeFile, "include", "", "Path to a file containing patterns to include (e.g. .catinclude).")
	flag.StringVar(&includeFile, "i", "", "Alias for --include.")

	flag.BoolVar(&forceFlag, "force", false, "Process files even if listed in .gitignore (only valid with --include).")
	flag.BoolVar(&forceFlag, "f", false, "Alias for --force.")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintln(os.Stderr, "Usage: procat [options] <project_directory> [output_file]")
		flag.PrintDefaults()
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

	output, err := processProject(projectDir, outputFile, excludedExts, includeFile, forceFlag)
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

func processProject(projectDir, outputFilename string, excludedExts []string, includeFile string, force bool) (string, error) {
	var builder strings.Builder

	// Load .gitignore
	ignore, err := gitignore.NewRepository(projectDir)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("error reading .gitignore: %w", err)
	}

	// Load Include File (if provided)
	var includer gitignore.GitIgnore
	if includeFile != "" {
		// We use NewFromFile. The library usually treats the file location as the root for patterns.
		// If the user provides a relative path like ".catinclude", it works relative to CWD.
		includer, err = gitignore.NewFromFile(includeFile)
		if err != nil {
			return "", fmt.Errorf("error reading include file %s: %w", includeFile, err)
		}
	}

	// Load .catignore (Only used if NOT in include mode)
	var catIgnore gitignore.GitIgnore
	if includer == nil {
		catIgnorePath := filepath.Join(projectDir, ".catignore")
		if _, err := os.Stat(catIgnorePath); err == nil {
			catIgnore, _ = gitignore.NewFromFile(catIgnorePath)
		}
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

		// Always skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if relPath == outputFilename {
			return nil
		}

		// Determine if the file is ignored by git
		isGitIgnored := false
		if ignore != nil {
			if match := ignore.Match(path); match != nil && match.Ignore() {
				isGitIgnored = true
			}
		}

		// --- INCLUDE MODE LOGIC ---
		if includer != nil {
			// In include mode, we generally don't skip directories during traversal,
			// because a subdirectory might contain a whitelisted file even if the directory
			// name doesn't explicitly match the whitelist pattern.
			if info.IsDir() {
				return nil
			}

			// Check if the file is in the include list
			// The library returns Match().Ignore() = true if it matches a pattern in the file.
			isIncluded := false
			if match := includer.Match(path); match != nil && match.Ignore() {
				isIncluded = true
			}

			if !isIncluded {
				return nil // File not whitelisted
			}

			// Conflict Check: Included but GitIgnored
			if isGitIgnored {
				if !force {
					fmt.Fprintf(os.Stderr, "Warning: File '%s' is included via %s but ignored by .gitignore. Skipping. Use --force to include.\n", relPath, includeFile)
					return nil
				}
				// If force is true, we proceed despite gitignore
			}

		} else {
			// --- STANDARD MODE LOGIC ---

			// 1. Check .gitignore
			if isGitIgnored {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 2. Check .catignore
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
		}

		// --- COMMON FILE PROCESSING ---

		// Extension exclusion check
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

		// Binary check
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
