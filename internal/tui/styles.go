package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("#7D56F4"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	methodGetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D700")).
			Bold(true)

	methodPostStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true)

	methodPutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	methodDeleteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4500")).
			Bold(true)

	methodPatchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8A2BE2")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(1, 0)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Bold(true).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D700")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	codeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#2E3440")).
			Padding(1, 2)

	statusCodeSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D700")).
				Bold(true)

	statusCodeErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)
)

func getMethodStyle(method string) lipgloss.Style {
	switch method {
	case "get":
		return methodGetStyle
	case "post":
		return methodPostStyle
	case "put":
		return methodPutStyle
	case "delete":
		return methodDeleteStyle
	case "patch":
		return methodPatchStyle
	default:
		return lipgloss.NewStyle()
	}
}
