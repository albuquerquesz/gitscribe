package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/models"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	configuredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	configuredIndicator = configuredStyle.Render("✓ Configured")

	providerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3C3C3C")).
			Background(lipgloss.Color("#D3D3D3")).
			Padding(0, 1).
			MarginTop(1)

	defaultIndicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true).
				Render(" ⭐ Default")

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")).
			MarginTop(1)

	configuredNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Bold(true)

	unconfiguredNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#C0C0C0"))

	configuredDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#808080"))

	unconfiguredDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#606060"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

type ModelItem struct {
	Model      models.ModelInfo
	Configured bool
	IsDefault  bool
}

func (i ModelItem) Title() string       { return i.Model.Name }
func (i ModelItem) Description() string { return i.Model.Description }
func (i ModelItem) FilterValue() string {
	return i.Model.Name + " " + i.Model.Provider + " " + i.Model.Description
}

type KeyMap struct {
	Select    key.Binding
	Configure key.Binding
	Remove    key.Binding
	Quit      key.Binding
	Help      key.Binding
}

var DefaultKeyMap = KeyMap{
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Configure: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "configure"),
	),
	Remove: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "remove"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

type Model struct {
	list           list.Model
	keys           KeyMap
	configured     map[string]bool
	currentDefault string
	selectedItem   *ModelItem
	err            error
	showHelp       bool
	width          int
	height         int
	quitting       bool
}

func NewModel(cfg *config.Config, keyMgr *secrets.AgentKeyManager) Model {

	configured := make(map[string]bool)
	for _, agent := range cfg.Agents {
		if agent.Enabled {
			configured[string(agent.Provider)] = true
		}
	}

	var items []list.Item
	providerKeys := models.GetProviderKeys()

	for _, providerKey := range providerKeys {
		modelList := models.GetModelsForProvider(providerKey)

		for _, m := range modelList {
			isDefault := cfg.Global.DefaultAgent == models.GenerateProfileName(providerKey, m.ID)

			item := ModelItem{
				Model:      m,
				Configured: configured[providerKey],
				IsDefault:  isDefault,
			}
			items = append(items, item)
		}
	}

	l := list.New(items, modelDelegate{}, 80, 20)
	l.Title = "AI Models"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			DefaultKeyMap.Configure,
			DefaultKeyMap.Remove,
		}
	}

	return Model{
		list:           l,
		keys:           DefaultKeyMap,
		configured:     configured,
		currentDefault: cfg.Global.DefaultAgent,
		showHelp:       true,
	}
}

type modelDelegate struct{}

func (d modelDelegate) Height() int  { return 2 }
func (d modelDelegate) Spacing() int { return 0 }

func (d modelDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d modelDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(ModelItem)
	if !ok {
		return
	}

	var lines []string

	cursor := "  "
	if index == m.Index() {
		cursor = selectedStyle.Render("❯ ")
	}

	nameStr := unconfiguredNameStyle.Render(i.Model.Name)
	if i.Configured {
		nameStr = configuredNameStyle.Render(i.Model.Name)
	}

	firstLine := cursor + nameStr

	if i.Configured {
		firstLine += "  " + configuredIndicator
	}

	if i.IsDefault {
		firstLine += defaultIndicator
	}

	lines = append(lines, firstLine)

	descStr := unconfiguredDescStyle.Render("    " + i.Model.Description)
	if i.Configured {
		descStr = configuredDescStyle.Render("    " + i.Model.Description)
	}
	lines = append(lines, descStr)

	fmt.Fprint(w, strings.Join(lines, "\n"))
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 6)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Select):
			if item, ok := m.list.SelectedItem().(ModelItem); ok {
				m.selectedItem = &item
				m.quitting = true
				return m, tea.Quit
			}

		case key.Matches(msg, m.keys.Configure):

			return m, nil

		case key.Matches(msg, m.keys.Remove):

			return m, nil

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("🤖 AI Model Selector"))
	s.WriteString("\n\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Configured Providers:"))
	s.WriteString("\n")

	hasConfigured := false
	for provider, isConfigured := range m.configured {
		if isConfigured {
			p := models.Providers[provider]
			s.WriteString(configuredStyle.Render(fmt.Sprintf("✓ %s %s", p.Icon, p.DisplayName)))
			s.WriteString("\n")
			hasConfigured = true
		}
	}

	if !hasConfigured {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render("  None configured yet"))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	s.WriteString(m.list.View())

	if m.showHelp {
		s.WriteString("\n")
		s.WriteString(helpStyle.Render(
			"enter: select  •  c: configure  •  r: remove  •  /: filter  •  ?: help  •  q: quit",
		))
	}

	if m.err != nil {
		s.WriteString("\n")
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return s.String()
}

func (m Model) GetSelected() *ModelItem {
	return m.selectedItem
}

func (m *Model) SetError(err error) {
	m.err = err
}
