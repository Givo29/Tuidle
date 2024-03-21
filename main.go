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
	Correct               bool   `json:"correct"`
	Guess                 string `json:"guess"`
	CorrectLettersIndex   []int  `json:"correctLettersIndex"`
	IncorrectLettersIndex []int  `json:"incorrectLettersIndex"`
}

type PlayerState string

const (
	Playing PlayerState = "playing"
	Win     PlayerState = "win"
	Lose    PlayerState = "lose"
)

type game struct {
	Date    string `json:"date"`
	Word    string
	State   PlayerState `json:"state"`
	Guesses []Guess     `json:"guesses"`
	Streak  int         `json:"streak"`
}

type Model struct {
	width    int
	height   int
	game     game
	lastGame game
	words    []string
	maxTries int

	guessInput textinput.Model
	inputStyle lipgloss.Style
}

func checkGuess(word, guess string) Guess {
	var g Guess
	g.Guess = guess

	// Handle correct guess first
	if word == guess {
		g.Correct = true
		g.CorrectLettersIndex = []int{0, 1, 2, 3, 4}
		return g
	}

	// Handle incorrect guess
	for i, l := range guess {
		if l == rune(word[i]) {
			g.CorrectLettersIndex = append(g.CorrectLettersIndex, i)
		} else if slices.Contains(strings.Split(word, ""), string(l)) {
			g.IncorrectLettersIndex = append(g.IncorrectLettersIndex, i)
		}
	}

	return g
}

func checkPlayerState(guesses []Guess, maxTries int) PlayerState {
	// Check if the last guess was correct
	if len(guesses) > 0 && guesses[len(guesses)-1].Correct {
		return Win
	}

	// Check if the player has run out of tries
	if len(guesses) == maxTries {
		return Lose
	}

	return Playing
}

func saveGameToFile(m Model) {
	var games []game
	var currentGame game

	file, err := os.ReadFile(fmt.Sprintf("%s/.tuidle.json", os.Getenv("HOME")))
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal([]byte(file), &games)

	// Either use the last result or make new one
	currentDate := time.Now().Format("2006-01-02")
	if len(games) > 0 && games[len(games)-1].Date == currentDate {
		currentGame = games[len(games)-1]
		games = games[:len(games)-1]
	} else {
		currentGame = game{
			Date: m.game.Date,
		}
	}

	currentGame.State = m.game.State
	currentGame.Guesses = m.game.Guesses

	streak, err := checkStreak(currentGame.State, m.lastGame.Streak, m.lastGame.Date, m.game.Date)
	if err != nil {
		fmt.Println("Error checking streak")
		return
	}
	currentGame.Streak = streak

	games = append(games, currentGame)
	marshalledGames, err := json.Marshal(games)
	if err != nil {
		return
	}
	err = os.WriteFile(fmt.Sprintf("%s/.tuidle.json", os.Getenv("HOME")), marshalledGames, 0666)
	if err != nil {
		fmt.Println(err)
	}
}

func getGameByIndex(index int) (game, error) {
	file, err := os.ReadFile(fmt.Sprintf("%s/.tuidle.json", os.Getenv("HOME")))
	if err != nil {
		return game{}, err
	}
	var games []game
	json.Unmarshal([]byte(file), &games)

	if len(games) == 0 {
		return game{}, nil
	}

	if len(games)+index < 0 {
		return game{}, nil
	}
	return games[len(games)+index], nil
}

func checkStreak(state PlayerState, lastStreak int, lastDateString, todayDateString string) (int, error) {
	if state == Lose {
		return 0, nil
	}

	if state == Playing {
		return lastStreak, nil
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
			if m.game.State != Playing || len(m.guessInput.Value()) != 5 {
				return m, nil
			}

			if !slices.Contains(m.words, m.guessInput.Value()) {
				m.guessInput.SetValue("")
				m.inputStyle = invalidInputStyle
				return m, nil
			}

			m.inputStyle = validInputStyle

			if len(m.game.Guesses) < m.maxTries {
				m.game.Guesses = append(m.game.Guesses, checkGuess(m.game.Word, m.guessInput.Value()))
				m.guessInput.SetValue("")
			}
			m.game.State = checkPlayerState(m.game.Guesses, m.maxTries)
			saveGameToFile(m)
			return m, nil
		}
	}

	m.guessInput, cmd = m.guessInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var s string

	switch m.game.State {
	case Playing:
		s += "Make a guess!\n\n"

	case Win:
		s += "You win!\n"
		s += fmt.Sprintf("The word was: %s\n", m.game.Word)
		s += fmt.Sprintf("You made %d guesses\n\n", len(m.game.Guesses))

	case Lose:
		s += "You lose!\n"
		s += fmt.Sprintf("The word was: %s\n\n", m.game.Word)
	}

	for _, g := range m.game.Guesses {
		var letters []string
		for i, l := range g.Guess {
			if slices.Contains(g.CorrectLettersIndex, i) {
				letters = append(letters, correctStyle.Render(fmt.Sprintf("%s", string(l))))
			} else if slices.Contains(g.IncorrectLettersIndex, i) {
				letters = append(letters, incorrectPositionStyle.Render(fmt.Sprintf("%s", string(l))))
			} else {
				letters = append(letters, incorrectLetterStyle.Render(fmt.Sprintf("%s", string(l))))
			}
		}

		s += lipgloss.JoinHorizontal(lipgloss.Left, letters...)
		s += "\n"
	}

	if m.game.State == Playing {
		s += m.inputStyle.Render(m.guessInput.View())
	} else {
		streak, _ := checkStreak(m.game.State, m.lastGame.Streak, m.lastGame.Date, m.game.Date)
		s += fmt.Sprintf("Your current streak is %d\n", streak)
	}
	s += "\nPress Ctrl+C to quit\n"

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, s))

}

func main() {
	currentTime := time.Now()
	currentGame, _ := getGameByIndex(-1)
	lastGameIdx := -1

	if currentGame.Date == time.Now().Format("2006-01-02") {
		// If game was already finished today
		if currentGame.State != Playing {
			fmt.Println("You've already played today, come back tomorrow for the next word!")
			fmt.Println("Your current streak is", currentGame.Streak)
			return
		}

		lastGameIdx = -2
	}
	lastGame, _ := getGameByIndex(lastGameIdx)

	words := strings.Split(wordsString, "\n")

	// Seed random generator with date and generate random word index
	year, month, day := currentTime.Date()
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
		words:      words,
		maxTries:   6,
		guessInput: textInput,
		inputStyle: validInputStyle,
	}

	if currentGame.Date == time.Now().Format("2006-01-02") {
		model.game = currentGame
	} else {
		model.game = game{
			Date:  date.Format("2006-01-02"),
			State: Playing,
		}
	}
	model.lastGame = lastGame
	model.game.Word = words[num]

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
