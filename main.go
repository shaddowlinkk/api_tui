package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	config := read_config(os.Args[1])
	items := make([]list.Item, 0)
	for i, epoint := range config.Endpoints {
		items = append(items, item{title: epoint.Name, desc: epoint.Description, idx: i})
	}
	m := listModel{list: list.New(items, list.NewDefaultDelegate(), 0, 0), config: config}
	m.list.Title = "API ENDPOINTS"
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
