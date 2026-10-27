// Command demo shows one parsed diff through the terminal and document renderers.
package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"os"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	booba "github.com/NimbleMarkets/go-booba"
	"github.com/NimbleMarkets/go-booba/actions"
	"github.com/NimbleMarkets/go-booba/actions/bubblehelp"
	"github.com/NimbleMarkets/ntdiff/diff"
	diffTypeset "github.com/NimbleMarkets/ntdiff/typeset"
	"github.com/charmbracelet/x/ansi"
	"nimbleterminal.dev/ntfonts"
	nt "nimbleterminal.dev/ntpicture"
	"nimbleterminal.dev/nttypeset"
)

//go:embed sample.diff
var sample string

// This deliberately small demo lexer annotates ranges rather than replacing
// source text. Applications can attach tokens from their own complete lexer.
var syntax = regexp.MustCompile(`"[^"\n]*"|\b(package|import|func|return|select|case|true)\b|\b[0-9]+\b`)

func sampleTree() (*diff.Diff, error) {
	tree, err := diff.Parse(sample)
	if err != nil {
		return nil, err
	}
	for fi := range tree.Files {
		for hi := range tree.Files[fi].Hunks {
			for li := range tree.Files[fi].Hunks[hi].Lines {
				line := &tree.Files[fi].Hunks[hi].Lines[li]
				for _, r := range syntax.FindAllStringIndex(line.Text, -1) {
					kind := "keyword"
					ink := color.RGBA{R: 197, G: 145, B: 255, A: 255}
					if line.Text[r[0]] == '"' {
						kind = "string"
						ink = color.RGBA{R: 130, G: 205, B: 237, A: 255}
					} else if line.Text[r[0]] >= '0' && line.Text[r[0]] <= '9' {
						kind = "number"
						ink = color.RGBA{R: 240, G: 190, B: 105, A: 255}
					}
					line.Tokens = append(line.Tokens, diff.Token{Start: r[0], End: r[1], Kind: kind, Style: diff.TextStyle{Foreground: ink}})
				}
			}
		}
	}
	return tree, nil
}

type viewer struct {
	tree               *diff.Diff
	terminal           diff.Model
	pic                *nt.Picture
	engine             *nttypeset.Engine
	layout             *nttypeset.Layout
	geom               nt.RasterGeometry
	width, height, top int
	renderedFile       int
	typeset, numbers   bool
	err                error
}

func newViewer(typeset bool) (*viewer, error) {
	tree, err := sampleTree()
	if err != nil {
		return nil, err
	}
	eng, err := nttypeset.NewEngine(nttypeset.WithFonts(ntfonts.Embedded), nttypeset.WithLook("technical", ""))
	if err != nil {
		return nil, err
	}
	terminal := diff.NewModel()
	terminal.LoadTree(tree)
	return &viewer{tree: tree, terminal: terminal, engine: eng, pic: nt.New(nt.Options{AutoMode: true}), typeset: typeset, numbers: true, renderedFile: -1}, nil
}

type keys []key.Binding

func (k keys) ShortHelp() []key.Binding  { return k }
func (k keys) FullHelp() [][]key.Binding { return [][]key.Binding{k} }
func binding(label, k string) key.Binding {
	return key.NewBinding(key.WithKeys(k), key.WithHelp(k, label))
}
func (m *viewer) keys() keys {
	k := keys{binding("Switch view", "v"), binding("Down", "j"), binding("Up", "k"), binding("Switch pane", "tab"), binding("Line numbers", "n"), binding("Glyph / Kitty", "t"), binding("Quit", "q")}
	k[4].SetEnabled(m.typeset)
	k[5].SetEnabled(m.typeset && m.pic.KittySupported() == nt.KittyCapabilitySupported)
	return k
}
func (m *viewer) Init() tea.Cmd { return tea.Batch(m.pic.Init(), m.pic.SetVisible(m.typeset)) }
func (m *viewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = size.Width, size.Height
		updated, cmd := m.terminal.Update(tea.WindowSizeMsg{Width: size.Width, Height: max(1, size.Height-3)})
		m.terminal = updated.(diff.Model)
		pc := m.pic.Update(msg)
		pw, ph := m.terminal.DiffPaneSize()
		sc := m.pic.SetSize(pw, ph)
		m.layout = nil
		return m, tea.Batch(cmd, pc, sc, m.redraw())
	}
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "q", "ctrl+c":
			return m, tea.Sequence(m.pic.Close(), tea.Quit)
		case "v":
			m.typeset = !m.typeset
			return m, tea.Batch(m.pic.SetVisible(m.typeset), m.redraw())
		}
		if m.typeset {
			switch k.String() {
			case "n":
				m.numbers = !m.numbers
				m.layout = nil
				return m, m.redraw()
			case "t":
				return m, tea.Batch(m.pic.ToggleMode(), m.redraw())
			}
			if m.terminal.FocusedPane() == diff.FocusDiff {
				handled := true
				switch k.String() {
				case "j", "down":
					m.top += 60
				case "k", "up":
					m.top = max(0, m.top-60)
				case "space", "pgdown":
					m.top += 400
				case "pgup":
					m.top = max(0, m.top-400)
				case "g", "home":
					m.top = 0
				default:
					handled = false
				}
				if handled {
					return m, m.redraw()
				}
			}
		}
		updated, cmd := m.terminal.Update(msg)
		m.terminal = updated.(diff.Model)
		return m, tea.Batch(cmd, m.redraw())
	}
	pc := m.pic.Update(msg)
	updated, cmd := m.terminal.Update(msg)
	m.terminal = updated.(diff.Model)
	if m.pic.RasterGeometry() != m.geom {
		m.layout = nil
		return m, tea.Batch(pc, cmd, m.redraw())
	}
	return m, tea.Batch(pc, cmd)
}
func (m *viewer) redraw() tea.Cmd {
	if !m.typeset {
		return nil
	}
	g := m.pic.RasterGeometry()
	if g.Pixels.X <= 0 || g.Pixels.Y <= 0 {
		return nil
	}
	m.geom = g
	// Stable text size in terminal-cell units across glyph and Kitty renderers.
	paneWidth, _ := m.terminal.DiffPaneSize()
	width := max(160, paneWidth*10)
	scale := float64(g.Pixels.X) / float64(width)
	height := max(1, int(float64(g.Pixels.Y)/scale))
	selected := m.terminal.SelectedFileIndex()
	if selected < 0 {
		return nil
	}
	if selected != m.renderedFile {
		m.layout = nil
		m.top = 0
		m.renderedFile = selected
	}
	if m.layout == nil {
		m.layout, m.err = m.engine.Layout(diffTypeset.Convert(&diff.Diff{Files: []diff.File{m.tree.Files[selected]}}, diffTypeset.Options{LineNumbers: m.numbers, Size: 18}), width, nttypeset.LayoutOptions{})
		if m.err != nil {
			return nil
		}
		m.err = m.layout.Complete(context.Background())
		if m.err != nil {
			return nil
		}
	}
	m.top = min(m.top, max(0, m.layout.Extent()-height))
	img, err := m.layout.Raster(context.Background(), nttypeset.View{Top: m.top, Width: width, Height: height, Scale: scale, Pixels: g.Pixels}, nttypeset.RasterOptions{Palette: nttypeset.DarkPalette()})
	m.err = err
	if err != nil {
		return nil
	}
	return m.pic.SetImageFor(img, g)
}
func (m *viewer) View() tea.View {
	name := "TERMINAL"
	content := m.terminal.View().Content
	if m.typeset {
		name = "TYPESET"
		content = m.terminal.ViewWithDiffContent(m.pic.View()).Content
	}
	status := fmt.Sprintf(" %s · 4 files · v switch view · Tab pane · j/k navigate · q Quit", name)
	if m.err != nil {
		status = "Error: " + m.err.Error()
	}
	lines := strings.Split(content, "\n")
	if len(lines) > max(1, m.height-1) {
		lines = lines[:max(1, m.height-1)]
	}
	content = strings.Join(lines, "\n")
	v := tea.NewView(content + "\n" + ansi.Truncate(status, max(1, m.width), "…"))
	v.AltScreen = true
	return v
}
func main() {
	mode := flag.Bool("typeset", false, "start in typeset view")
	flag.Parse()
	m, err := newViewer(*mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer m.pic.Close()
	wrapped, err := bubblehelp.NewModel(m, func(inner tea.Model) bubblehelp.Description {
		v := inner.(*viewer)
		mode := "terminal"
		if v.typeset {
			mode = "typeset"
		}
		return bubblehelp.Description{Context: mode, Keys: v.keys(), Status: actions.ProgramStatus{App: "ntdiff", Summary: "One diff, two renderers"}}
	}, bubblehelp.ModelOptions{})
	if err == nil {
		err = booba.Run(wrapped)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
