# procat

A command-line tool that concatenates all files in a project directory into a single file, respecting `.gitignore` rules. This is useful for feeding an entire project's context into a Large Language Model (LLM).

## Features

*   Recursively scans a directory.
*   Intelligently ignores files and directories listed in `.gitignore`.
*   **Supports `.catignore`**: Use a specific `.catignore` file to exclude files from concatenation without ignoring them in Git (e.g., `pnpm-lock.json`, `package-lock.json`).
*   **Supports `.catinclude`**: Use a whitelist mode to specific exactly which files to include (bypassing `.catignore` but respecting `.gitignore` unless forced).
*   Skips binary files to keep the output clean.
*   Wraps each file's content with clear `Start` and `End` markers.
*   Outputs to a file, to standard output, or directly to the system clipboard.

## Installation

To install `procat`, you need to have the Go toolchain installed on your system. Then, you can install the tool with a single command, which will download the source code, compile it, and place the executable in your Go binary path.

```sh
go install github.com/puffins-one/procat@latest
```
*(**Note:** Make sure your `GOPATH/bin` directory is in your shell's `PATH` environment variable to run the command from anywhere.)*

## Usage

Usage: procat [options] <project_directory> [output_file]

Options:
  -c, --clipboard       Copy output to the system clipboard.
  -x, --exclude <exts>  Comma-separated list of extensions to exclude (e.g. "md,jpg").
  -i, --include <file>  Path to a file (e.g., .catinclude) specifying files to include.
                        Only files matching patterns in this file will be processed.
                        This overrides .catignore.
  -f, --force           Force inclusion of files even if they are listed in .gitignore.
                        (Only applies when using --include).

### The .catignore file

You can place a `.catignore` file in the root of your project directory. This works exactly like `.gitignore`, but it only affects `procat`.

This is useful for files you want to keep in version control but don't want to feed to an LLM (like massive lock files or specific documentation).

Example `.catignore`:
```text
pnpm-lock.json
package-lock.json
*.svg
docs/
```

### The .catinclude file

You can create a file (e.g., `.catinclude`) to act as a whitelist. If used, `procat` will **only** include files matching patterns in this file.

Example `.catinclude`:
```text
src/
README.md
!src/legacy_code.go
```

**Note:** If a file is in `.catinclude` but also in `.gitignore`, `procat` will warn you and skip it to prevent accidental leakage of secrets. Use `-f` / `--force` to override this.

### Examples

```sh
# Standard usage: Concatenate everything not in .gitignore/.catignore
procat .

# Concatenate only files matching patterns in .catinclude
procat -i .catinclude .

# Concatenate files in .catinclude, even if they are gitignored
procat -i .catinclude --force .

# Concatenate, skipping markdown and jpg files
procat -x md,jpg .

# Concatenate and output to a file named 'project_context.txt'
procat . project_context.txt
```