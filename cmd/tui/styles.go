package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#63")
	successColor   = lipgloss.Color("#04B575")
	errorColor     = lipgloss.Color("#FF0000")
	warningColor   = lipgloss.Color("#FFA500")
	textColor      = lipgloss.Color("#FAFAFA")
	dimColor       = lipgloss.Color("#666666")
	borderColor    = lipgloss.Color("#383838")

	// Status code colors
	status2xxColor = successColor
	status3xxColor = warningColor
	status4xxColor = errorColor
	status5xxColor = lipgloss.Color("#8B0000")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(false)

	// Border styles
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	// Menu styles
	menuItemStyle = lipgloss.NewStyle().
			Foreground(textColor)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				SetString("▶ ")

	// Input styles
	inputLabelStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(textColor)

	focusedInputStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	// Status badge styles
	successBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000")).
				Background(successColor).
				Bold(true)

	errorBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(errorColor).
			Bold(true)

	warningBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000")).
				Background(warningColor).
				Bold(true)

	// Response styles
	responseHeaderStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	responseKeyStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	responseValueStyle = lipgloss.NewStyle().
				Foreground(textColor)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true)

	// Footer style
	footerStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			BorderTop(true).
			BorderForeground(borderColor).
			MarginTop(1)

	// Error message style
	errorMessageStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	// Info message style
	infoStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)
)

// getStatusStyle returns the appropriate style for an HTTP status code
func getStatusStyle(statusCode int) lipgloss.Style {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return successBadgeStyle
	case statusCode >= 300 && statusCode < 400:
		return warningBadgeStyle
	case statusCode >= 400 && statusCode < 500:
		return errorBadgeStyle
	case statusCode >= 500:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(status5xxColor).
			Bold(true)
	default:
		return baseStyle
	}
}

// renderStatusBadge renders a styled status code badge
func renderStatusBadge(statusCode int, statusText string) string {
	style := getStatusStyle(statusCode)
	// Just render the status text, which already includes the code
	return style.Render(statusText)
}
