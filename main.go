package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/timer"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

const (
	timeout      = time.Minute
	visibleLines = 3
	maxExtra     = 8
	hint         = "tab + enter  restart   |   esc  quit"
)

var (
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Green)
	wrongStyle   = lipgloss.NewStyle().Foreground(lipgloss.Red)
	pendingStyle = lipgloss.NewStyle().Faint(true)
	cursorStyle  = lipgloss.NewStyle().Underline(true).UnderlineColor(lipgloss.Yellow)
	timerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Yellow).Bold(true)
	hintStyle    = lipgloss.NewStyle().Faint(true)
	bigStyle     = lipgloss.NewStyle().Foreground(lipgloss.Yellow).Bold(true)
	labelStyle   = lipgloss.NewStyle().Faint(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Yellow).
			Padding(1, 4)
)

func randomText() []string {
	text := make([]string, 300)

	for i := range text {
		text[i] = words[rand.Intn(len(words))]
	}

	return text
}

func matches(a, b string) int {
	aRunes := []rune(a)
	bRunes := []rune(b)

	count := 0

	for i := 0; i < min(len(aRunes), len(bRunes)); i++ {
		if aRunes[i] == bRunes[i] {
			count++
		}
	}

	return count
}

type model struct {
	text       []string
	timer      timer.Model
	started    bool
	finished   bool
	tabPressed bool

	width  int
	height int

	wordIndex  int
	typedWords []string
	typedChars string

	totalKeys   int
	correctKeys int
}

func initialModel() model {
	return model{
		text:  randomText(),
		timer: timer.New(timeout),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) stats() (float64, float64) {
	correct := matches(m.typedChars, m.text[m.wordIndex])

	for i, word := range m.typedWords {
		if word == m.text[i] {
			correct += len([]rune(word)) + 1
		}
	}

	wpm := float64(correct) / 5 / timeout.Minutes()

	var accuracy float64
	if m.totalKeys > 0 {
		accuracy = float64(m.correctKeys) / float64(m.totalKeys) * 100
	}

	return wpm, accuracy
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case timer.TickMsg, timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		if msg.ID == m.timer.ID() {
			m.finished = true
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		key := msg.String()

		if key == "ctrl+c" || key == "esc" {
			return m, tea.Quit
		}

		if key == "tab" {
			m.tabPressed = true
			return m, nil
		}

		if key == "enter" {
			if m.tabPressed {
				newModel := initialModel()
				newModel.width = m.width
				newModel.height = m.height
				return newModel, nil
			}

			return m, nil
		}

		m.tabPressed = false

		if m.finished {
			return m, nil
		}

		var cmd tea.Cmd

		if !m.started {
			m.started = true
			cmd = m.timer.Init()
		}

		switch key {

		case "space":
			if m.typedChars != "" && m.wordIndex < len(m.text)-1 {
				m.typedWords = append(m.typedWords, m.typedChars)
				m.typedChars = ""
				m.wordIndex++
			}

		case "backspace":
			runes := []rune(m.typedChars)

			if len(runes) > 0 {
				m.typedChars = string(runes[:len(runes)-1])
			}

		default:
			word := []rune(m.text[m.wordIndex])
			position := len([]rune(m.typedChars))

			if msg.Text == "" || position >= len(word)+maxExtra {
				return m, cmd
			}

			m.totalKeys++

			if position < len(word) && string(word[position]) == msg.Text {
				m.correctKeys++
			}

			m.typedChars += msg.Text
		}

		return m, cmd
	}

	return m, nil
}

func renderWord(word, typed string, active bool) string {
	wordRunes := []rune(word)
	typedRunes := []rune(typed)

	var result strings.Builder

	for i, char := range wordRunes {
		style := pendingStyle

		if i < len(typedRunes) {
			if typedRunes[i] == char {
				style = correctStyle
			} else {
				style = wrongStyle
			}
		} else if active && i == len(typedRunes) {
			style = cursorStyle
		}

		result.WriteString(style.Render(string(char)))
	}

	if len(typedRunes) > len(wordRunes) {
		extra := string(typedRunes[len(wordRunes):])
		result.WriteString(wrongStyle.Render(extra))
	}

	return result.String()
}

func layoutLines(text []string, width int) [][2]int {
	var lines [][2]int

	start := 0
	currentWidth := 0

	for i, word := range text {
		wordLength := len([]rune(word))

		if currentWidth == 0 {
			currentWidth = wordLength
			continue
		}

		if currentWidth+1+wordLength > width {
			lines = append(lines, [2]int{start, i})
			start = i
			currentWidth = wordLength
		} else {
			currentWidth += 1 + wordLength
		}
	}

	if start < len(text) {
		lines = append(lines, [2]int{start, len(text)})
	}

	return lines
}

func stat(label, value string, style lipgloss.Style) string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		style.Render(value),
		labelStyle.Render(label),
	)
}

func (m model) resultsView() string {
	wpm, accuracy := m.stats()
	wrong := m.totalKeys - m.correctKeys

	gap := "     "

	headline := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stat("wpm", fmt.Sprintf("%.0f", wpm), bigStyle),
		gap,
		stat("accuracy", fmt.Sprintf("%.0f%%", accuracy), bigStyle),
	)

	details := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stat("typed", fmt.Sprint(m.totalKeys), boldStyle),
		gap,
		stat("correct", fmt.Sprint(m.correctKeys), correctStyle),
		gap,
		stat("wrong", fmt.Sprint(wrong), wrongStyle),
	)

	box := boxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			headline,
			"",
			details,
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		box,
		"",
		hintStyle.Render(hint),
	)
}

func (m model) typingView() string {
	lineWidth := max(min(m.width-10, 100), 20)
	lines := layoutLines(m.text, lineWidth)

	currentLine := len(lines) - 1

	for i, line := range lines {
		if m.wordIndex >= line[0] && m.wordIndex < line[1] {
			currentLine = i
			break
		}
	}

	startLine := max(0, currentLine-1)
	endLine := min(len(lines), startLine+visibleLines)

	rowStyle := lipgloss.NewStyle().Width(lineWidth + maxExtra)

	var rows []string

	for _, line := range lines[startLine:endLine] {
		var words []string

		for i := line[0]; i < line[1]; i++ {
			switch {
			case i < m.wordIndex:
				words = append(words, renderWord(m.text[i], m.typedWords[i], false))

			case i == m.wordIndex:
				words = append(words, renderWord(m.text[i], m.typedChars, true))

			default:
				words = append(words, pendingStyle.Render(m.text[i]))
			}
		}

		rows = append(rows, rowStyle.Render(strings.Join(words, " ")))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		timerStyle.Render(fmt.Sprintf("%d", int(m.timer.Timeout.Seconds()))),
		"",
		lipgloss.JoinVertical(lipgloss.Left, rows...),
		"",
		hintStyle.Render(hint),
	)
}

func (m model) View() tea.View {
	content := ""

	if m.width != 0 {
		view := m.typingView

		if m.finished {
			view = m.resultsView
		}

		content = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			view(),
		)
	}

	v := tea.NewView(content)
	v.AltScreen = true

	return v
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
