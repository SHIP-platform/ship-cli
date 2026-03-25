package ui

import (
	"context"
	"fmt"
	"strings"

	"ship-cli/api"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logsTuiMaxChunks = 5000

var logsTuiHelpStyle = lipgloss.NewStyle().Foreground(colorGray)

// RunLogsTUI streams application logs in a scrollable full-screen TUI (bubbletea + viewport).
func RunLogsTUI(client *api.Client, appID string, follow bool, tail int) error {
	m := newLogsTUIModel(client, appID, follow, tail)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

type logsTUIModel struct {
	client   *api.Client
	appID    string
	follow   bool
	tail     int
	viewport viewport.Model
	chunks   []string
	logChan  chan string
	cancel   context.CancelFunc
	ready    bool
}

func newLogsTUIModel(client *api.Client, appID string, follow bool, tail int) *logsTUIModel {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true
	return &logsTUIModel{
		client:   client,
		appID:    appID,
		follow:   follow,
		tail:     tail,
		viewport: vp,
	}
}

func (m *logsTUIModel) Init() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.logChan = make(chan string)
	startLogsStream(ctx, m.client, m.appID, m.follow, m.tail, m.logChan)
	return waitForLogLine(m.logChan)
}

func (m *logsTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		frameX, frameY := docStyle.GetFrameSize()
		headerLines := 3
		m.viewport.Width = msg.Width - frameX - 2
		if m.viewport.Width < 20 {
			m.viewport.Width = 20
		}
		m.viewport.Height = msg.Height - frameY - headerLines
		if m.viewport.Height < 5 {
			m.viewport.Height = 5
		}
		m.ready = true
		m.viewport.SetContent(strings.Join(m.chunks, ""))
		m.viewport.GotoBottom()
		return m, nil
	case newLogLineMsg:
		if msg.line == "" {
			return m, nil
		}
		m.chunks = append(m.chunks, msg.line)
		if len(m.chunks) > logsTuiMaxChunks {
			m.chunks = m.chunks[len(m.chunks)-logsTuiMaxChunks:]
		}
		if m.ready {
			m.viewport.SetContent(strings.Join(m.chunks, ""))
			m.viewport.GotoBottom()
		}
		return m, waitForLogLine(m.logChan)
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *logsTUIModel) View() string {
	title := titleStyle.Render(fmt.Sprintf("Logs · %s", m.appID))
	if !m.follow {
		title += " " + lipgloss.NewStyle().Foreground(colorGray).Render("(snapshot)")
	}
	body := m.viewport.View()
	help := logsTuiHelpStyle.Render("↑/↓/PgUp/PgDn scroll · mouse wheel · q/esc quit")
	return docStyle.Render(title + "\n\n" + body + "\n" + help)
}
