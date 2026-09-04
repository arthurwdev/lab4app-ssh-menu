// Command sshm reads ~/.ssh/config and shows an interactive menu of
// configured hosts; picking one runs "ssh <alias>".
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/arthurwdev/lab4app-ssh-menu/internal/sshconfig"
	"github.com/arthurwdev/lab4app-ssh-menu/internal/sysinfo"
	"github.com/arthurwdev/lab4app-ssh-menu/internal/ui"
)

// version is set at build time via:
//
//	go build -ldflags "-X main.version=1.2.3"
var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	configPath := flag.String("config", "", "path to the SSH config file (defaults to ~/.ssh/config)")
	flag.Parse()

	path := *configPath
	if path == "" {
		defaultPath, err := sshconfig.DefaultPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, "sshm:", err)
			return 1
		}
		path = defaultPath
	}

	hosts, err := sshconfig.Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sshm:", err)
		return 1
	}

	model := ui.New(hosts, version, sysinfo.Hostname(), sysinfo.Username())

	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sshm:", err)
		return 1
	}

	alias := finalModel.(ui.Model).Selected()
	if alias == "" {
		return 0
	}

	return connect(alias)
}

func connect(alias string) int {
	cmd := exec.Command("ssh", alias)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "sshm:", err)
		return 1
	}
	return 0
}
