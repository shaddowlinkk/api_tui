package main

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	titleStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#6163f2"))
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	nonval              = lipgloss.NewStyle().Foreground(lipgloss.Color("#f20000"))
	focusedButton       = focusedStyle.Render("[ Submit ]")
	blurredButton       = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type inputsModle struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	config     apiConfig
	isAuth     bool
	idx        int
	lastModle  tea.Model
	values     []string
}

func buildInputs(m inputsModle, keys []string) inputsModle {
	m.inputs = make([]textinput.Model, 0)
	var t textinput.Model
	for i, key := range keys {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.Placeholder = key
		t.CharLimit = 64
		if i == 0 {
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		}
		m.inputs = append(m.inputs, t)
	}
	m.values = make([]string, len(m.inputs))
	return m
}
func initInputModel(config apiConfig, idx int, lastModel tea.Model) inputsModle {
	modle := inputsModle{config: config, isAuth: false, lastModle: lastModel, idx: idx}
	if config.Endpoints[idx].Auth {
		modle.isAuth = true
		return buildInputs(modle, config.AuthKeys)
	}
	return buildInputs(modle, config.Endpoints[idx].Keys)
}

func (m inputsModle) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputsModle) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		// Change cursor mode
		case "esc":
			return m.lastModle, nil

		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				none := true
				for i := range m.inputs {
					if len(m.inputs[i].Value()) == 0 {
						m.inputs[i].PlaceholderStyle = nonval
						none = false
					} else {
						m.values[i] = m.inputs[i].Value()
					}

				}
				if none {
					if m.isAuth {
						m.config.AuthValues = m.values
						modle := inputsModle{config: m.config, isAuth: false, lastModle: m.lastModle, idx: m.idx}
						if len(m.config.Endpoints[m.idx].Keys) > 0 {
							return buildInputs(modle, m.config.Endpoints[m.idx].Keys), nil
						}
					}
					httpm := httpModel{lastModle: m.lastModle, config: m.config, values: m.values, idx: m.idx}
					return httpm, httpm.Init()
				}
				return m, nil
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *inputsModle) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m inputsModle) View() string {
	var b strings.Builder
	if m.isAuth {
		fmt.Fprintf(&b, " ")
		b.WriteString(titleStyle.Render(" ENTERING AUTH VALUES"))
	} else {
		fmt.Fprintf(&b, " ")
		b.WriteString(titleStyle.Render("ENTERING API VALUES"))
	}
	fmt.Fprintf(&b, "\n")
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)
	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}
