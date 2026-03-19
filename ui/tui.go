package ui

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"ship-cli/api"
	"ship-cli/config"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

var (
	// Colors from ship-console tailwind.config.js
	colorPrimary   = lipgloss.Color("#ACFF00") // Neon Green
	colorSecondary = lipgloss.Color("#FF52B5") // Neon Pink
	colorAccent    = lipgloss.Color("#000000") // Black
	colorText      = lipgloss.Color("#FFFFFF") // White text
	colorGray      = lipgloss.Color("#4B5563") // Gray for inactive/borders

	docStyle = lipgloss.NewStyle().Margin(1, 2)
	
	titleStyle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Background(colorPrimary).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	statusStyle = lipgloss.NewStyle().Foreground(colorGray)
	errorStyle  = lipgloss.NewStyle().Foreground(colorSecondary).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)

	// Custom list styles
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(colorPrimary).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorPrimary)

	logoutBinding = key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "log out"),
	)
)

type state int

const (
	stateInputToken state = iota
	stateLoadingProjects
	stateSelectProject
	stateLoadingApps
	stateSelectApp
	stateSelectAction
	stateInputLocalPort
	stateInputTargetPort
	stateViewPortForward
	stateError
)

type item struct {
	title, desc string
	id          string
	data        interface{}
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type PortForwardSession struct {
	AppID      string
	AppName    string
	LocalPort  int
	TargetPort int
	Listener   net.Listener
	Cancel     context.CancelFunc
}

type Model struct {
	state       state
	client      *api.Client
	list        list.Model
	textInput   textinput.Model
	
	projects    []api.Project
	apps        []api.Application
	
	selectedProject api.Project
	selectedApp     api.Application
	
	localPort   int
	targetPort  int
	
	err         error
	statusMsg   string
	
	activeForwards map[string]*PortForwardSession
}

func NewModel(client *api.Client) Model {
	ti := textinput.New()
	ti.Focus()
	ti.PromptStyle = lipgloss.NewStyle().Foreground(colorPrimary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(colorText)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(colorSecondary)
	
	initialState := stateLoadingProjects
	if client.Token == "" {
		initialState = stateInputToken
		ti.Placeholder = "ship_pat_..."
		ti.CharLimit = 100
		ti.Width = 50
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	} else {
		ti.CharLimit = 5
		ti.Width = 20
	}

	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = selectedItemStyle.Copy().Bold(true)
	d.Styles.SelectedDesc = selectedItemStyle.Copy().Foreground(colorGray)
	d.Styles.NormalTitle = itemStyle.Copy().Foreground(colorText)
	d.Styles.NormalDesc = itemStyle.Copy().Foreground(colorGray)

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Loading..."
	l.Styles.Title = titleStyle
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(colorPrimary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(colorSecondary)
	
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{logoutBinding}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{logoutBinding}
	}

	return Model{
		state:          initialState,
		client:         client,
		list:           l,
		textInput:      ti,
		activeForwards: make(map[string]*PortForwardSession),
	}
}

func (m Model) Init() tea.Cmd {
	if m.state == stateInputToken {
		return textinput.Blink
	}
	return fetchProjects(m.client)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, logoutBinding) && m.state != stateInputToken && m.state != stateInputLocalPort && m.state != stateInputTargetPort {
			// Clean up forwards
			for _, session := range m.activeForwards {
				session.Cancel()
				session.Listener.Close()
			}
			m.activeForwards = make(map[string]*PortForwardSession)

			cfg, err := config.LoadConfig()
			if err == nil {
				cfg.Token = ""
				_ = config.SaveConfig(cfg)
			}
			m.client.Token = ""
			m.state = stateInputToken
			m.textInput.Placeholder = "ship_pat_..."
			m.textInput.SetValue("")
			m.textInput.EchoMode = textinput.EchoPassword
			m.textInput.EchoCharacter = '•'
			m.textInput.CharLimit = 100
			m.textInput.Width = 50
			m.textInput.Focus()
			return m, textinput.Blink
		}

		if msg.String() == "ctrl+c" {
			for _, session := range m.activeForwards {
				session.Cancel()
				session.Listener.Close()
			}
			return m, tea.Quit
		}
		
		if msg.String() == "esc" || msg.String() == "q" {
			if m.state == stateSelectProject || m.state == stateError || m.state == stateInputToken {
				for _, session := range m.activeForwards {
					session.Cancel()
					session.Listener.Close()
				}
				return m, tea.Quit
			}
			if m.state == stateViewPortForward {
				m.state = stateSelectAction
				m.list.Title = "Select Action"
				m.list.SetItems(actionItems(m.selectedApp.ID, m.activeForwards))
				return m, nil
			}
			if m.state == stateSelectApp {
				m.state = stateSelectProject
				m.list.Title = "Select Project"
				m.list.SetItems(projectsToItems(m.projects))
				return m, nil
			}
			if m.state == stateSelectAction {
				m.state = stateSelectApp
				m.list.Title = fmt.Sprintf("Applications in %s", m.selectedProject.Name)
				m.list.SetItems(appsToItems(m.apps, m.activeForwards))
				return m, nil
			}
			if m.state == stateInputLocalPort || m.state == stateInputTargetPort {
				m.state = stateSelectAction
				m.list.Title = "Select Action"
				m.list.SetItems(actionItems(m.selectedApp.ID, m.activeForwards))
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-2)
		return m, nil

	case projectsMsg:
		m.projects = msg.projects
		m.state = stateSelectProject
		m.list.Title = "Select Project"
		m.list.SetItems(projectsToItems(m.projects))
		return m, nil

	case appsMsg:
		m.apps = msg.apps
		m.state = stateSelectApp
		m.list.Title = fmt.Sprintf("Applications in %s", m.selectedProject.Name)
		m.list.SetItems(appsToItems(m.apps, m.activeForwards))
		return m, nil

	case errMsg:
		m.err = msg.err
		m.state = stateError
		return m, nil

	case portForwardErrMsg:
		m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		m.state = stateSelectApp
		m.list.Title = fmt.Sprintf("Applications in %s", m.selectedProject.Name)
		m.list.SetItems(appsToItems(m.apps, m.activeForwards))
		return m, nil
		
	case portForwardStartedMsg:
		m.activeForwards[msg.session.AppID] = msg.session
		m.statusMsg = fmt.Sprintf("Started forwarding %s to localhost:%d", msg.session.AppName, msg.session.LocalPort)
		m.state = stateSelectApp
		m.list.Title = fmt.Sprintf("Applications in %s", m.selectedProject.Name)
		m.list.SetItems(appsToItems(m.apps, m.activeForwards))
		return m, nil
	}

	var cmd tea.Cmd

	switch m.state {
	case stateInputToken:
		m.textInput, cmd = m.textInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			token := strings.TrimSpace(m.textInput.Value())
			if token != "" {
				m.client.Token = token
				
				cfg, err := config.LoadConfig()
				if err == nil {
					cfg.Token = token
					_ = config.SaveConfig(cfg)
				}

				m.state = stateLoadingProjects
				m.textInput.SetValue("")
				m.textInput.EchoMode = textinput.EchoNormal
				m.textInput.CharLimit = 5
				m.textInput.Width = 20
				return m, fetchProjects(m.client)
			}
		}
		return m, cmd

	case stateSelectProject:
		m.list, cmd = m.list.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			if i, ok := m.list.SelectedItem().(item); ok {
				m.selectedProject = i.data.(api.Project)
				m.state = stateLoadingApps
				m.list.Title = "Loading Applications..."
				m.list.SetItems(nil)
				return m, fetchApps(m.client, m.selectedProject.ID)
			}
		}
		return m, cmd

	case stateSelectApp:
		m.list, cmd = m.list.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			if i, ok := m.list.SelectedItem().(item); ok {
				m.selectedApp = i.data.(api.Application)
				m.state = stateSelectAction
				m.list.Title = fmt.Sprintf("Action for %s", m.selectedApp.Name)
				m.list.SetItems(actionItems(m.selectedApp.ID, m.activeForwards))
			}
		}
		return m, cmd

	case stateSelectAction:
		m.list, cmd = m.list.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			if i, ok := m.list.SelectedItem().(item); ok {
				if i.id == "start-pf" {
					m.state = stateInputLocalPort
					m.textInput.Placeholder = "e.g. 5432"
					m.textInput.SetValue("")
					m.textInput.Focus()
				} else if i.id == "stop-pf" {
					if session, exists := m.activeForwards[m.selectedApp.ID]; exists {
						session.Cancel()
						session.Listener.Close()
						delete(m.activeForwards, m.selectedApp.ID)
						m.statusMsg = fmt.Sprintf("Stopped forwarding %s", session.AppName)
						m.state = stateSelectApp
						m.list.Title = fmt.Sprintf("Applications in %s", m.selectedProject.Name)
						m.list.SetItems(appsToItems(m.apps, m.activeForwards))
					}
				} else if i.id == "view-pf" {
					m.state = stateViewPortForward
				}
			}
		}
		return m, cmd

	case stateInputLocalPort:
		m.textInput, cmd = m.textInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			port, err := strconv.Atoi(m.textInput.Value())
			if err == nil && port > 0 && port < 65536 {
				m.localPort = port
				m.state = stateInputTargetPort
				m.textInput.Placeholder = "e.g. 80 or 5432"
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}
		return m, cmd

	case stateInputTargetPort:
		m.textInput, cmd = m.textInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			port, err := strconv.Atoi(m.textInput.Value())
			if err == nil && port > 0 && port < 65536 {
				m.targetPort = port
				ctx, cancel := context.WithCancel(context.Background())
				return m, startPortForward(ctx, cancel, m.client, m.selectedApp, m.localPort, m.targetPort)
			}
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return docStyle.Render(errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to quit.", m.err)))
	}

	switch m.state {
	case stateInputToken:
		return docStyle.Render(
			titleStyle.Render("Authentication") + "\n\n" +
			"Please enter your Personal Access Token (PAT):\n" +
			m.textInput.View() + "\n\n" +
			statusStyle.Render("(esc to quit)"),
		)
	case stateLoadingProjects:
		return docStyle.Render("Loading projects...")
	case stateLoadingApps:
		return docStyle.Render("Loading applications...")
	case stateSelectProject, stateSelectApp, stateSelectAction:
		view := m.list.View()
		if m.statusMsg != "" && m.state == stateSelectApp {
			if strings.HasPrefix(m.statusMsg, "Error:") {
				view += "\n" + errorStyle.Render(m.statusMsg)
			} else {
				view += "\n" + successStyle.Render(m.statusMsg)
			}
		}
		return docStyle.Render(view)
	case stateInputLocalPort:
		return docStyle.Render(
			titleStyle.Render("Port Forwarding") + "\n\n" +
			fmt.Sprintf("App: %s\n\n", lipgloss.NewStyle().Foreground(colorPrimary).Render(m.selectedApp.Name)) +
			"Enter Local Port:\n" +
			m.textInput.View() + "\n\n" +
			statusStyle.Render("(esc to go back)"),
		)
	case stateInputTargetPort:
		return docStyle.Render(
			titleStyle.Render("Port Forwarding") + "\n\n" +
			fmt.Sprintf("App: %s\n", lipgloss.NewStyle().Foreground(colorPrimary).Render(m.selectedApp.Name)) +
			fmt.Sprintf("Local Port: %s\n\n", lipgloss.NewStyle().Foreground(colorPrimary).Render(strconv.Itoa(m.localPort))) +
			"Enter Target Port (Pod Port):\n" +
			m.textInput.View() + "\n\n" +
			statusStyle.Render("(esc to go back)"),
		)
	case stateViewPortForward:
		session := m.activeForwards[m.selectedApp.ID]
		if session == nil {
			return docStyle.Render("No active port-forward session found.\n\n" + statusStyle.Render("(esc to go back)"))
		}
		return docStyle.Render(
			titleStyle.Render("Port Forward Details") + "\n\n" +
			fmt.Sprintf("App: %s\n", lipgloss.NewStyle().Foreground(colorPrimary).Render(session.AppName)) +
			fmt.Sprintf("Status: %s\n", successStyle.Render("Active ⚡")) +
			fmt.Sprintf("Forwarding: %s -> %s\n\n", 
				lipgloss.NewStyle().Foreground(colorPrimary).Render(fmt.Sprintf("127.0.0.1:%d", session.LocalPort)),
				lipgloss.NewStyle().Foreground(colorSecondary).Render(fmt.Sprintf("pod:%d", session.TargetPort)),
			) +
			"You can connect to this application using the local port above.\n\n" +
			statusStyle.Render("Press 'esc' to go back."),
		)
	}

	return "Unknown state"
}

// --- Messages & Commands ---

type projectsMsg struct{ projects []api.Project }
type appsMsg struct{ apps []api.Application }
type errMsg struct{ err error }
type portForwardErrMsg struct{ err error }
type portForwardStartedMsg struct{ 
	session *PortForwardSession
}

func fetchProjects(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.GetProjects()
		if err != nil {
			return errMsg{err}
		}
		return projectsMsg{projects}
	}
}

func fetchApps(client *api.Client, projectID string) tea.Cmd {
	return func() tea.Msg {
		apps, err := client.GetApplications(projectID)
		if err != nil {
			return errMsg{err}
		}
		return appsMsg{apps}
	}
}

func projectsToItems(projects []api.Project) []list.Item {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = item{title: p.Name, desc: "ID: " + p.ID, id: p.ID, data: p}
	}
	return items
}

func appsToItems(apps []api.Application, activeForwards map[string]*PortForwardSession) []list.Item {
	items := make([]list.Item, len(apps))
	for i, a := range apps {
		title := a.Name
		desc := fmt.Sprintf("Status: %s | Type: %s", a.Status, a.Type)
		
		if session, ok := activeForwards[a.ID]; ok {
			title = fmt.Sprintf("⚡ %s", title)
			desc = fmt.Sprintf("%s | Fwd: :%d->:%d", desc, session.LocalPort, session.TargetPort)
		}
		
		items[i] = item{title: title, desc: desc, id: a.ID, data: a}
	}
	return items
}

func actionItems(appID string, activeForwards map[string]*PortForwardSession) []list.Item {
	if _, exists := activeForwards[appID]; exists {
		return []list.Item{
			item{title: "View Details", desc: "View port-forward connection details", id: "view-pf"},
			item{title: "Stop Port Forward", desc: "Close the active port-forward tunnel", id: "stop-pf"},
		}
	}
	return []list.Item{
		item{title: "Start Port Forward", desc: "Forward local port to pod", id: "start-pf"},
	}
}

func startPortForward(ctx context.Context, cancel context.CancelFunc, client *api.Client, app api.Application, localPort, targetPort int) tea.Cmd {
	return func() tea.Msg {
		listenAddr := fmt.Sprintf("localhost:%d", localPort)
		l, err := net.Listen("tcp", listenAddr)
		if err != nil {
			cancel()
			return portForwardErrMsg{fmt.Errorf("failed to listen on %d: %v", localPort, err)}
		}
		
		go func() {
			<-ctx.Done()
			l.Close()
		}()

		go func() {
			for {
				conn, err := l.Accept()
				if err != nil {
					return // closed
				}
				go handleConnection(conn, client, app.ID, targetPort)
			}
		}()

		session := &PortForwardSession{
			AppID:      app.ID,
			AppName:    app.Name,
			LocalPort:  localPort,
			TargetPort: targetPort,
			Listener:   l,
			Cancel:     cancel,
		}

		return portForwardStartedMsg{session: session}
	}
}

func handleConnection(localConn net.Conn, client *api.Client, appID string, targetPort int) {
	defer localConn.Close()

	wsBase := "wss://console.ship-platform.com"
	wsURL := fmt.Sprintf("%s/ws/portforward/%s?port=%d&token=%s", wsBase, appID, targetPort, client.Token)

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	errCh := make(chan error, 2)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := localConn.Read(buf)
			if n > 0 {
				if wErr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); wErr != nil {
					errCh <- wErr
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					errCh <- err
				} else {
					errCh <- nil
				}
				return
			}
		}
	}()

	go func() {
		for {
			mt, data, err := ws.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if mt == websocket.BinaryMessage {
				if _, wErr := localConn.Write(data); wErr != nil {
					errCh <- wErr
					return
				}
			}
		}
	}()

	<-errCh
}
