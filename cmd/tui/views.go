package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderWithBorderTitleAndHeight renders content with a title embedded in the top border and minimum height
func renderWithBorderTitleAndHeight(content, title string, width int, height int) string {
	// Trim any leading/trailing whitespace from content
	content = strings.TrimSpace(content)

	// Calculate content width and height (minus borders)
	contentWidth := width - 2 // subtract left and right borders
	if contentWidth < 1 {
		contentWidth = 1
	}
	contentHeight := height - 2 // subtract top and bottom borders
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Apply border to content (without top border) with specified width and height
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, true, true, true). // no top border
		BorderForeground(borderColor).
		Width(contentWidth).
		Height(contentHeight)

	borderedContent := contentStyle.Render(content)

	// Use the specified width for the border
	actualWidth := width

	// Border characters
	topLeft := "╭"
	topRight := "╮"
	horizontal := "─"

	// Render styled title without margins for border use
	borderTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor)
	styledTitle := borderTitleStyle.Render(title)
	titleWidth := lipgloss.Width(styledTitle)

	// Calculate horizontal bars needed
	// Total: actualWidth = corner(1) + leftBars + space(1) + title + space(1) + rightBars + corner(1)
	innerWidth := actualWidth - 2 // subtract both corners
	leftBars := 1
	titleWithSpaces := 1 + titleWidth + 1 // space + title + space
	rightBars := innerWidth - leftBars - titleWithSpaces

	if rightBars < 0 {
		rightBars = 0
	}

	// Build top border: ╭─ Title ─────────╮
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	topBorder := borderStyle.Render(topLeft+strings.Repeat(horizontal, leftBars)) +
		borderStyle.Render(" ") +
		styledTitle +
		borderStyle.Render(" "+strings.Repeat(horizontal, rightBars)+topRight)

	// Combine top border with content
	return topBorder + "\n" + borderedContent
}

// renderWithBorderTitle renders content with a title embedded in the top border
func renderWithBorderTitle(content, title string, width int) string {
	// Trim any leading/trailing whitespace from content
	content = strings.TrimSpace(content)

	// Calculate content width (width minus borders)
	contentWidth := width - 2 // subtract left and right borders
	if contentWidth < 1 {
		contentWidth = 1
	}

	// Apply border to content (without top border) with specified width
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, true, true, true). // no top border
		BorderForeground(borderColor).
		Width(contentWidth)

	borderedContent := contentStyle.Render(content)

	// Use the specified width for the border
	actualWidth := width

	// Border characters
	topLeft := "╭"
	topRight := "╮"
	horizontal := "─"

	// Render styled title without margins for border use
	borderTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor)
	styledTitle := borderTitleStyle.Render(title)
	titleWidth := lipgloss.Width(styledTitle)

	// Calculate horizontal bars needed
	// Total: actualWidth = corner(1) + leftBars + space(1) + title + space(1) + rightBars + corner(1)
	innerWidth := actualWidth - 2 // subtract both corners
	leftBars := 1
	titleWithSpaces := 1 + titleWidth + 1 // space + title + space
	rightBars := innerWidth - leftBars - titleWithSpaces

	if rightBars < 0 {
		rightBars = 0
	}

	// Build top border: ╭─ Title ─────────╮
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	topBorder := borderStyle.Render(topLeft+strings.Repeat(horizontal, leftBars)) +
		borderStyle.Render(" ") +
		styledTitle +
		borderStyle.Render(" "+strings.Repeat(horizontal, rightBars)+topRight)

	// Combine top border with content
	return topBorder + "\n" + borderedContent
}

// renderSplitView renders the endpoint list and request builder side by side
func renderSplitView(m Model) string {
	// Fixed width for left panel (endpoint list)
	leftWidth := 60
	rightWidth := m.width - leftWidth

	// Ensure minimum widths
	if leftWidth > m.width-20 {
		leftWidth = m.width / 2
		rightWidth = m.width - leftWidth
	}

	// Left panel: Endpoint list (fixed width)
	leftPanel := renderEndpointListPanel(m, leftWidth)

	// Right panel: Request builder (fills remaining width)
	rightPanel := renderRequestBuilderPanel(m, rightWidth)

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

// renderEndpointListPanel renders just the endpoint list panel
func renderEndpointListPanel(m Model, width int) string {
	// Endpoint list
	listContent := m.endpointList.View()

	// Footer - show different help based on search state
	var help string
	if m.currentScreen == ScreenEndpointList {
		if m.searchMode {
			searchText := m.searchBuffer
			if searchText == "" {
				searchText = "_"
			}
			help = footerStyle.Render(fmt.Sprintf("search: %s • enter/esc: exit search", searchText))
		} else {
			help = footerStyle.Render("j/k: navigate • /: search • enter: send • tab: configure →")
		}
	} else {
		help = footerStyle.Render("j/k: navigate • enter: send • esc: back to list")
	}

	// Calculate how many lines we need to fill to push help to bottom
	listHeight := lipgloss.Height(listContent)
	helpHeight := lipgloss.Height(help) + 2                  // +2 for spacing before help
	contentHeight := m.height - 3                            // minus top and bottom borders, and global status bar
	fillLines := contentHeight - listHeight - helpHeight + 1 // +1 to move help one row lower
	if fillLines < 0 {
		fillLines = 0
	}

	var b strings.Builder
	b.WriteString(listContent)
	b.WriteString(strings.Repeat("\n", fillLines))
	b.WriteString("\n\n" + help)

	return renderWithBorderTitleAndHeight(b.String(), "Endpoints", width, m.height-1)
}

// renderRequestBuilderPanel renders just the request builder panel with response below
func renderRequestBuilderPanel(m Model, width int) string {
	if m.selectedEndpoint.Name == "" {
		// No endpoint selected
		emptyMsg := helpStyle.Render("← Select an endpoint to configure request")
		return borderStyle.Render(emptyMsg)
	}

	// Top: Request configuration (render first to get actual height)
	configPanel := renderRequestBuilderContent(m, width)

	// Get actual height of config panel
	configPanelHeight := lipgloss.Height(configPanel)

	// Bottom: Response panel fills remaining space (accounting for global status bar)
	responseHeight := m.height - configPanelHeight - 1
	if responseHeight < 10 {
		responseHeight = 10 // Minimum height for response panel
	}
	responsePanel := renderResponsePanel(m, width, responseHeight)

	// Join vertically
	return lipgloss.JoinVertical(lipgloss.Left, configPanel, responsePanel)
}

// renderRequestBuilderContent renders the full request configuration content
func renderRequestBuilderContent(m Model, width int) string {
	var b strings.Builder

	// Method, name and description
	methodName := titleStyle.Render(fmt.Sprintf("%s %s", m.selectedEndpoint.Method, m.selectedEndpoint.Name))
	description := subtitleStyle.Render(" - " + m.selectedEndpoint.Description)
	titleLine := lipgloss.JoinHorizontal(lipgloss.Left, methodName, description)
	b.WriteString(titleLine + "\n")

	// Build URL with current path parameter values
	url := m.baseURL + m.selectedEndpoint.Path
	for paramName, input := range m.pathParamInputs {
		placeholder := fmt.Sprintf("{%s}", paramName)
		value := input.Value()
		if value != "" {
			url = strings.ReplaceAll(url, placeholder, value)
		}
	}
	urlText := responseKeyStyle.Render("URL: ") + responseValueStyle.Render(url)
	b.WriteString(urlText + "\n\n")

	// Path parameters
	if len(m.selectedEndpoint.PathParams) > 0 {
		b.WriteString(responseHeaderStyle.Render("Path Parameters") + "\n")

		// Find max label width for alignment
		maxLabelWidth := 0
		for _, param := range m.selectedEndpoint.PathParams {
			if len(param.Name) > maxLabelWidth {
				maxLabelWidth = len(param.Name)
			}
		}

		for i, param := range m.selectedEndpoint.PathParams {
			input := m.allInputs[i]
			// Pad label to max width for alignment
			labelText := fmt.Sprintf("%-*s:", maxLabelWidth, param.Name)
			label := inputLabelStyle.Render(labelText)
			inputView := input.View()
			b.WriteString(fmt.Sprintf("  %s %s\n", label, inputView))
		}
		b.WriteString("\n")
	}

	// Query parameters
	if len(m.selectedEndpoint.QueryParams) > 0 {
		b.WriteString(responseHeaderStyle.Render("Query Parameters (optional)") + "\n")
		offset := len(m.selectedEndpoint.PathParams)

		// Find max label width for alignment
		maxLabelWidth := 0
		for _, param := range m.selectedEndpoint.QueryParams {
			if len(param.Name) > maxLabelWidth {
				maxLabelWidth = len(param.Name)
			}
		}

		for i, param := range m.selectedEndpoint.QueryParams {
			input := m.allInputs[offset+i]
			// Pad label to max width for alignment
			labelText := fmt.Sprintf("%-*s:", maxLabelWidth, param.Name)
			label := inputLabelStyle.Render(labelText)
			hint := helpStyle.Render(" (" + param.Description + ")")
			inputView := input.View()
			b.WriteString(fmt.Sprintf("  %s %s%s\n", label, inputView, hint))
		}
		b.WriteString("\n")
	}

	// Request body fields
	if len(m.bodyFieldInputs) > 0 {
		b.WriteString(responseHeaderStyle.Render("Request Body Fields") + "\n")
		offset := len(m.selectedEndpoint.PathParams) + len(m.selectedEndpoint.QueryParams)

		// Find max label width for alignment
		maxLabelWidth := 0
		for key := range m.bodyFieldInputs {
			if len(key) > maxLabelWidth {
				maxLabelWidth = len(key)
			}
		}

		// Iterate through allInputs to maintain order
		inputIndex := 0
		for i := offset; i < len(m.allInputs); i++ {
			// Find the corresponding key for this input
			for key, input := range m.bodyFieldInputs {
				if input.Value() == m.allInputs[i].Value() || input.Placeholder == m.allInputs[i].Placeholder {
					labelText := fmt.Sprintf("%-*s:", maxLabelWidth, key)
					label := inputLabelStyle.Render(labelText)
					inputView := m.allInputs[i].View()
					b.WriteString(fmt.Sprintf("  %s %s\n", label, inputView))
					inputIndex++
					break
				}
			}
		}
		b.WriteString("\n")
	}

	// Error message if any
	if m.errorMsg != "" {
		b.WriteString("\n" + errorMessageStyle.Render("Error: "+m.errorMsg) + "\n")
	}

	// Loading indicator
	if m.isLoading {
		b.WriteString("\n" + infoStyle.Render("⏳ Sending request...") + "\n")
	}

	// Footer
	help := footerStyle.Render("tab: next field • enter: send request • esc/q: back")
	b.WriteString("\n" + help)

	content := b.String()

	// Use custom border with title
	return renderWithBorderTitle(content, "Request", width)
}

// renderResponseViewer renders the response display screen
func renderResponseViewer(m Model) string {
	if m.currentResult == nil {
		return "No response data"
	}

	var b strings.Builder

	b.WriteString("TOP OF VIEW\n\n")

	result := m.currentResult

	// Title
	title := titleStyle.Render(fmt.Sprintf("📊 Response: %s", result.EndpointName))
	b.WriteString(title + "\n\n")

	// Status and timing
	statusBadge := renderStatusBadge(result.StatusCode, result.Status)
	timing := infoStyle.Render(fmt.Sprintf("Duration: %v", result.Duration))

	statusLine := lipgloss.JoinHorizontal(lipgloss.Left, statusBadge, "  ", timing)
	b.WriteString(statusLine + "\n\n")

	// Request Details Section
	requestSection := responseHeaderStyle.Render("📤 Request") + "\n"
	requestMethod := responseKeyStyle.Render(string(result.Method))
	requestURL := responseValueStyle.Render(result.RequestURL)
	requestSection += fmt.Sprintf("  %s %s\n", requestMethod, requestURL)

	// Request headers
	requestSection += fmt.Sprintf("  %s %s\n",
		responseKeyStyle.Render("X-User-ID:"),
		responseValueStyle.Render(result.RequestUserID),
	)

	// Request body (if any)
	if result.RequestBody != "" {
		requestSection += fmt.Sprintf("  %s\n", responseKeyStyle.Render("Body:"))
		// Format and indent the JSON - show first 3 lines
		lines := strings.Split(result.RequestBody, "\n")
		displayLines := 0
		for _, line := range lines {
			if line != "" && displayLines < 3 {
				requestSection += "    " + responseValueStyle.Render(line) + "\n"
				displayLines++
			}
		}
		if len(lines) > 3 {
			requestSection += "    " + helpStyle.Render("...") + "\n"
		}
	}
	requestSection += "\n"

	// Response headers (selected important ones)
	responseHeaderSection := responseHeaderStyle.Render("📥 Response Headers") + "\n"
	responseHeaderSection += formatHeaders(result.Headers) + "\n"

	// Combine fixed sections
	b.WriteString("RRRS\n")
	b.WriteString(requestSection)
	b.WriteString("RS\n")
	b.WriteString(responseHeaderSection)

	// Response body
	b.WriteString(responseHeaderStyle.Render("Response Body") + "\n")
	if result.Error != nil {
		errorBox := errorBadgeStyle.Render(fmt.Sprintf("Error: %v", result.Error))
		b.WriteString(errorBox + "\n")
	} else {
		b.WriteString(m.responseViewport.View() + "\n")
	}

	// Footer
	help := footerStyle.Render("↑/↓: scroll • esc/q: back • enter: send again")
	b.WriteString("\n" + help)

	return b.String()
}

// renderHistory renders the request history screen
func renderHistory(m Model) string {
	var b strings.Builder

	title := titleStyle.Render("📜 Request History")
	b.WriteString(title + "\n\n")

	if len(m.history) == 0 {
		b.WriteString(helpStyle.Render("No requests in history yet.") + "\n")
	} else {
		for i := len(m.history) - 1; i >= 0 && i > len(m.history)-10; i-- {
			result := m.history[i]
			statusBadge := renderStatusBadge(result.StatusCode, result.Status)
			timestamp := result.RequestTime.Format("15:04:05")
			line := fmt.Sprintf("%s  %s  %s %s",
				helpStyle.Render(timestamp),
				statusBadge,
				responseKeyStyle.Render(string(result.Method)),
				result.EndpointName,
			)
			b.WriteString(line + "\n")
		}
	}

	help := footerStyle.Render("esc/q: back")
	b.WriteString("\n" + help)

	return borderStyle.Render(b.String())
}

// renderSettings renders the settings screen
func renderSettings(m Model) string {
	var b strings.Builder

	title := titleStyle.Render("⚙️  Settings")
	b.WriteString(title + "\n\n")

	b.WriteString(inputLabelStyle.Render("Base URL:") + "\n")
	b.WriteString(m.baseURL + "\n\n")

	b.WriteString(inputLabelStyle.Render("User ID (X-User-ID header):") + "\n")
	b.WriteString(m.userID + "\n\n")

	help := footerStyle.Render("esc/q: back")
	b.WriteString("\n" + help)

	return borderStyle.Render(b.String())
}

// renderMakeScreen renders the make commands screen with split panes
func renderMakeScreen(m Model) string {
	// Fixed width for left panel (command list)
	leftWidth := 60
	rightWidth := m.width - leftWidth

	// Ensure minimum widths
	if leftWidth > m.width-20 {
		leftWidth = m.width / 2
		rightWidth = m.width - leftWidth
	}

	// Left panel: Make command list
	leftPanel := renderMakeCommandList(m, leftWidth)

	// Right panel: Command output
	rightPanel := renderMakeOutput(m, rightWidth)

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

// renderMakeCommandList renders the make command list panel
func renderMakeCommandList(m Model, width int) string {
	// Command list
	listContent := m.makeCommandList.View()

	// Footer
	help := footerStyle.Render("j/k: navigate • enter: run • esc/q: back")

	// Calculate how many lines we need to fill to push help to bottom
	listHeight := lipgloss.Height(listContent)
	helpHeight := lipgloss.Height(help) + 2                  // +2 for spacing before help
	contentHeight := m.height - 3                            // minus top and bottom borders, and global status bar
	fillLines := contentHeight - listHeight - helpHeight + 1 // +1 to move help one row lower
	if fillLines < 0 {
		fillLines = 0
	}

	var b strings.Builder
	b.WriteString(listContent)
	b.WriteString(strings.Repeat("\n", fillLines))
	b.WriteString("\n\n" + help)

	return renderWithBorderTitleAndHeight(b.String(), "Make Commands", width, m.height-1)
}

// renderMakeOutput renders the command output panel
func renderMakeOutput(m Model, width int) string {
	var b strings.Builder

	if m.commandRunning {
		b.WriteString(infoStyle.Render("⏳ Running command...") + "\n\n")
	}

	if m.makeCommandOutput != "" {
		// Calculate viewport height
		viewportHeight := m.height - 8
		if viewportHeight < 5 {
			viewportHeight = 5
		}
		m.makeOutputViewport.Height = viewportHeight
		m.makeOutputViewport.SetContent(m.makeCommandOutput)
		b.WriteString(m.makeOutputViewport.View())
	} else if m.selectedMakeCommand.Name != "" {
		b.WriteString(helpStyle.Render("Press enter to run: make " + m.selectedMakeCommand.Name))
	} else {
		b.WriteString(helpStyle.Render("← Select a make command"))
	}

	return renderWithBorderTitleAndHeight(b.String(), "Output", width, m.height-1)
}

// renderResponsePanel renders a compact response panel with fixed height
func renderResponsePanel(m Model, width int, height int) string {
	var b strings.Builder

	if m.currentResult == nil {
		// No response yet
		emptyMsg := helpStyle.Render("No response yet - press enter to send request")
		b.WriteString(emptyMsg)
	} else {
		result := m.currentResult

		// Status, timing, and content type on one line
		statusBadge := renderStatusBadge(result.StatusCode, result.Status)
		timing := infoStyle.Render(fmt.Sprintf("Duration: %v", result.Duration))

		// Build status line with optional content type
		statusLineParts := []string{statusBadge, "  ", timing}
		if contentType, ok := result.Headers["Content-Type"]; ok && len(contentType) > 0 {
			contentTypeText := "  " + responseKeyStyle.Render("Content-Type: ") +
				responseValueStyle.Render(contentType[0])
			statusLineParts = append(statusLineParts, contentTypeText)
		}
		statusLine := lipgloss.JoinHorizontal(lipgloss.Left, statusLineParts...)
		b.WriteString(statusLine + "\n\n")

		// Response body
		if result.Error != nil {
			errorBox := errorBadgeStyle.Render(fmt.Sprintf("Error: %v", result.Error))
			b.WriteString(errorBox + "\n")
		} else {
			// Calculate viewport height based on available space
			// Account for: borders (2), status line (3)
			viewportHeight := height - 2 - 3
			if viewportHeight < 5 {
				viewportHeight = 5
			}
			// Update viewport height to fill available space
			m.responseViewport.Height = viewportHeight
			b.WriteString(m.responseViewport.View() + "\n")
		}
	}

	content := b.String()

	// Use custom border with title and set minimum height to fill panel
	return renderWithBorderTitleAndHeight(content, "Response", width, height)
}

// renderStatsModal renders a modal overlay showing stats
func renderStatsModal(m Model, bgContent string) string {
	// Modal dimensions
	modalWidth := 50

	// Modal content
	var content strings.Builder

	title := titleStyle.Render("📊 API Stats")
	content.WriteString(title + "\n\n")

	if m.statsLoading {
		content.WriteString(infoStyle.Render("⏳ Loading stats...") + "\n")
	} else {
		// Total counts from API
		templatesLine := lipgloss.JoinHorizontal(lipgloss.Left,
			responseKeyStyle.Render("Templates: "),
			responseValueStyle.Render(fmt.Sprintf("%d", m.templateCount)),
		)
		chainsLine := lipgloss.JoinHorizontal(lipgloss.Left,
			responseKeyStyle.Render("Chains: "),
			responseValueStyle.Render(fmt.Sprintf("%d", m.chainCount)),
		)
		content.WriteString(templatesLine + "\n")
		content.WriteString(chainsLine + "\n")

		// Cached data from background fetch
		content.WriteString("\n" + responseHeaderStyle.Render("Cached Data") + "\n")

		cachedChainsLine := lipgloss.JoinHorizontal(lipgloss.Left,
			responseKeyStyle.Render("Cached Chains: "),
			responseValueStyle.Render(fmt.Sprintf("%d", len(m.cachedChains))),
		)
		cachedTemplatesLine := lipgloss.JoinHorizontal(lipgloss.Left,
			responseKeyStyle.Render("Cached Templates: "),
			responseValueStyle.Render(fmt.Sprintf("%d", len(m.cachedTemplates))),
		)
		content.WriteString(cachedChainsLine + "\n")
		content.WriteString(cachedTemplatesLine + "\n")

		// Show first chain if available
		if len(m.cachedChains) > 0 {
			firstChainLine := lipgloss.JoinHorizontal(lipgloss.Left,
				responseKeyStyle.Render("First Chain: "),
				responseValueStyle.Render(truncateString(m.cachedChains[0].Name, 20)),
			)
			firstChainIDLine := lipgloss.JoinHorizontal(lipgloss.Left,
				responseKeyStyle.Render("  ID: "),
				helpStyle.Render(truncateString(m.cachedChains[0].ID, 30)),
			)
			content.WriteString(firstChainLine + "\n")
			content.WriteString(firstChainIDLine + "\n")
		}
	}

	content.WriteString("\n")
	help := footerStyle.Render("press any key to close")
	content.WriteString(help)

	// Create modal box
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(modalWidth).
		Render(content.String())

	// Overlay modal on background content
	return lipgloss.Place(
		m.width,
		m.height-1, // -1 for global help bar
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
	)
}

// renderGlobalHelp renders the global keyboard shortcuts help line at bottom of display
func renderGlobalHelp(width int) string {
	// Create styled shortcuts
	f1Key := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render("(F1)")
	f1Label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render(" Endpoints")

	f2Key := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render("(F2)")
	f2Label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render(" Make")

	dKey := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render("(d)")
	dLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render(" Stats")

	qKey := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Render("(q)")
	qLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render(" Quit")

	// Join left shortcuts with spacing
	shortcut1 := f1Key + f1Label
	shortcut2 := f2Key + f2Label
	leftShortcuts := lipgloss.JoinHorizontal(lipgloss.Left, shortcut1, "  ", shortcut2)

	// Right shortcuts (d and q)
	shortcut3 := dKey + dLabel
	quitShortcut := qKey + qLabel
	rightShortcuts := lipgloss.JoinHorizontal(lipgloss.Left, shortcut3, "  ", quitShortcut)

	// Calculate spacing to push right shortcuts to far right
	leftWidth := lipgloss.Width(leftShortcuts)
	rightWidth := lipgloss.Width(rightShortcuts)
	spacingWidth := width - leftWidth - rightWidth
	if spacingWidth < 2 {
		spacingWidth = 2
	}

	// Join with right shortcuts on the far right
	helpLine := lipgloss.JoinHorizontal(lipgloss.Left,
		leftShortcuts,
		strings.Repeat(" ", spacingWidth),
		rightShortcuts,
	)

	// Style the entire help line with a background and full width
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a1a")).
		Foreground(lipgloss.Color("#FAFAFA")).
		Width(width).
		Render(helpLine)
}

// Helper functions

// formatHeaders formats HTTP headers for display
func formatHeaders(headers map[string][]string) string {
	var b strings.Builder
	importantHeaders := []string{"Content-Type", "Content-Length", "Date"}

	for _, key := range importantHeaders {
		if values, ok := headers[key]; ok {
			b.WriteString(fmt.Sprintf("  %s: %s\n",
				responseKeyStyle.Render(key),
				responseValueStyle.Render(strings.Join(values, ", ")),
			))
		}
	}

	if b.Len() == 0 {
		return "  " + responseValueStyle.Render("(none)")
	}

	return b.String()
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
