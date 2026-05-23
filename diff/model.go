package diff

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type FileChange struct {
	Path      string
	Content   string
	Diff      string
	Additions int
	Deletions int
}

type FocusPane int

const (
	FocusFiles FocusPane = iota
	FocusDiff
)

type Model struct {
	files       list.Model
	viewport    viewport.Model
	filesList   []FileChange
	selected    int
	width       int
	height      int
	showHelp    bool
	activeFocus FocusPane
}

func NewModel() Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Changed Files"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	vp := viewport.New()
	vp.SetContent("No diff loaded")

	return Model{
		files:       l,
		viewport:    vp,
		activeFocus: FocusFiles,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		leftOuterWidth := m.width / 3
		rightOuterWidth := m.width - leftOuterWidth
		leftInnerWidth := leftOuterWidth - 2
		rightInnerWidth := rightOuterWidth - 2
		innerHeight := m.height - 3

		m.files.SetSize(leftInnerWidth, innerHeight)
		m.viewport.SetWidth(rightInnerWidth)
		m.viewport.SetHeight(innerHeight)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.activeFocus == FocusFiles {
				m.activeFocus = FocusDiff
			} else {
				m.activeFocus = FocusFiles
			}
			return m, nil
		case "h", "left":
			m.activeFocus = FocusFiles
			return m, nil
		case "l", "right":
			m.activeFocus = FocusDiff
			return m, nil
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}
	}

	// Dispatch keys only to the focused panel to avoid navigation conflicts
	_, isKey := msg.(tea.KeyMsg)
	if isKey {
		if m.activeFocus == FocusFiles {
			prevIndex := m.files.Index()
			m.files, cmd = m.files.Update(msg)
			cmds = append(cmds, cmd)

			// Automatically update diff preview on navigation
			if m.files.Index() != prevIndex && len(m.filesList) > 0 {
				m.selected = m.files.Index()
				m.viewport.SetContent(ColorizeDiff(m.filesList[m.selected].Diff))
				m.viewport.SetYOffset(0)
			}
		} else {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	} else {
		// Non-key messages (e.g. WindowSizeMsg, MouseMsg) are sent to both
		m.files, cmd = m.files.Update(msg)
		cmds = append(cmds, cmd)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.NewView("Initializing...")
	}

	var content string
	if len(m.filesList) == 0 {
		content = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("No diff loaded.")
	} else {
		leftOuterWidth := m.width / 3
		rightOuterWidth := m.width - leftOuterWidth
		leftInnerWidth := leftOuterWidth - 2
		rightInnerWidth := rightOuterWidth - 2
		innerHeight := m.height - 3
		leftBorderColor := ColorInactiveBorder
		rightBorderColor := ColorInactiveBorder
		if m.activeFocus == FocusFiles {
			leftBorderColor = ColorActiveBorder
		} else {
			rightBorderColor = ColorActiveBorder
		}

		left := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(leftBorderColor).
			Width(leftInnerWidth).
			Height(innerHeight).
			Render(m.files.View())

		right := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(rightBorderColor).
			Width(rightInnerWidth).
			Height(innerHeight).
			Render(m.viewport.View())

		content = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}

	// Status / Help bar
	var helpText string
	if m.showHelp {
		helpText = " [q] Quit  [Tab] Switch Pane  [↑/↓] Navigate  [?] Hide Help "
	} else {
		helpText = " Active: "
		if m.activeFocus == FocusFiles {
			helpText += "Files Pane"
		} else {
			helpText += "Diff Pane"
		}
		helpText += " • Press [?] for help"
	}

	statusBar := lipgloss.NewStyle().
		Foreground(ColorStatusText).
		Background(ColorStatusBg).
		Width(m.width).
		Render(helpText)

	finalContent := lipgloss.JoinVertical(lipgloss.Left, content, statusBar)

	v := tea.NewView(finalContent)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// LoadDiff loads unified diff text and parses it into files
func (m *Model) LoadDiff(diffText string) {
	files := ParseDiff(diffText)
	m.filesList = files

	items := make([]list.Item, len(files))
	for i, f := range files {
		items[i] = listItem{
			title: fmt.Sprintf("%s (+%d, -%d)", f.Path, f.Additions, f.Deletions),
		}
	}

	m.files.SetItems(items)
	if len(files) > 0 {
		m.viewport.SetContent(ColorizeDiff(files[0].Diff))
	}
}

type listItem struct {
	title string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return i.title }
