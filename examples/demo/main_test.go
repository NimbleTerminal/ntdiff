package main

import (
	tea "charm.land/bubbletea/v2"
	"strings"
	"testing"
)

func TestSampleAndViewSwitch(t *testing.T) {
	m, err := newViewer(false)
	if err != nil {
		t.Fatal(err)
	}
	defer m.pic.Close()
	if len(m.tree.Files) != 4 {
		t.Fatal("sample files")
	}
	count := 0
	for _, f := range m.tree.Files {
		for _, h := range f.Hunks {
			for _, l := range h.Lines {
				count += len(l.Tokens)
			}
		}
	}
	if count == 0 {
		t.Fatal("no syntax annotations")
	}
	tree := m.tree
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 35})
	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	if !m.typeset || m.tree != tree {
		t.Fatal("view switch reparsed or failed")
	}
	if !strings.Contains(m.View().Content, "TYPESET") {
		t.Fatal("missing typeset view")
	}
	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if m.numbers {
		t.Fatal("numbers toggle")
	}
	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	if m.typeset || !strings.Contains(m.View().Content, "TERMINAL") {
		t.Fatal("terminal return")
	}
}

func TestTypesetSharesFileSelectionAndFocus(t *testing.T) {
	m, err := newViewer(false)
	if err != nil {
		t.Fatal(err)
	}
	defer m.pic.Close()
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 35})
	press := func(s string) { t.Helper(); m.Update(tea.KeyPressMsg{Code: []rune(s)[0], Text: s}) }
	press("j")
	if m.terminal.SelectedFileIndex() != 1 {
		t.Fatal("terminal selection")
	}
	press("v")
	if m.renderedFile != 1 || m.layout == nil {
		t.Fatal("typeset did not render selected file")
	}
	for _, name := range []string{"Changed Files", "server.go", "config.json", "notes.txt"} {
		if !strings.Contains(m.View().Content, name) {
			t.Errorf("typeset missing sidebar text %q", name)
		}
	}
	press("j")
	if m.terminal.SelectedFileIndex() != 2 || m.renderedFile != 2 {
		t.Fatal("typeset file navigation")
	}
	// Right-pane scrolling must not move the file selection.
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	press("k")
	if m.terminal.SelectedFileIndex() != 2 {
		t.Fatal("scroll changed selected file")
	}
	press("v")
	if m.terminal.SelectedFileIndex() != 2 || !strings.Contains(m.View().Content, "A greeting server") {
		t.Fatal("selection was lost switching back")
	}
	press("v")
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	press("j")
	if m.renderedFile != 3 || m.top != 0 {
		t.Fatal("metadata-only file navigation")
	}
	if !m.keys()[3].Enabled() {
		t.Fatal("pane switch unavailable in typeset mode")
	}
}
