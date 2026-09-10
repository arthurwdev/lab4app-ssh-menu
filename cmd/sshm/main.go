// Command sshm reads ~/.ssh/config and shows an interactive menu of
// configured hosts; picking one runs "ssh <alias>".
//
// Copyright (C) 2025 Arthur Souza
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
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
