# Go Terminal Text Editor

## Description

This project is a lightweight, terminal-based text editor written in Go. It provides a vim-like interface with basic text editing capabilities, search functionality, and file management features. The editor is designed to be efficient and easy to use, making it suitable for quick edits and working in terminal environments.

## Features

- Vim-like modal editing (Normal, Insert, Search modes)
- Basic text manipulation (insert, delete, copy, paste)
- File operations (open, save)
- Search functionality with highlighting
- Undo/Redo capabilities
- Line numbering
- Status bar with file and cursor information

## Installation

### Prerequisites

- Go 1.16 or higher
- Git

### Steps

1. Clone the repository:
   ```
   git clone https://github.com/IvanGael/Go-terminal-editor.git
   cd go-terminal-editor
   ```

2. Install dependencies:
   ```
   go get
   ```

3. Build the project:
   ```
   go build -o editor
   ```

## Usage

To start the editor, run:

```
./editor [filename]
```

If a filename is provided, the editor will open that file. Otherwise, it will start with a blank document.

### Key Bindings

#### Normal Mode
- `i`: Enter Insert mode
- `h`, `j`, `k`, `l` or arrow keys: Move cursor
- `x`: Delete character under cursor
- `dd`: Delete current line
- `yy`: Yank (copy) current line
- `p`: Paste yanked or deleted content
- `/`: Enter Search mode
- `n`: Find next occurrence
- `N`: Find previous occurrence
- `u`: Undo
- `Ctrl+r`: Redo
- `:w`: Save file
- `:q`: Quit (will warn if unsaved changes)
- `:q!`: Force quit without saving

#### Insert Mode
- `Esc`: Return to Normal mode
- Any character: Insert at cursor position
- `Enter`: Insert new line
- `Backspace`: Delete character before cursor

#### Search Mode
- Type to enter search term
- `Enter`: Confirm search and return to Normal mode
- `Esc`: Cancel search and return to Normal mode

## Configuration

The editor uses some default settings that can be modified in the source code:

- Tab size: 4 spaces (adjustable in the `initialModel` function)
- Color scheme: Can be modified by changing the ANSI color codes in the `View` function

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## Acknowledgements

This project uses the following open-source libraries:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for terminal UI
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) for styling
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) for terminal handling
