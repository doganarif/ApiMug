package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// InputField represents a text input field
type InputField struct {
	Label    string
	Input    textinput.Model
	Required bool
}

// NewInputField creates a new input field
func NewInputField(label, placeholder string, required bool) InputField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 500
	ti.Width = 50

	return InputField{
		Label:    label,
		Input:    ti,
		Required: required,
	}
}

// Update updates the input field
func (f *InputField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.Input, cmd = f.Input.Update(msg)
	return cmd
}

// View renders the input field
func (f *InputField) View() string {
	label := f.Label
	if f.Required {
		label += " *"
	}
	return inputLabelStyle.Render(label) + "\n" + f.Input.View()
}

// Value returns the input value
func (f *InputField) Value() string {
	return f.Input.Value()
}

// SetValue sets the input value
func (f *InputField) SetValue(v string) {
	f.Input.SetValue(v)
}

// Focus focuses the input
func (f *InputField) Focus() tea.Cmd {
	return f.Input.Focus()
}

// Blur removes focus
func (f *InputField) Blur() {
	f.Input.Blur()
}
