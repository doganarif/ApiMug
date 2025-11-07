package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/doganarif/ApiMug/internal/api"
)

type viewMode int

const (
	viewList viewMode = iota
	viewDetail
	viewRequest
	viewResponse
	viewAuth
	viewSettings
)

type responseMsg struct {
	response *api.Response
}

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Select   key.Binding
	Back     key.Binding
	Quit     key.Binding
	Server   key.Binding
	Settings key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Server: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "configure auth"),
	),
	Settings: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "settings"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type item struct {
	endpoint api.Endpoint
}

func (i item) FilterValue() string {
	return i.endpoint.Path + " " + i.endpoint.Method
}

func (i item) Title() string {
	style := getMethodStyle(i.endpoint.Method)
	return fmt.Sprintf("%s %s", style.Render(strings.ToUpper(i.endpoint.Method)), i.endpoint.Path)
}

func (i item) Description() string {
	if i.endpoint.Summary != "" {
		return i.endpoint.Summary
	}
	return i.endpoint.Description
}

type Model struct {
	spec           *api.Spec
	authMgr        *api.AuthManager
	client         *api.Client
	list           list.Model
	mode           viewMode
	selected       *api.Endpoint
	response       *api.Response
	width          int
	height         int
	err            error

	// Request form state
	paramInputs    map[string]*InputField
	bodyInput      textarea.Model
	focusedInput   int
	baseURL        string

	// Auth state
	authInputs     map[string]*InputField
	authSchemes    []string
	selectedScheme int

	// Settings state
	settingsInputs map[string]*InputField
	port           int
	onSettingsChange func(baseURL string, port int)
}

func NewModel(spec *api.Spec, baseURL string, port int, onSettingsChange func(string, int)) Model {
	endpoints := spec.GetEndpoints()
	items := make([]list.Item, len(endpoints))
	for i, ep := range endpoints {
		items[i] = item{endpoint: ep}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	title, version, _ := spec.GetInfo()
	l.Title = fmt.Sprintf("%s (v%s)", title, version)
	l.SetShowStatusBar(false)

	authMgr := api.NewAuthManager(spec)

	bodyInput := textarea.New()
	bodyInput.Placeholder = `{"key": "value"}`
	bodyInput.CharLimit = 5000

	return Model{
		spec:             spec,
		authMgr:          authMgr,
		list:             l,
		mode:             viewList,
		baseURL:          baseURL,
		bodyInput:        bodyInput,
		paramInputs:      make(map[string]*InputField),
		authInputs:       make(map[string]*InputField),
		authSchemes:      authMgr.GetAvailableAuthSchemes(),
		settingsInputs:   make(map[string]*InputField),
		port:             port,
		onSettingsChange: onSettingsChange,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case responseMsg:
		m.response = msg.response
		m.mode = viewResponse
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m.updateCurrentView(msg)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case viewList:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Select):
			if i, ok := m.list.SelectedItem().(item); ok {
				m.selected = &i.endpoint
				m.mode = viewDetail
			}
			return m, nil
		case key.Matches(msg, keys.Server):
			m.mode = viewAuth
			m.initAuthInputs()
			return m, nil
		case key.Matches(msg, keys.Settings):
			m.mode = viewSettings
			m.initSettingsInputs()
			return m, nil
		}

	case viewDetail:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Back):
			m.mode = viewList
			m.selected = nil
			return m, nil
		case key.Matches(msg, keys.Select):
			m.mode = viewRequest
			m.initRequestInputs()
			return m, nil
		}

	case viewRequest:
		switch msg.String() {
		case "esc":
			m.mode = viewDetail
			return m, nil
		case "ctrl+s":
			return m, m.sendRequest()
		case "tab", "shift+tab":
			m.cycleFocus(msg.String() == "shift+tab")
			return m, nil
		}

	case viewResponse:
		switch {
		case key.Matches(msg, keys.Back):
			m.mode = viewRequest
			return m, nil
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}

	case viewAuth:
		switch msg.String() {
		case "esc":
			m.mode = viewList
			return m, nil
		case "ctrl+s":
			m.applyAuth()
			m.mode = viewList
			return m, nil
		case "tab", "shift+tab":
			m.cycleFocus(msg.String() == "shift+tab")
			return m, nil
		case "up", "down":
			if msg.String() == "up" && m.selectedScheme > 0 {
				m.selectedScheme--
				m.initAuthInputs()
			} else if msg.String() == "down" && m.selectedScheme < len(m.authSchemes)-1 {
				m.selectedScheme++
				m.initAuthInputs()
			}
			return m, nil
		}

	case viewSettings:
		switch msg.String() {
		case "esc":
			m.mode = viewList
			return m, nil
		case "ctrl+s":
			m.applySettings()
			m.mode = viewList
			return m, nil
		case "tab", "shift+tab":
			m.cycleFocus(msg.String() == "shift+tab")
			return m, nil
		}
	}

	return m.updateCurrentView(msg)
}

func (m Model) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.mode {
	case viewList:
		m.list, cmd = m.list.Update(msg)

	case viewRequest:
		if m.focusedInput == len(m.paramInputs) && m.selected.HasBody {
			m.bodyInput, cmd = m.bodyInput.Update(msg)
		} else {
			idx := 0
			for _, input := range m.paramInputs {
				if idx == m.focusedInput {
					cmd = input.Update(msg)
					break
				}
				idx++
			}
		}

	case viewAuth:
		idx := 0
		for _, input := range m.authInputs {
			if idx == m.focusedInput {
				cmd = input.Update(msg)
				break
			}
			idx++
		}

	case viewSettings:
		idx := 0
		for _, input := range m.settingsInputs {
			if idx == m.focusedInput {
				cmd = input.Update(msg)
				break
			}
			idx++
		}
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.mode {
	case viewList:
		return m.listView()
	case viewDetail:
		return m.detailView()
	case viewRequest:
		return m.requestView()
	case viewResponse:
		return m.responseView()
	case viewAuth:
		return m.authView()
	case viewSettings:
		return m.settingsView()
	}
	return ""
}

func (m Model) listView() string {
	help := helpStyle.Render("\n↑/↓: navigate • enter: view details • s: auth • c: settings • q: quit")
	return m.list.View() + help
}

func (m Model) detailView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Endpoint Details"))
	b.WriteString("\n\n")

	style := getMethodStyle(m.selected.Method)
	b.WriteString(fmt.Sprintf("%s %s\n\n", style.Render(strings.ToUpper(m.selected.Method)), m.selected.Path))

	if m.selected.Summary != "" {
		b.WriteString(headerStyle.Render("Summary"))
		b.WriteString("\n")
		b.WriteString(m.selected.Summary)
		b.WriteString("\n\n")
	}

	if m.selected.Description != "" {
		b.WriteString(headerStyle.Render("Description"))
		b.WriteString("\n")
		b.WriteString(m.selected.Description)
		b.WriteString("\n\n")
	}

	if len(m.selected.Parameters) > 0 {
		b.WriteString(headerStyle.Render("Parameters"))
		b.WriteString("\n")
		for _, p := range m.selected.Parameters {
			req := ""
			if p.Required {
				req = " (required)"
			}
			b.WriteString(fmt.Sprintf("  • %s [%s]%s: %s\n", p.Name, p.In, req, p.Description))
		}
		b.WriteString("\n")
	}

	if m.selected.HasBody {
		b.WriteString(headerStyle.Render("Request Body"))
		b.WriteString("\n")
		b.WriteString("  application/json\n\n")
	}

	b.WriteString(helpStyle.Render("\nenter: send request • esc: back • q: quit"))

	return b.String()
}

func (m *Model) initRequestInputs() {
	m.paramInputs = make(map[string]*InputField)
	m.focusedInput = 0

	for _, param := range m.selected.Parameters {
		field := NewInputField(
			fmt.Sprintf("%s (%s)", param.Name, param.In),
			param.Example,
			param.Required,
		)
		m.paramInputs[param.Name] = &field
	}

	// Focus first input
	if len(m.paramInputs) > 0 {
		idx := 0
		for _, input := range m.paramInputs {
			if idx == 0 {
				input.Focus()
			}
			idx++
		}
	} else if m.selected.HasBody {
		m.bodyInput.Focus()
	}
}

func (m *Model) cycleFocus(reverse bool) {
	totalInputs := len(m.paramInputs)
	if m.selected.HasBody {
		totalInputs++
	}

	if totalInputs == 0 {
		return
	}

	// Blur current
	if m.focusedInput == len(m.paramInputs) {
		m.bodyInput.Blur()
	} else {
		idx := 0
		for _, input := range m.paramInputs {
			if idx == m.focusedInput {
				input.Blur()
				break
			}
			idx++
		}
	}

	// Update focus index
	if reverse {
		m.focusedInput--
		if m.focusedInput < 0 {
			m.focusedInput = totalInputs - 1
		}
	} else {
		m.focusedInput++
		if m.focusedInput >= totalInputs {
			m.focusedInput = 0
		}
	}

	// Focus new
	if m.focusedInput == len(m.paramInputs) {
		m.bodyInput.Focus()
	} else {
		idx := 0
		for _, input := range m.paramInputs {
			if idx == m.focusedInput {
				input.Focus()
				break
			}
			idx++
		}
	}
}

func (m Model) requestView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Send Request"))
	b.WriteString("\n\n")

	style := getMethodStyle(m.selected.Method)
	b.WriteString(fmt.Sprintf("%s %s\n\n", style.Render(strings.ToUpper(m.selected.Method)), m.selected.Path))

	if len(m.paramInputs) > 0 {
		b.WriteString(headerStyle.Render("Parameters"))
		b.WriteString("\n\n")
		for _, input := range m.paramInputs {
			b.WriteString(input.View())
			b.WriteString("\n")
		}
	}

	if m.selected.HasBody {
		b.WriteString("\n")
		b.WriteString(headerStyle.Render("Request Body (JSON)"))
		b.WriteString("\n\n")
		b.WriteString(m.bodyInput.View())
	}

	b.WriteString(helpStyle.Render("\n\ntab: next field • ctrl+s: send • esc: back"))

	return b.String()
}

func (m *Model) sendRequest() tea.Cmd {
	return func() tea.Msg {
		// Build request
		req := &api.Request{
			Method:      strings.ToUpper(m.selected.Method),
			Path:        m.selected.Path,
			QueryParams: make(map[string]string),
			Headers:     make(map[string]string),
		}

		// Apply parameters
		for name, input := range m.paramInputs {
			val := input.Value()
			if val == "" {
				continue
			}

			// Find param type
			for _, p := range m.selected.Parameters {
				if p.Name == name {
					switch p.In {
					case "query":
						req.QueryParams[name] = val
					case "header":
						req.Headers[name] = val
					case "path":
						req.Path = strings.ReplaceAll(req.Path, "{"+name+"}", val)
					}
					break
				}
			}
		}

		// Apply body
		if m.selected.HasBody {
			req.Body = m.bodyInput.Value()
			req.ContentType = "application/json"
		}

		// Ensure client is initialized
		if m.client == nil {
			m.client = api.NewClient(m.baseURL, m.authMgr)
		}

		// Send request
		resp := m.client.Send(context.Background(), req)

		return responseMsg{response: resp}
	}
}

func (m Model) responseView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Response"))
	b.WriteString("\n\n")

	if m.response.Error != nil {
		b.WriteString(errorStyle.Render("Error: "))
		b.WriteString(m.response.Error.Error())
		b.WriteString("\n\n")
	} else {
		// Status
		statusStyle := statusCodeSuccessStyle
		if m.response.StatusCode >= 400 {
			statusStyle = statusCodeErrorStyle
		}
		b.WriteString(statusStyle.Render(fmt.Sprintf("%d %s", m.response.StatusCode, m.response.Status)))
		b.WriteString(infoStyle.Render(fmt.Sprintf("  (%s)", m.response.Duration)))
		b.WriteString("\n\n")

		// Headers
		b.WriteString(headerStyle.Render("Headers"))
		b.WriteString("\n")
		for k, v := range m.response.Headers {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, strings.Join(v, ", ")))
		}
		b.WriteString("\n")

		// Body
		b.WriteString(headerStyle.Render("Body"))
		b.WriteString("\n\n")
		b.WriteString(codeStyle.Render(m.response.FormatResponseBody()))
	}

	b.WriteString(helpStyle.Render("\n\nesc: back • q: quit"))

	return b.String()
}

func (m *Model) initAuthInputs() {
	m.authInputs = make(map[string]*InputField)
	m.focusedInput = 0

	schemeName := m.authSchemes[m.selectedScheme]
	if schemeName == "none" {
		return
	}

	config, err := m.authMgr.ParseAuthScheme(schemeName)
	if err != nil {
		return
	}

	switch config.Type {
	case api.AuthTypeBearer, api.AuthTypeOAuth2:
		field := NewInputField("Token", "Enter token", true)
		field.Focus()
		m.authInputs["token"] = &field

	case api.AuthTypeAPIKey:
		field := NewInputField(fmt.Sprintf("API Key (%s in %s)", config.KeyName, config.APIKeyIn), "Enter API key", true)
		field.Focus()
		m.authInputs["apikey"] = &field

	case api.AuthTypeBasic:
		userField := NewInputField("Username", "Enter username", true)
		userField.Focus()
		m.authInputs["username"] = &userField

		passField := NewInputField("Password", "Enter password", true)
		m.authInputs["password"] = &passField
	}
}

func (m *Model) applyAuth() {
	schemeName := m.authSchemes[m.selectedScheme]
	if schemeName == "none" {
		m.authMgr.SetAuth(&api.AuthConfig{Type: api.AuthTypeNone})
		return
	}

	config, err := m.authMgr.ParseAuthScheme(schemeName)
	if err != nil {
		return
	}

	switch config.Type {
	case api.AuthTypeBearer, api.AuthTypeOAuth2:
		if input, ok := m.authInputs["token"]; ok {
			config.Token = input.Value()
		}

	case api.AuthTypeAPIKey:
		if input, ok := m.authInputs["apikey"]; ok {
			config.APIKey = input.Value()
		}

	case api.AuthTypeBasic:
		if input, ok := m.authInputs["username"]; ok {
			config.Username = input.Value()
		}
		if input, ok := m.authInputs["password"]; ok {
			config.Password = input.Value()
		}
	}

	m.authMgr.SetAuth(config)
}

func (m Model) authView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Configure Authentication"))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render("Available Schemes"))
	b.WriteString("\n\n")

	for i, scheme := range m.authSchemes {
		if i == m.selectedScheme {
			b.WriteString(selectedStyle.Render("► " + scheme))
		} else {
			b.WriteString("  " + scheme)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if m.selectedScheme < len(m.authSchemes) && m.authSchemes[m.selectedScheme] != "none" {
		b.WriteString(headerStyle.Render("Configuration"))
		b.WriteString("\n\n")

		for _, input := range m.authInputs {
			b.WriteString(input.View())
			b.WriteString("\n")
		}
	}

	b.WriteString(helpStyle.Render("\n\n↑/↓: select scheme • tab: next field • ctrl+s: save • esc: cancel"))

	return b.String()
}

func (m *Model) initSettingsInputs() {
	m.settingsInputs = make(map[string]*InputField)
	m.focusedInput = 0

	baseURLField := NewInputField("Base URL", "https://api.example.com", false)
	baseURLField.SetValue(m.baseURL)
	baseURLField.Focus()
	m.settingsInputs["baseURL"] = &baseURLField

	portField := NewInputField("Swagger UI Port", "8080", false)
	portField.SetValue(fmt.Sprintf("%d", m.port))
	m.settingsInputs["port"] = &portField
}

func (m *Model) applySettings() {
	if urlInput, ok := m.settingsInputs["baseURL"]; ok {
		newBaseURL := urlInput.Value()
		if newBaseURL != "" {
			m.baseURL = newBaseURL
			// Recreate client with new base URL
			m.client = api.NewClient(m.baseURL, m.authMgr)
		}
	}

	newPort := m.port
	if portInput, ok := m.settingsInputs["port"]; ok {
		portStr := portInput.Value()
		if portStr != "" {
			fmt.Sscanf(portStr, "%d", &newPort)
		}
	}

	// Call the callback to notify about settings change
	if m.onSettingsChange != nil && (m.port != newPort || m.baseURL != m.settingsInputs["baseURL"].Value()) {
		m.onSettingsChange(m.baseURL, newPort)
		m.port = newPort
	}
}

func (m Model) settingsView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Settings"))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render("API Configuration"))
	b.WriteString("\n\n")

	// Maintain order: baseURL then port
	if input, ok := m.settingsInputs["baseURL"]; ok {
		b.WriteString(input.View())
		b.WriteString("\n")
	}

	if input, ok := m.settingsInputs["port"]; ok {
		b.WriteString(input.View())
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("\n\ntab: next field • ctrl+s: save • esc: cancel"))

	return b.String()
}
