package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type mode int

const (
	normalMode mode = iota
	insertMode
	searchMode
	replaceMode
)

type action struct {
	content [][]rune
	cursorX int
	cursorY int
}

type model struct {
	content     [][]rune
	cursorX     int
	cursorY     int
	offsetY     int
	width       int
	height      int
	mode        mode
	filename    string
	statusMsg   string
	searchTerm  string
	replaceTerm string
	clipboard   string
	modified    bool
	tabSize     int
	undoStack   []action
	redoStack   []action
}

func initialModel(filename string) model {
	content := [][]rune{{}}
	if filename != "" {
		if data, err := os.ReadFile(filename); err == nil {
			lines := strings.Split(string(data), "\n")
			content = make([][]rune, len(lines))
			for i, line := range lines {
				content[i] = []rune(line)
			}
		}
	}
	return model{
		content:   content,
		cursorX:   0,
		cursorY:   0,
		offsetY:   0,
		mode:      normalMode,
		filename:  filename,
		statusMsg: "Normal mode",
		tabSize:   4,
	}
}

func (m model) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case normalMode:
			return m.handleNormalMode(msg)
		case insertMode:
			return m.handleInsertMode(msg)
		case searchMode:
			return m.handleSearchMode(msg)
		case replaceMode:
			return m.handleReplaceMode(msg)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 2 // Reserve 2 lines for status bar
	}
	return m, nil
}

func (m *model) saveAction() {
	m.undoStack = append(m.undoStack, action{
		content: deepCopyContent(m.content),
		cursorX: m.cursorX,
		cursorY: m.cursorY,
	})
	m.redoStack = nil // Clear redo stack when a new action is performed
}

func (m *model) undo() {
	if len(m.undoStack) > 0 {
		// Save current state to redo stack
		m.redoStack = append(m.redoStack, action{
			content: deepCopyContent(m.content),
			cursorX: m.cursorX,
			cursorY: m.cursorY,
		})

		// Pop the last action from undo stack
		lastAction := m.undoStack[len(m.undoStack)-1]
		m.undoStack = m.undoStack[:len(m.undoStack)-1]

		// Apply the last action
		m.content = deepCopyContent(lastAction.content)
		m.cursorX = lastAction.cursorX
		m.cursorY = lastAction.cursorY

		m.modified = true
		m.statusMsg = "Undo performed"
	} else {
		m.statusMsg = "Nothing to undo"
	}
}

func (m *model) redo() {
	if len(m.redoStack) > 0 {
		// Save current state to undo stack
		m.undoStack = append(m.undoStack, action{
			content: deepCopyContent(m.content),
			cursorX: m.cursorX,
			cursorY: m.cursorY,
		})

		// Pop the last action from redo stack
		lastAction := m.redoStack[len(m.redoStack)-1]
		m.redoStack = m.redoStack[:len(m.redoStack)-1]

		// Apply the last action
		m.content = deepCopyContent(lastAction.content)
		m.cursorX = lastAction.cursorX
		m.cursorY = lastAction.cursorY

		m.modified = true
		m.statusMsg = "Redo performed"
	} else {
		m.statusMsg = "Nothing to redo"
	}
}

func deepCopyContent(content [][]rune) [][]rune {
	newContent := make([][]rune, len(content))
	for i, line := range content {
		newContent[i] = make([]rune, len(line))
		copy(newContent[i], line)
	}
	return newContent
}

func (m model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if m.modified {
			m.statusMsg = "Unsaved changes. Use :q! to force quit."
		} else {
			return m, tea.Sequence(tea.ClearScreen, tea.Quit)
		}
	case "i":
		m.mode = insertMode
		m.statusMsg = "Insert mode"
	case "h", "left":
		m.moveCursor(-1, 0)
	case "l", "right":
		m.moveCursor(1, 0)
	case "k", "up":
		m.moveCursor(0, -1)
	case "j", "down":
		m.moveCursor(0, 1)
	case "g":
		m.cursorY = 0
		m.offsetY = 0
	case "G":
		m.cursorY = len(m.content) - 1
		m.adjustOffset()
	case "0":
		m.cursorX = 0
	case "$":
		m.cursorX = len(m.content[m.cursorY])
	case "x":
		if m.cursorX < len(m.content[m.cursorY]) {
			m.content[m.cursorY] = append(m.content[m.cursorY][:m.cursorX], m.content[m.cursorY][m.cursorX+1:]...)
			m.modified = true
		}
	case "d":
		if m.cursorY < len(m.content)-1 {
			m.clipboard = string(m.content[m.cursorY])
			m.content = append(m.content[:m.cursorY], m.content[m.cursorY+1:]...)
			m.modified = true
			if m.cursorY >= len(m.content) {
				m.cursorY = len(m.content) - 1
			}
		}
	case "u":
		m.undo()
	case "ctrl+r":
		m.redo()
	case "y":
		if m.cursorY < len(m.content) {
			m.clipboard = string(m.content[m.cursorY])
			m.statusMsg = "Line yanked to clipboard"
		}
	case "p":
		if m.clipboard != "" {
			m.saveAction() // Save current state for undo
			m.content = append(m.content[:m.cursorY+1], m.content[m.cursorY:]...)
			m.content[m.cursorY+1] = []rune(m.clipboard)
			m.cursorY++
			m.modified = true
			m.statusMsg = "Line pasted from clipboard"
		}
	case "/":
		m.mode = searchMode
		m.statusMsg = "/"
		m.searchTerm = ""
	case "n":
		m.findNext()
	case "N":
		m.findPrevious()
	case ":":
		m.statusMsg = ":"
	case "w":
		if m.statusMsg == ":" {
			m.saveFile()
		}
	case "q!":
		if m.statusMsg == ":" {
			return m, tea.Sequence(tea.ClearScreen, tea.Quit)
		}
	case "ctrl+c":
		return m, tea.Sequence(tea.ClearScreen, tea.Quit)
	case "pageup":
		m.moveCursor(0, -m.height)
	case "pagedown":
		m.moveCursor(0, m.height)
	}
	return m, nil
}

func (m model) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = normalMode
		m.statusMsg = "Normal mode"
		if m.cursorX > 0 {
			m.cursorX--
		}
	case "enter":
		m.saveAction() // Save current state for undo
		newLine := append([]rune{}, m.content[m.cursorY][m.cursorX:]...)
		m.content[m.cursorY] = m.content[m.cursorY][:m.cursorX]
		m.content = append(m.content[:m.cursorY+1], append([][]rune{newLine}, m.content[m.cursorY+1:]...)...)
		m.cursorY++
		m.cursorX = 0
		m.modified = true
	case "backspace":
		if m.cursorX > 0 {
			m.content[m.cursorY] = append(m.content[m.cursorY][:m.cursorX-1], m.content[m.cursorY][m.cursorX:]...)
			m.cursorX--
			m.modified = true
		} else if m.cursorY > 0 {
			m.cursorY--
			m.cursorX = len(m.content[m.cursorY])
			m.content[m.cursorY] = append(m.content[m.cursorY], m.content[m.cursorY+1]...)
			m.content = append(m.content[:m.cursorY+1], m.content[m.cursorY+2:]...)
			m.modified = true
		}
	case "tab":
		for i := 0; i < m.tabSize; i++ {
			m.content[m.cursorY] = append(m.content[m.cursorY][:m.cursorX], append([]rune{' '}, m.content[m.cursorY][m.cursorX:]...)...)
			m.cursorX++
		}
		m.modified = true
	default:
		if len(msg.Runes) == 1 {
			m.saveAction() // Save current state for undo
			m.content[m.cursorY] = append(m.content[m.cursorY][:m.cursorX], append([]rune{msg.Runes[0]}, m.content[m.cursorY][m.cursorX:]...)...)
			m.cursorX++
			m.modified = true
		}
	}
	m.adjustOffset()
	return m, nil
}

func (m model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = normalMode
		m.statusMsg = "Normal mode"
	case "enter":
		m.findNext()
		m.mode = normalMode
	case "backspace":
		if len(m.searchTerm) > 0 {
			m.searchTerm = m.searchTerm[:len(m.searchTerm)-1]
			m.statusMsg = "/" + m.searchTerm
		}
	default:
		if len(msg.Runes) == 1 {
			m.searchTerm += string(msg.Runes[0])
			m.statusMsg = "/" + m.searchTerm
		}
	}
	return m, nil
}

func (m model) handleReplaceMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = normalMode
		m.statusMsg = "Normal mode"
	case "enter":
		m.replaceAll()
		m.mode = normalMode
	case "backspace":
		if len(m.replaceTerm) > 0 {
			m.replaceTerm = m.replaceTerm[:len(m.replaceTerm)-1]
			m.statusMsg = "Replace with: " + m.replaceTerm
		}
	default:
		if len(msg.Runes) == 1 {
			m.replaceTerm += string(msg.Runes[0])
			m.statusMsg = "Replace with: " + m.replaceTerm
		}
	}
	return m, nil
}

func (m *model) moveCursor(dx, dy int) {
	m.cursorX += dx
	m.cursorY += dy

	if m.cursorY < 0 {
		m.cursorY = 0
	} else if m.cursorY >= len(m.content) {
		m.cursorY = len(m.content) - 1
	}

	if m.cursorX < 0 {
		m.cursorX = 0
	} else if m.cursorX > len(m.content[m.cursorY]) {
		m.cursorX = len(m.content[m.cursorY])
	}

	m.adjustOffset()
}

func (m *model) adjustOffset() {
	if m.cursorY < m.offsetY {
		m.offsetY = m.cursorY
	} else if m.cursorY >= m.offsetY+m.height {
		m.offsetY = m.cursorY - m.height + 1
	}
}

func (m *model) findNext() {
	startY, startX := m.cursorY, m.cursorX+1
	for y := startY; y < len(m.content); y++ {
		x := strings.Index(string(m.content[y][startX:]), m.searchTerm)
		if x != -1 {
			m.cursorY = y
			m.cursorX = startX + x
			m.adjustOffset()
			return
		}
		startX = 0
	}
	m.statusMsg = "Pattern not found: " + m.searchTerm
}

func (m *model) findPrevious() {
	startY, startX := m.cursorY, m.cursorX-1
	for y := startY; y >= 0; y-- {
		if startX < 0 {
			startX = len(m.content[y]) - 1
		}
		x := strings.LastIndex(string(m.content[y][:startX+1]), m.searchTerm)
		if x != -1 {
			m.cursorY = y
			m.cursorX = x
			m.adjustOffset()
			return
		}
		startX = -1
	}
	m.statusMsg = "Pattern not found: " + m.searchTerm
}

func (m *model) replaceAll() {
	count := 0
	for y := range m.content {
		line := string(m.content[y])
		newLine := strings.ReplaceAll(line, m.searchTerm, m.replaceTerm)
		if newLine != line {
			m.content[y] = []rune(newLine)
			count += strings.Count(line, m.searchTerm)
			m.modified = true
		}
	}
	m.statusMsg = fmt.Sprintf("Replaced %d occurrences", count)
}

func (m *model) saveFile() {
	content := ""
	for _, line := range m.content {
		content += string(line) + "\n"
	}
	if m.filename != "" {
		err := os.WriteFile(m.filename, []byte(content), 0644)
		if err != nil {
			m.statusMsg = "Error saving file: " + err.Error()
		} else {
			m.statusMsg = "File saved successfully"
			m.modified = false
		}
	} else {
		err := os.WriteFile("samples/output.txt", []byte(content), 0644)
		if err != nil {
			m.statusMsg = "Error saving file: " + err.Error()
		} else {
			m.statusMsg = "File saved successfully"
			m.modified = false
		}
	}
}

func (m model) View() string {
	var s strings.Builder

	// Ensure content is never empty
	if len(m.content) == 0 {
		m.content = [][]rune{{}}
	}

	// Ensure cursor is within bounds
	if m.cursorY >= len(m.content) {
		m.cursorY = len(m.content) - 1
	}
	if m.cursorX > len(m.content[m.cursorY]) {
		m.cursorX = len(m.content[m.cursorY])
	}

	// Content area
	for i := 0; i < m.height; i++ {
		lineNum := m.offsetY + i
		if lineNum < len(m.content) {
			line := m.content[lineNum]
			lineStr := expandTabs(string(line), m.tabSize)

			// Apply search highlighting
			if m.searchTerm != "" {
				lineStr = highlightSearch(lineStr, m.searchTerm)
			}

			if lineNum == m.cursorY && m.mode != normalMode {
				cursorRune := '|'
				if m.cursorX < len(lineStr) {
					cursorRune = rune(lineStr[m.cursorX])
				}
				if m.cursorX < len(lineStr) {
					lineStr = lineStr[:m.cursorX] + string(cursorRune) + lineStr[m.cursorX+1:]
				} else {
					lineStr += string(cursorRune)
				}
			}
			s.WriteString(fmt.Sprintf("%4d %s\n", lineNum+1, lineStr))
		} else {
			s.WriteString("~\n")
		}
	}

	// Status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("57"))

	modeInfo := fmt.Sprintf("%s", m.mode)
	fileInfo := fmt.Sprintf("%-20s", m.filename)
	cursorInfo := fmt.Sprintf("(%d,%d)", m.cursorY+1, m.cursorX+1)
	modifiedInfo := ""
	if m.modified {
		modifiedInfo = "[+]"
	}
	statusBar := statusStyle.Render(fmt.Sprintf("%s %s %s %s %s", modeInfo, m.statusMsg, fileInfo, cursorInfo, modifiedInfo))

	s.WriteString(statusBar)

	// s.WriteString(statusBar + "\n")
	// s.WriteString(m.statusMsg)

	return s.String()
}

func highlightSearch(text, searchTerm string) string {
	if searchTerm == "" {
		return text
	}

	highlightStyle := "\033[43m%s\033[0m" // Yellow background
	parts := strings.Split(text, searchTerm)
	for i := 0; i < len(parts)-1; i++ {
		parts[i] += fmt.Sprintf(highlightStyle, searchTerm)
	}
	return strings.Join(parts, "")
}

func expandTabs(s string, tabSize int) string {
	var result strings.Builder
	column := 0
	for _, r := range s {
		if r == '\t' {
			spaces := tabSize - (column % tabSize)
			result.WriteString(strings.Repeat(" ", spaces))
			column += spaces
		} else {
			result.WriteRune(r)
			column++
		}
	}
	return result.String()
}

func main() {
	filename := ""
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Failed to set terminal to raw mode:", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	p := tea.NewProgram(initialModel(filename), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
