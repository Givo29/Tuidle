package main

// Wordle game written in Golang with Bubbletea

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed words.txt
var wordsString string

var incorrectLetterStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
var incorrectPositionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff8100")).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#ff8100")).Padding(0, 1)
var correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#04B575")).Padding(0, 1)
var validInputStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(25)
var invalidInputStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#ff0000")).Width(25)

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

type game struct {
	words          []string
	word           string
	dateString     string
	state          PlayerState
	maxTries       int
	guesses        []Guess
	previousResult result
}

type Model struct {
	width  int
	height int
	game   game

	guessInput textinput.Model
	inputStyle lipgloss.Style
}

type result struct {
	Date    string `json:"date"`
	Word    string `json:"word"`
	Guesses int    `json:"guesses"`
	Win     bool   `json:"win"`
	Streak  int    `json:"streak"`
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

func saveGameToFile(m Model) {
	var results []result

	file, err := os.ReadFile("~/.termdle.json")
	json.Unmarshal([]byte(file), &results)

	result := result{
		Date:    m.game.dateString,
		Word:    m.game.word,
		Guesses: len(m.game.guesses),
		Win:     m.game.state == Win,
	}

	streak, err := checkStreak(result.Win, m.game.previousResult.Streak, m.game.previousResult.Date, m.game.dateString)
	if err != nil {
		fmt.Println("Error checking streak")
		return
	}
	result.Streak = streak

	results = append(results, result)
	marshalledResults, err := json.Marshal(results)
	if err != nil {
		return
	}
	err = os.WriteFile(fmt.Sprintf("%s/.termdle.json", os.Getenv("HOME")), marshalledResults, 0666)
	if err != nil {
		fmt.Println(err)
	}
}

func getPreviousResult() (result, error) {
	file, err := os.ReadFile(fmt.Sprintf("%s/.termdle.json", os.Getenv("HOME")))
	if err != nil {
		return result{}, err
	}
	var results []result
	json.Unmarshal([]byte(file), &results)
	if len(results) == 0 {
		return result{}, nil
	}

	return results[len(results)-1], nil
}

func checkStreak(win bool, lastStreak int, lastDateString, todayDateString string) (int, error) {
	if !win {
		return 0, nil
	}

	todayDate, err := time.Parse("2006-01-02", todayDateString)
	if err != nil {
		return 1, err
	}
	lastDate, _ := time.Parse("2006-01-02", lastDateString)
	if err != nil {
		return 1, err
	}

	if todayDate.Sub(lastDate).Hours() == 24 {
		return lastStreak + 1, nil
	} else {
		return 1, nil
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.game.state != Playing || len(m.guessInput.Value()) != 5 {
				return m, nil
			}

			if !slices.Contains(m.game.words, m.guessInput.Value()) {
				m.guessInput.SetValue("")
				m.inputStyle = invalidInputStyle
				return m, nil
			}

			m.inputStyle = validInputStyle

			if len(m.game.guesses) < m.game.maxTries {
				m.game.guesses = append(m.game.guesses, checkGuess(m.game.word, m.guessInput.Value()))
				m.guessInput.SetValue("")
			}
			m.game.state = checkPlayerState(m.game.guesses, m.game.maxTries)
			// Should only save once here
			if m.game.state != Playing {
				saveGameToFile(m)
			}
			return m, nil
		}
	}

	m.guessInput, cmd = m.guessInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var s string

	switch m.game.state {
	case Playing:
		s += "Make a guess!\n\n"

	case Win:
		s += "You win!\n"
		s += fmt.Sprintf("The word was: %s\n", m.game.word)
		s += fmt.Sprintf("You made %d guesses\n\n", len(m.game.guesses))

	case Lose:
		s += "You lose!\n"
		s += fmt.Sprintf("The word was: %s\n\n", m.game.word)
	}

	for _, g := range m.game.guesses {
		var letters []string
		for i, l := range g.guess {
			if slices.Contains(g.correctLettersIndex, i) {
				letters = append(letters, correctStyle.Render(fmt.Sprintf("%s", string(l))))
			} else if slices.Contains(g.incorrectLettersIndex, i) {
				letters = append(letters, incorrectPositionStyle.Render(fmt.Sprintf("%s", string(l))))
			} else {
				letters = append(letters, incorrectLetterStyle.Render(fmt.Sprintf("%s", string(l))))
			}
		}

		s += lipgloss.JoinHorizontal(lipgloss.Left, letters...)
		s += "\n"
	}

	if m.game.state == Playing {
		s += m.inputStyle.Render(m.guessInput.View())
	} else {
		streak, _ := checkStreak(m.game.state == Win, m.game.previousResult.Streak, m.game.previousResult.Date, m.game.dateString)
		s += fmt.Sprintf("Your current streak is %d\n", streak)
	}
	s += "\nPress Ctrl+C to quit\n"

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, s))

}

func main() {
	previousResult, err := getPreviousResult()
	if err != nil {
		fmt.Println("Error reading previous result")
		os.Exit(1)
	}
	if previousResult.Date == time.Now().Format("2006-01-02") {
		fmt.Println("You've already played today, come back tomorrow for the next word!")
		fmt.Println("Your current streak is", previousResult.Streak)
		return
	}
	words := strings.Split(wordsString, "\n")

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
		game: game{
			words:          words,
			word:           words[num],
			dateString:     date.Format("2006-01-02"),
			state:          Playing,
			maxTries:       6,
			previousResult: previousResult,
		},
		guessInput: textInput,
		inputStyle: validInputStyle,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
