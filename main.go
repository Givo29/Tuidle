package main

// Wordle game written in Golang with Bubbletea

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Guess struct {
	correct               bool
	guess                 string
	correctLettersIndex   []int
	incorrectLettersIndex []int
}

type Model struct {
	word     string
	msg      string
	maxTries int
	guesses  []Guess

	guessInput textinput.Model
}

func checkGuess(word, guess string) Guess {
	var g Guess
	g.guess = guess

	// Handle correct guess first
	if word == guess {
		g.correct = true
		g.correctLettersIndex = []int{0, 1, 2, 3, 4}
		return g
	}

	// Handle incorrect guess
	for i, l := range guess {
		if l == rune(word[i]) {
			g.correctLettersIndex = append(g.correctLettersIndex, i)
		} else if slices.Contains(strings.Split(word, ""), string(l)) {
			g.incorrectLettersIndex = append(g.incorrectLettersIndex, i)
		}
	}

	return g
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if len(m.guesses) < m.maxTries && len(m.guessInput.Value()) == 5 {
				m.guesses = append(m.guesses, checkGuess(m.word, m.guessInput.Value()))
				m.guessInput.SetValue("")
			}
			return m, nil
		}
	}

	m.guessInput, cmd = m.guessInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	s := "Make a guess!\n"
	s += fmt.Sprintf("Legend: (correct) [wrong position]\n\n")

	for _, g := range m.guesses {
		for i, l := range g.guess {
			if slices.Contains(g.correctLettersIndex, i) {
				s += fmt.Sprintf("(%s)", string(l))
			} else if slices.Contains(g.incorrectLettersIndex, i) {
				s += fmt.Sprintf("[%s]", string(l))
			} else {
				s += fmt.Sprintf("%s", string(l))
			}
		}
		s += "\n"
	}
	s += fmt.Sprintf("\n%s", m.guessInput.View())
	s += "\n\nPress Ctrl+C to quit\n"

	return s
}

func main() {
	words := []string{"about", "penne"}

	textInput := textinput.New()
	textInput.Focus()
	textInput.Placeholder = "Guess a word"
	textInput.CharLimit = 5
	textInput.Width = 20

	model := Model{
		// Choose first word for now
		word:       words[0],
		maxTries:   6,
		guessInput: textInput,
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
