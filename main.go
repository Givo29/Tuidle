package main

// Wordle game written in Golang with Bubbletea

import (
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Guess struct {
	correct               bool
	guess                 string
	correctLettersIndex   []int
	incorrectLettersIndex []int
}

type PlayerState string

const (
	Playing PlayerState = "playing"
	Win     PlayerState = "win"
	Lose    PlayerState = "lose"
)

type Model struct {
	word     string
	state    PlayerState
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

func checkPlayerState(guesses []Guess, maxTries int) PlayerState {
	// Check if the last guess was correct
	if len(guesses) > 0 && guesses[len(guesses)-1].correct {
		return Win
	}

	// Check if the player has run out of tries
	if len(guesses) == maxTries {
		return Lose
	}

	return Playing
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
			m.state = checkPlayerState(m.guesses, m.maxTries)
			return m, nil
		}
	}

	m.guessInput, cmd = m.guessInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var s string

	switch m.state {
	case Playing:
		s += "Make a guess!\n"
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

	case Win:
		s += "You win!\n"
		s += fmt.Sprintf("The word was: %s\n", m.word)
		s += fmt.Sprintf("You made %d guesses\n", len(m.guesses))
		s += fmt.Sprintf("\nPress Ctrl+C to quit\n")

	case Lose:
		s += "You lose!\n"
		s += fmt.Sprintf("The word was: %s\n", m.word)
		s += fmt.Sprintf("\nPress Ctrl+C to quit\n")
	}
	return s
}

func main() {
	words := []string{"about", "penne", "apple", "table", "hello", "world", "globe", "water", "earth", "space", "music"}

	// Seed random generator with date and generate random word index
	year, month, day := time.Now().Date()
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	r := rand.New(rand.NewSource(date.UnixNano()))
	num := r.Intn(len(words))

	textInput := textinput.New()
	textInput.Focus()
	textInput.Placeholder = "Guess a word"
	textInput.CharLimit = 5
	textInput.Width = 20

	model := Model{
		// Choose first word for now
		word:       words[num],
		state:      Playing,
		maxTries:   6,
		guessInput: textInput,
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
