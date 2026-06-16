// Package ui holds terminal presentation helpers built on lipgloss (styling)
// and huh (interactive prompts). It centralizes styling so command code stays
// focused on behavior.
package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle   = lipgloss.NewStyle().Faint(true)
)

// Title prints a bold heading.
func Title(s string) { fmt.Println(titleStyle.Render(s)) }

// Success prints a green confirmation line.
func Success(format string, a ...any) {
	fmt.Println(okStyle.Render("✓ ") + fmt.Sprintf(format, a...))
}

// Warn prints a yellow warning line to stderr.
func Warn(format string, a ...any) {
	fmt.Fprintln(os.Stderr, warnStyle.Render("! ")+fmt.Sprintf(format, a...))
}

// Error prints a red error line to stderr.
func Error(format string, a ...any) {
	fmt.Fprintln(os.Stderr, errStyle.Render("✗ ")+fmt.Sprintf(format, a...))
}

// Info prints a plain informational line.
func Info(format string, a ...any) { fmt.Printf(format+"\n", a...) }

// Dim prints faint, secondary text.
func Dim(format string, a ...any) { fmt.Println(dimStyle.Render(fmt.Sprintf(format, a...))) }

// Confirm asks a yes/no question, defaulting to No. Returns true only on an
// explicit yes. Used for dangerous actions that require explicit choice.
func Confirm(prompt string) bool {
	confirmed := false
	err := huh.NewConfirm().
		Title(prompt).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed).
		Run()
	if err != nil {
		return false
	}
	return confirmed
}
