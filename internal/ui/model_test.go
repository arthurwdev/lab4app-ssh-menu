package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/arthurwdev/lab4app-ssh-menu/internal/sshconfig"
)

func sampleHosts() []sshconfig.Host {
	return []sshconfig.Host{
		{Alias: "vuecorevm", Description: "Docker core server", HostName: "10.10.2.4", User: "vueadmin", UseKey: true, VPN: true, VPNName: "OpenVPN Vuepay"},
		{Alias: "no-comments", HostName: "10.10.2.9", User: "defaultuser"},
	}
}

func TestModel_RendersBeforeFirstResize(t *testing.T) {
	m := New(sampleHosts(), "0.1.0", "myhost", "arthur")
	view := m.View()

	for _, want := range []string{"SSH Menu v0.1.0", "arthur@myhost", "vuecorevm", "no-comments", "ALIAS"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() missing %q before any WindowSizeMsg; got:\n%s", want, view)
		}
	}
}

func TestModel_ResizeThenNavigateAndSelect(t *testing.T) {
	m := New(sampleHosts(), "0.1.0", "myhost", "arthur")

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	m = updated.(Model)

	view := m.View()
	if !strings.Contains(view, "vuecorevm") || !strings.Contains(view, "OpenVPN Vuepay") {
		t.Fatalf("View() after resize missing expected content:\n%s", view)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 after down", m.cursor)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("expected tea.Quit command after enter")
	}
	if got := m.Selected(); got != "no-comments" {
		t.Fatalf("Selected() = %q, want no-comments", got)
	}
}

func TestModel_EscCancelsWithoutSelection(t *testing.T) {
	m := New(sampleHosts(), "0.1.0", "myhost", "arthur")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("expected tea.Quit command after esc")
	}
	if got := m.Selected(); got != "" {
		t.Fatalf("Selected() = %q, want empty after esc", got)
	}
}

func TestModel_CursorWrapsAtListBounds(t *testing.T) {
	m := New(sampleHosts(), "0.1.0", "myhost", "arthur")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	m = updated.(Model)

	last := len(sampleHosts()) - 1

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.cursor != last {
		t.Fatalf("cursor = %d, want %d (up from first item wraps to last)", m.cursor, last)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 (down from last item wraps to first)", m.cursor)
	}
}

func TestModel_EmptyHostList(t *testing.T) {
	m := New(nil, "0.1.0", "myhost", "arthur")
	view := m.View()
	if !strings.Contains(view, "No hosts found") {
		t.Fatalf("expected empty-state message, got:\n%s", view)
	}
}
