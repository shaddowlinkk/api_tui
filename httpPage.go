package main

// A simple program that makes a GET request and prints the response status.

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type httpModel struct {
	status    string
	err       error
	lastModle tea.Model
	idx       int
	config    apiConfig
	values    []string
}

type statusMsg string

type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

func (m httpModel) Init() tea.Cmd {
	return m.checkServer
}

func (m httpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "backspace":
			return m.lastModle, nil
		default:
			return m, nil
		}
	case statusMsg:
		m.status = string(msg)
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	default:
		return m, nil
	}
}

func (m httpModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, " ")
	b.WriteString(titleStyle.Render("EXECUTING HTTP"))
	fmt.Fprintf(&b, "\n")
	s := fmt.Sprintf("executing %s...", m.config.BaseUrl+m.config.Endpoints[m.idx].Endpoint)

	if m.err != nil {
		s += fmt.Sprintf("something went wrong: %s", m.err)
	} else if len(m.status) != 0 {
		s += fmt.Sprintf("%s\n", m.status)
	}

	b.WriteString(s + "\n")
	b.WriteString(helpStyle.Render("Press Backspace to return to main menu or q, ctrl+c, or esc to exit"))
	return b.String()
}

func (m httpModel) checkServer() tea.Msg {
	data, err := exec_http(m.config, m.idx, m.values)
	if err != nil {
		return errMsg{err}
	}
	return statusMsg(string(data))
}
