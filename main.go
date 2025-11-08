// Package main is the entry point for the Chronos calendar application.
package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samuelstranges/chronos/app"
	"github.com/samuelstranges/chronos/util"
)

func main() {
	bgHex := util.GetBackgroundColorHex()
	model := app.NewModel(bgHex)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
