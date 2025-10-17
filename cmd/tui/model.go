package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen represents the current screen
type Screen int

const (
	ScreenEndpointList Screen = iota
	ScreenRequestBuilder
	ScreenResponseViewer
	ScreenHistory
	ScreenSettings
	ScreenMakeCommands
	ScreenStatsModal
)

// EndpointState stores the request inputs and response for an endpoint
type EndpointState struct {
	pathParamInputs  map[string]textinput.Model
	queryParamInputs map[string]textinput.Model
	bodyFieldInputs  map[string]textinput.Model
	response         *RequestResult
}

// MakeCommand represents a Makefile command
type MakeCommand struct {
	Name string
	Desc string
}

// Title implements list.Item interface
func (mc MakeCommand) Title() string { return mc.Name }

// Description implements list.Item interface
func (mc MakeCommand) Description() string { return mc.Desc }

// FilterValue implements list.Item interface
func (mc MakeCommand) FilterValue() string { return mc.Name }

// Model represents the application state
type Model struct {
	// Current screen
	currentScreen Screen

	// Configuration
	baseURL string
	userID  string

	// Endpoint list (all endpoints in one menu)
	endpoints     []Endpoint
	endpointIndex int
	endpointList  list.Model

	// Search state
	searchMode   bool
	searchBuffer string

	// Request builder
	selectedEndpoint  Endpoint
	pathParamInputs   map[string]textinput.Model
	queryParamInputs  map[string]textinput.Model
	bodyFieldInputs   map[string]textinput.Model
	focusedInputIndex int
	allInputs         []textinput.Model // For tab navigation

	// Response viewer
	currentResult    *RequestResult
	responseViewport viewport.Model

	// State storage per endpoint (keyed by endpoint name)
	endpointStates map[string]*EndpointState

	// History
	history     []RequestResult
	historyList list.Model

	// Settings
	settingsInputs []textinput.Model
	settingsIndex  int

	// UI dimensions
	width  int
	height int

	// Make commands screen
	makeCommands        []MakeCommand
	makeCommandList     list.Model
	makeCommandOutput   string
	makeOutputViewport  viewport.Model
	selectedMakeCommand MakeCommand
	commandRunning      bool

	// Stats modal
	templateCount  int
	chainCount     int
	statsLoading   bool
	previousScreen Screen

	// Error message
	errorMsg string

	// Loading state
	isLoading       bool
	loadingMsg      string
	requestInFlight bool

	// Cached data from background fetches
	cachedChains    []CachedChain
	cachedTemplates []CachedTemplate
}

// CachedChain represents a chain from the API
type CachedChain struct {
	ID   string `json:"id"`
	Name string `json:"chain_name"`
}

// CachedTemplate represents a template from the API
type CachedTemplate struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// requestMsg is sent when a request completes
type requestMsg RequestResult

// errMsg is sent when an error occurs
type errMsg struct {
	err error
}

// chainsUpdatedMsg is sent when background fetch updates the chains list
type chainsUpdatedMsg struct {
	chains []CachedChain
}

// templatesUpdatedMsg is sent when background fetch updates the templates list
type templatesUpdatedMsg struct {
	templates []CachedTemplate
}

// tickMsg is sent periodically to trigger background fetches
type tickMsg time.Time

// parseMakefileCommands parses Makefile and extracts commands with descriptions
func parseMakefileCommands(makefilePath string) []MakeCommand {
	data, err := os.ReadFile(makefilePath)
	if err != nil {
		return []MakeCommand{{Name: "error", Desc: "Could not read Makefile"}}
	}

	var commands []MakeCommand
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		// Look for lines with format: target: ## description
		if strings.Contains(line, ":") && strings.Contains(line, "##") {
			parts := strings.Split(line, "##")
			if len(parts) == 2 {
				targetPart := strings.TrimSpace(strings.Split(parts[0], ":")[0])
				description := strings.TrimSpace(parts[1])
				if targetPart != "" && !strings.HasPrefix(targetPart, ".") {
					commands = append(commands, MakeCommand{
						Name: targetPart,
						Desc: description,
					})
				}
			}
		}
	}

	return commands
}

// Initialize creates the initial model
func Initialize() Model {
	// Get all endpoints
	allEndpoints := GetAllEndpoints()

	// Create endpoint list items
	endpointItems := make([]list.Item, len(allEndpoints))
	for i, ep := range allEndpoints {
		endpointItems[i] = endpointItem{
			endpoint: ep,
			index:    i,
		}
	}

	// Create custom delegate with single line items and no spacing
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	endpointList := list.New(endpointItems, delegate, 0, 0)
	endpointList.Title = "🚀 Launchpad API Tester - Select Endpoint"
	endpointList.SetShowHelp(false)
	endpointList.SetShowStatusBar(false)
	endpointList.SetFilteringEnabled(false) // Disable built-in filtering, use custom search

	// Initialize response viewport
	vp := viewport.New(80, 20)

	// Parse Makefile commands
	// Path is relative to where the binary is run from (launchpad root)
	makeCommands := parseMakefileCommands("Makefile")
	makeItems := make([]list.Item, len(makeCommands))
	for i, cmd := range makeCommands {
		makeItems[i] = cmd
	}

	// Create make command list
	makeDelegate := list.NewDefaultDelegate()
	makeDelegate.ShowDescription = true
	makeDelegate.SetSpacing(0)

	makeList := list.New(makeItems, makeDelegate, 0, 0)
	makeList.Title = "Make Commands"
	makeList.SetShowHelp(false)
	makeList.SetShowStatusBar(false)
	makeList.SetFilteringEnabled(false)

	// Initialize make output viewport
	makeVp := viewport.New(80, 20)

	m := Model{
		currentScreen:      ScreenEndpointList,
		baseURL:            "http://localhost:3001",
		userID:             "550e8400-e29b-41d4-a716-446655440000",
		endpoints:          allEndpoints,
		endpointList:       endpointList,
		pathParamInputs:    make(map[string]textinput.Model),
		queryParamInputs:   make(map[string]textinput.Model),
		bodyFieldInputs:    make(map[string]textinput.Model),
		responseViewport:   vp,
		history:            []RequestResult{},
		endpointStates:     make(map[string]*EndpointState),
		makeCommands:       makeCommands,
		makeCommandList:    makeList,
		makeOutputViewport: makeVp,
	}

	// Initialize first endpoint
	if len(allEndpoints) > 0 {
		m.selectedEndpoint = allEndpoints[0]
		// Prepare inputs for first endpoint (no async fetch yet, that happens in Init())
		m.pathParamInputs = make(map[string]textinput.Model)
		m.queryParamInputs = make(map[string]textinput.Model)
		m.bodyFieldInputs = make(map[string]textinput.Model)
		m.allInputs = []textinput.Model{}
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Start background data fetching and prepare first endpoint
	return tea.Batch(
		m.prepareRequestBuilder(),
		m.fetchBackgroundData(), // Initial fetch
		tickCmd(),               // Start periodic updates
	)
}

// tickCmd returns a command that sends a tick message every 30 seconds
func tickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle modal first - any key closes it
		if m.currentScreen == ScreenStatsModal {
			m.currentScreen = m.previousScreen
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			// ctrl+c always quits
			return m, tea.Quit

		case "q":
			// q quits from main screens, goes back from others
			if m.currentScreen == ScreenEndpointList || m.currentScreen == ScreenMakeCommands {
				return m, tea.Quit
			}
			// Go back to previous screen on other screens
			m = m.goBack()
			return m, nil

		case "esc":
			// If in search mode, exit search mode
			if m.currentScreen == ScreenEndpointList && m.searchMode {
				m.searchMode = false
				m.searchBuffer = ""
				return m, nil
			}
			m = m.goBack()
			return m, nil

		case "enter":
			// If in search mode, exit search mode
			if m.currentScreen == ScreenEndpointList && m.searchMode {
				m.searchMode = false
				m.searchBuffer = ""
				return m, nil
			}
			// Send request from either screen
			return m.handleEnter()

		case "/":
			// Enter search mode
			if m.currentScreen == ScreenEndpointList && !m.searchMode {
				m.searchMode = true
				m.searchBuffer = ""
				return m, nil
			}

		case "f1":
			// Switch to tests screen (not implemented yet, go to endpoint list)
			m.currentScreen = ScreenEndpointList
			return m, nil

		case "f2":
			// Switch to make commands screen (global shortcut)
			m.currentScreen = ScreenMakeCommands
			return m, nil

		case "d":
			// Show stats modal (global shortcut)
			m.previousScreen = m.currentScreen
			m.currentScreen = ScreenStatsModal
			m.statsLoading = true
			return m, m.fetchStats()

		case "ctrl+j", "ctrl+n":
			// Navigate down in lists
			if m.currentScreen == ScreenEndpointList || m.currentScreen == ScreenRequestBuilder {
				// Save current state before navigating
				m.saveCurrentEndpointState()
				m.endpointList.CursorDown()
				// Update selected endpoint when navigating
				if selectedItem := m.endpointList.SelectedItem(); selectedItem != nil {
					if item, ok := selectedItem.(endpointItem); ok {
						m.selectedEndpoint = item.endpoint
						cmd := m.prepareRequestBuilder()
						cmds = append(cmds, cmd)
					}
				}
			} else if m.currentScreen == ScreenMakeCommands {
				m.makeCommandList.CursorDown()
				if selectedItem := m.makeCommandList.SelectedItem(); selectedItem != nil {
					if cmd, ok := selectedItem.(MakeCommand); ok {
						m.selectedMakeCommand = cmd
					}
				}
			}

		case "ctrl+k", "ctrl+p":
			// Navigate up in lists
			if m.currentScreen == ScreenEndpointList || m.currentScreen == ScreenRequestBuilder {
				// Save current state before navigating
				m.saveCurrentEndpointState()
				m.endpointList.CursorUp()
				// Update selected endpoint when navigating
				if selectedItem := m.endpointList.SelectedItem(); selectedItem != nil {
					if item, ok := selectedItem.(endpointItem); ok {
						m.selectedEndpoint = item.endpoint
						cmd := m.prepareRequestBuilder()
						cmds = append(cmds, cmd)
					}
				}
			} else if m.currentScreen == ScreenMakeCommands {
				m.makeCommandList.CursorUp()
				if selectedItem := m.makeCommandList.SelectedItem(); selectedItem != nil {
					if cmd, ok := selectedItem.(MakeCommand); ok {
						m.selectedMakeCommand = cmd
					}
				}
			}

		case "tab":
			if m.currentScreen == ScreenEndpointList {
				// Switch to request builder panel
				m.currentScreen = ScreenRequestBuilder
				m.focusedInputIndex = 0
				m.updateInputFocus()
			} else if m.currentScreen == ScreenRequestBuilder {
				m.focusedInputIndex = (m.focusedInputIndex + 1) % len(m.allInputs)
				m.updateInputFocus()
			}

		case "shift+tab":
			if m.currentScreen == ScreenRequestBuilder {
				m.focusedInputIndex--
				if m.focusedInputIndex < 0 {
					m.focusedInputIndex = len(m.allInputs) // textarea is last
				}
				m.updateInputFocus()
			}

		case "j", "down":
			// Navigate down in make commands screen
			if m.currentScreen == ScreenMakeCommands {
				m.makeCommandList.CursorDown()
				if selectedItem := m.makeCommandList.SelectedItem(); selectedItem != nil {
					if cmd, ok := selectedItem.(MakeCommand); ok {
						m.selectedMakeCommand = cmd
					}
				}
			}

		case "k", "up":
			// Navigate up in make commands screen
			if m.currentScreen == ScreenMakeCommands {
				m.makeCommandList.CursorUp()
				if selectedItem := m.makeCommandList.SelectedItem(); selectedItem != nil {
					if cmd, ok := selectedItem.(MakeCommand); ok {
						m.selectedMakeCommand = cmd
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Fixed width for endpoint list panel
		leftWidth := 60
		if leftWidth > msg.Width-20 {
			leftWidth = msg.Width / 2
		}

		// Calculate list height accounting for border, title, footer
		listHeight := msg.Height - 8
		if listHeight < 5 {
			listHeight = 5
		}

		// Set endpoint list size to fixed width
		m.endpointList.SetSize(leftWidth-8, listHeight)

		// Set make command list size (same dimensions as endpoint list)
		m.makeCommandList.SetSize(leftWidth-8, listHeight)

		// Response viewport uses right panel width
		rightWidth := msg.Width - leftWidth
		m.responseViewport.Width = rightWidth - 8
		// Viewport height will be calculated dynamically based on actual config panel height
		// Set a reasonable default, will be recalculated at render time
		m.responseViewport.Height = msg.Height - 20
		if m.responseViewport.Height < 10 {
			m.responseViewport.Height = 10
		}

		// Make output viewport uses right panel width
		m.makeOutputViewport.Width = rightWidth - 8

	case requestMsg:
		m.currentResult = (*RequestResult)(&msg)
		m.history = append(m.history, *m.currentResult)
		// Stay on current screen, don't switch to separate response viewer
		m.responseViewport.SetContent(formatResponseBody(m.currentResult))
		m.requestInFlight = false
		m.isLoading = false
		return m, nil

	case errMsg:
		m.errorMsg = msg.err.Error()
		m.requestInFlight = false
		m.isLoading = false
		return m, nil

	case makeCommandResult:
		m.makeCommandOutput = msg.output
		m.commandRunning = false
		return m, nil

	case statsResult:
		m.templateCount = msg.templateCount
		m.chainCount = msg.chainCount
		m.statsLoading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		}
		return m, nil

	case chainsUpdatedMsg:
		m.cachedChains = msg.chains
		return m, nil

	case templatesUpdatedMsg:
		m.cachedTemplates = msg.templates
		return m, nil

	case tickMsg:
		// Periodic background data fetch
		return m, tea.Batch(
			m.fetchBackgroundData(),
			tickCmd(), // Schedule next tick
		)
	}

	// Update current screen's components
	var cmd tea.Cmd
	switch m.currentScreen {
	case ScreenEndpointList:
		// If in search mode, handle text input for searching
		if m.searchMode {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.Type {
				case tea.KeyBackspace:
					if len(m.searchBuffer) > 0 {
						m.searchBuffer = m.searchBuffer[:len(m.searchBuffer)-1]
						m.searchAndNavigate()
					}
				case tea.KeyRunes:
					m.searchBuffer += keyMsg.String()
					m.searchAndNavigate()
				}
			}
		} else {
			oldSelection := m.endpointList.Index()
			m.endpointList, cmd = m.endpointList.Update(msg)
			cmds = append(cmds, cmd)
			// Update selected endpoint when list selection changes
			if m.endpointList.Index() != oldSelection {
				// Save current state before switching
				m.saveCurrentEndpointState()
				if selectedItem := m.endpointList.SelectedItem(); selectedItem != nil {
					if item, ok := selectedItem.(endpointItem); ok {
						m.selectedEndpoint = item.endpoint
						cmd := m.prepareRequestBuilder()
						cmds = append(cmds, cmd)
					}
				}
			}
		}

	case ScreenRequestBuilder:
		if m.focusedInputIndex < len(m.allInputs) {
			m.allInputs[m.focusedInputIndex], cmd = m.allInputs[m.focusedInputIndex].Update(msg)
			cmds = append(cmds, cmd)
			// Sync the updated input back to the maps
			m.syncInputToMaps(m.focusedInputIndex)
		}
		// Also update response viewport for scrolling
		m.responseViewport, cmd = m.responseViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	var content string
	switch m.currentScreen {
	case ScreenEndpointList, ScreenRequestBuilder:
		// Always show split view with endpoint list and request/response panels
		content = renderSplitView(m)
	case ScreenHistory:
		content = renderHistory(m)
	case ScreenSettings:
		content = renderSettings(m)
	case ScreenMakeCommands:
		content = renderMakeScreen(m)
	case ScreenStatsModal:
		// Render previous screen in background, then overlay modal
		var bgContent string
		switch m.previousScreen {
		case ScreenMakeCommands:
			bgContent = renderMakeScreen(m)
		default:
			bgContent = renderSplitView(m)
		}
		content = renderStatsModal(m, bgContent)
	default:
		content = renderSplitView(m)
	}

	// Global keyboard shortcuts help line at bottom
	globalHelp := renderGlobalHelp(m.width)

	// Combine content with global help line
	mainContent := content
	if m.height > 1 {
		// Reserve one line for global help
		mainContentHeight := m.height - 1
		mainContent = lipgloss.Place(m.width, mainContentHeight, lipgloss.Left, lipgloss.Top, content)
		return lipgloss.JoinVertical(lipgloss.Left, mainContent, globalHelp)
	}

	// Use lipgloss.Place to properly position content in terminal
	if m.height > 0 && m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
	}
	return content
}

// handleEnter handles the enter key press
func (m Model) handleEnter() (Model, tea.Cmd) {
	switch m.currentScreen {
	case ScreenEndpointList:
		// Execute the request with default/current values
		return m, m.executeRequest()

	case ScreenRequestBuilder:
		// Execute the request
		return m, m.executeRequest()

	case ScreenMakeCommands:
		// Execute the selected make command
		if m.selectedMakeCommand.Name != "" {
			return m, m.executeMakeCommand()
		}
	}

	return m, nil
}

// executeMakeCommand runs the selected make command
func (m Model) executeMakeCommand() tea.Cmd {
	if m.selectedMakeCommand.Name == "" {
		return nil
	}

	commandName := m.selectedMakeCommand.Name

	return func() tea.Msg {
		// Run the make command and capture output
		// Command runs from current directory (launchpad root)
		cmd := exec.Command("make", commandName)
		output, err := cmd.CombinedOutput()

		result := string(output)
		if err != nil {
			result = fmt.Sprintf("Error executing 'make %s':\n%s\n\nOutput:\n%s",
				commandName, err.Error(), string(output))
		}

		return makeCommandResult{
			command: commandName,
			output:  result,
		}
	}
}

// makeCommandResult is sent when a make command completes
type makeCommandResult struct {
	command string
	output  string
}

// statsResult is sent when stats are fetched
type statsResult struct {
	templateCount int
	chainCount    int
	err           error
}

// fetchBackgroundData fetches chains and templates in the background
func (m Model) fetchBackgroundData() tea.Cmd {
	return tea.Batch(
		m.fetchChains(),
		m.fetchTemplates(),
	)
}

// fetchChains fetches chains from the API
func (m Model) fetchChains() tea.Cmd {
	return func() tea.Msg {
		chainsReq, _ := http.NewRequest("GET", m.baseURL+"/api/v1/chains", nil)
		chainsReq.Header.Set("X-User-ID", m.userID)
		chainsResp, err := http.DefaultClient.Do(chainsReq)
		if err == nil {
			defer chainsResp.Body.Close()
			var data struct {
				Data []CachedChain `json:"data"`
			}
			if err := json.NewDecoder(chainsResp.Body).Decode(&data); err == nil {
				return chainsUpdatedMsg{chains: data.Data}
			}
		}
		return nil
	}
}

// fetchTemplates fetches templates from the API
func (m Model) fetchTemplates() tea.Cmd {
	return func() tea.Msg {
		templatesResp, err := http.Get(m.baseURL + "/api/v1/templates")
		if err == nil {
			defer templatesResp.Body.Close()
			var data struct {
				Data []CachedTemplate `json:"data"`
			}
			if err := json.NewDecoder(templatesResp.Body).Decode(&data); err == nil {
				return templatesUpdatedMsg{templates: data.Data}
			}
		}
		return nil
	}
}

// fetchStats fetches template and chain counts from API
func (m Model) fetchStats() tea.Cmd {
	return func() tea.Msg {
		var result statsResult

		// Fetch templates count
		templatesResp, err := http.Get(m.baseURL + "/api/v1/templates")
		if err == nil {
			defer templatesResp.Body.Close()
			var data struct {
				Data []interface{} `json:"data"`
			}
			if err := json.NewDecoder(templatesResp.Body).Decode(&data); err == nil {
				result.templateCount = len(data.Data)
			}
		}

		// Fetch chains count
		chainsReq, _ := http.NewRequest("GET", m.baseURL+"/api/v1/chains", nil)
		chainsReq.Header.Set("X-User-ID", m.userID)
		chainsResp, err := http.DefaultClient.Do(chainsReq)
		if err == nil {
			defer chainsResp.Body.Close()
			var data struct {
				Data []interface{} `json:"data"`
			}
			if err := json.NewDecoder(chainsResp.Body).Decode(&data); err == nil {
				result.chainCount = len(data.Data)
			}
		}

		if err != nil {
			result.err = err
		}

		return result
	}
}

// saveCurrentEndpointState saves the current endpoint's state before switching
func (m *Model) saveCurrentEndpointState() {
	if m.selectedEndpoint.Name == "" {
		return
	}

	// Deep copy the input maps to preserve state
	pathCopy := make(map[string]textinput.Model)
	for k, v := range m.pathParamInputs {
		pathCopy[k] = v
	}

	queryCopy := make(map[string]textinput.Model)
	for k, v := range m.queryParamInputs {
		queryCopy[k] = v
	}

	bodyCopy := make(map[string]textinput.Model)
	for k, v := range m.bodyFieldInputs {
		bodyCopy[k] = v
	}

	m.endpointStates[m.selectedEndpoint.Name] = &EndpointState{
		pathParamInputs:  pathCopy,
		queryParamInputs: queryCopy,
		bodyFieldInputs:  bodyCopy,
		response:         m.currentResult,
	}
}

// prepareRequestBuilder sets up inputs for the selected endpoint
func (m *Model) prepareRequestBuilder() tea.Cmd {
	// Check if we have saved state for this endpoint
	if state, exists := m.endpointStates[m.selectedEndpoint.Name]; exists {
		// Restore saved state
		m.pathParamInputs = state.pathParamInputs
		m.queryParamInputs = state.queryParamInputs
		m.bodyFieldInputs = state.bodyFieldInputs
		m.currentResult = state.response
		if m.currentResult != nil {
			m.responseViewport.SetContent(formatResponseBody(m.currentResult))
		}

		// Rebuild allInputs array from saved inputs
		m.allInputs = []textinput.Model{}
		for _, param := range m.selectedEndpoint.PathParams {
			if input, ok := m.pathParamInputs[param.Name]; ok {
				m.allInputs = append(m.allInputs, input)
			}
		}
		for _, param := range m.selectedEndpoint.QueryParams {
			if input, ok := m.queryParamInputs[param.Name]; ok {
				m.allInputs = append(m.allInputs, input)
			}
		}
		for key := range m.bodyFieldInputs {
			if input, ok := m.bodyFieldInputs[key]; ok {
				m.allInputs = append(m.allInputs, input)
			}
		}
	} else {
		// Create new inputs with defaults
		m.pathParamInputs = make(map[string]textinput.Model)
		m.queryParamInputs = make(map[string]textinput.Model)
		m.allInputs = []textinput.Model{}
		m.currentResult = nil

		// Create inputs for path parameters
		for _, param := range m.selectedEndpoint.PathParams {
			ti := textinput.New()
			ti.Placeholder = param.Name
			ti.CharLimit = 100
			ti.Width = 40
			// Set initial value from cached data if available
			if param.InitialValue != nil {
				initialVal := param.InitialValue(m)
				if initialVal != "" {
					ti.SetValue(initialVal)
				}
			}
			m.pathParamInputs[param.Name] = ti
			m.allInputs = append(m.allInputs, ti)
		}

		// Create inputs for query parameters
		for _, param := range m.selectedEndpoint.QueryParams {
			ti := textinput.New()
			ti.Placeholder = param.Name
			ti.CharLimit = 100
			ti.Width = 40
			m.queryParamInputs[param.Name] = ti
			m.allInputs = append(m.allInputs, ti)
		}

		// Parse example body and create inputs for body fields
		m.bodyFieldInputs = make(map[string]textinput.Model)
		if m.selectedEndpoint.ExampleBody != "" {
			// Parse JSON to extract field names and default values
			var bodyMap map[string]interface{}
			if err := json.Unmarshal([]byte(m.selectedEndpoint.ExampleBody), &bodyMap); err == nil {
				for key, value := range bodyMap {
					ti := textinput.New()
					ti.Placeholder = key
					ti.CharLimit = 200
					ti.Width = 40
					// Set default value from example
					if value != nil {
						ti.SetValue(fmt.Sprintf("%v", value))
					}
					m.bodyFieldInputs[key] = ti
					m.allInputs = append(m.allInputs, ti)
				}
			}
		}
	}

	// Focus first input
	m.focusedInputIndex = 0
	m.updateInputFocus()

	return nil
}

// updateInputFocus updates which input is focused
func (m *Model) updateInputFocus() {
	for i := range m.allInputs {
		if i == m.focusedInputIndex {
			m.allInputs[i].Focus()
		} else {
			m.allInputs[i].Blur()
		}
	}
}

// syncInputToMaps syncs the input at the given index back to the appropriate map
func (m *Model) syncInputToMaps(index int) {
	if index >= len(m.allInputs) {
		return
	}

	currentIndex := 0
	// Check if it's a path param
	for _, param := range m.selectedEndpoint.PathParams {
		if currentIndex == index {
			m.pathParamInputs[param.Name] = m.allInputs[index]
			return
		}
		currentIndex++
	}

	// Check if it's a query param
	for _, param := range m.selectedEndpoint.QueryParams {
		if currentIndex == index {
			m.queryParamInputs[param.Name] = m.allInputs[index]
			return
		}
		currentIndex++
	}

	// Must be a body field - find which one by matching the input
	bodyStartIndex := len(m.selectedEndpoint.PathParams) + len(m.selectedEndpoint.QueryParams)
	if index >= bodyStartIndex {
		// Find the key by iterating through body fields
		bodyIndex := index - bodyStartIndex
		i := 0
		for key := range m.bodyFieldInputs {
			if i == bodyIndex {
				m.bodyFieldInputs[key] = m.allInputs[index]
				return
			}
			i++
		}
	}
}

// executeRequest executes the current request
func (m Model) executeRequest() tea.Cmd {
	return func() tea.Msg {
		// Collect path param values
		pathParamValues := make(map[string]string)
		for param, input := range m.pathParamInputs {
			pathParamValues[param] = input.Value()
		}

		// Collect query param values
		queryParamValues := make(map[string]string)
		for param, input := range m.queryParamInputs {
			queryParamValues[param] = input.Value()
		}

		// Build request body from body field inputs
		var requestBody string
		if len(m.bodyFieldInputs) > 0 {
			bodyMap := make(map[string]interface{})
			for key, input := range m.bodyFieldInputs {
				value := input.Value()
				if value != "" {
					// Try to parse as JSON first, otherwise use as string
					var jsonValue interface{}
					if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
						bodyMap[key] = jsonValue
					} else {
						bodyMap[key] = value
					}
				}
			}
			if bodyBytes, err := json.Marshal(bodyMap); err == nil {
				requestBody = string(bodyBytes)
			}
		}

		// Execute request
		result := ExecuteRequest(
			m.baseURL,
			m.userID,
			m.selectedEndpoint,
			pathParamValues,
			requestBody,
			queryParamValues,
		)

		return requestMsg(result)
	}
}

// goBack navigates to the previous screen
func (m Model) goBack() Model {
	switch m.currentScreen {
	case ScreenRequestBuilder:
		m.currentScreen = ScreenEndpointList
	case ScreenHistory:
		m.currentScreen = ScreenEndpointList
	case ScreenSettings:
		m.currentScreen = ScreenEndpointList
	}
	m.errorMsg = ""
	return m
}

// formatResponseBody formats the response body for display
func formatResponseBody(result *RequestResult) string {
	if result == nil {
		return "No response"
	}

	if result.Error != nil {
		return fmt.Sprintf("Error: %v", result.Error)
	}

	return result.Body
}

// searchAndNavigate searches through endpoints and navigates to the first match
func (m *Model) searchAndNavigate() {
	if m.searchBuffer == "" {
		return
	}

	searchLower := strings.ToLower(m.searchBuffer)

	// Search through all endpoints
	for i, ep := range m.endpoints {
		// Check if any part matches the search
		categoryMatch := strings.Contains(strings.ToLower(string(ep.Category)), searchLower)
		methodMatch := strings.Contains(strings.ToLower(string(ep.Method)), searchLower)
		nameMatch := strings.Contains(strings.ToLower(ep.Name), searchLower)

		if categoryMatch || methodMatch || nameMatch {
			// Save current state before switching
			m.saveCurrentEndpointState()
			// Navigate to this item
			m.endpointList.Select(i)
			// Update selected endpoint
			m.selectedEndpoint = ep
			m.prepareRequestBuilder()
			return
		}
	}
}

// List item types

type endpointItem struct {
	endpoint Endpoint
	index    int
}

func (i endpointItem) Title() string {
	return fmt.Sprintf("[%s] %s %s", i.endpoint.Category, i.endpoint.Method, i.endpoint.Name)
}
func (i endpointItem) Description() string { return i.endpoint.Description }
func (i endpointItem) FilterValue() string {
	return fmt.Sprintf("%s %s %s", i.endpoint.Category, i.endpoint.Method, i.endpoint.Name)
}
