# procat

A command-line tool that concatenates all files in a project directory into a single file, respecting `.gitignore` rules. This is useful for feeding an entire project's context into a Large Language Model (LLM).

## Features

*   Recursively scans a directory.
*   Intelligently ignores files and directories listed in `.gitignore`.
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

Examples:

```sh
# Concatenate, skipping markdown and jpg files
procat -x md,jpg

# Concatenate, exclude pngs, copy to clipboard
procat -c -x png 

# Concatenate and output to a file named 'project_context.txt'
procat . project_context.txt

# Concatenate and print to the terminal (standard output)
procat .

# Use a different project directory and copy to clipboard
procat -c /path/to/your/project
```
