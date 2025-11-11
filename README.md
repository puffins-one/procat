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

Usage: procat [--clipboard | -c] <project_directory> [output_file]

```sh
# Concatenate the current directory and copy the result to your clipboard
procat . --clipboard

# You can also use the short alias -c
procat . -c

# Concatenate and output to a file named 'project_context.txt'
procat . project_context.txt

# Concatenate and print to the terminal (standard output)
procat .

# Use a different project directory and copy to clipboard
procat /path/to/your/project --clipboard
```
