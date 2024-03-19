package main

// Wordle game written in Golang with Bubbletea

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	word    string
	guesses []string
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string { return "Hello, World" }

func main() {
	words := []string{"about", "penne"}
	model := Model{
		word: words[0],
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
