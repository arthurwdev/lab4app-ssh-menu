# SSH Menu (sshm)

<img src="assets/sshm.png" alt="SSH Menu logo" width="96" height="96">

[![Build](https://github.com/arthurwdev/lab4app-ssh-menu/actions/workflows/build.yml/badge.svg)](https://github.com/arthurwdev/lab4app-ssh-menu/actions/workflows/build.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

An interactive terminal menu for connecting to hosts defined in your `~/.ssh/config`. Pick a host with the arrow keys, press Enter, and `sshm` runs `ssh <alias>` for you.

## Features

- Reads `~/.ssh/config`, following `Include` directives.
- Header shows the app name, version, and the current user/hostname.
- Table columns: alias, description, user, hostname, key usage, VPN.
- Description and VPN info are read from comments inside each `Host` block.
- Arrow keys to navigate, Enter to connect, Esc to quit without connecting.

## Comment conventions

Add these comments inside a `Host` block to enrich its row in the menu:

```
Host myserver
    # Description Short description of this host
    # VPN: true
    # VPN Name: Company VPN
    HostName 10.0.0.1
    User admin
    IdentityFile ~/.ssh/keys/myserver.pem
```

- `# Description ...` — shown in the DESCRIPTION column. The colon is optional.
- `# VPN: true` — marks the host as requiring a VPN.
- `# VPN Name: ...` — free-text VPN name; also implies `VPN: true` on its own.
- The KEY column is filled in automatically whenever `IdentityFile` is set, whether directly on the host or inherited from a wildcard block such as `Host *`.

## Download

Prebuilt binaries for Windows and Linux (amd64) are published on the [Releases](https://github.com/arthurwdev/lab4app-ssh-menu/releases) page for every `v*` tag, alongside a `checksums.txt`.

## Install / Build

Requires Go 1.25+.

```
go build -o sshm ./cmd/sshm
```

The Windows build embeds `assets/sshm.ico` as the executable's icon automatically — see [CLAUDE.md](CLAUDE.md#windows-icon-pipeline) if you need to regenerate it after changing the logo.

## Usage

```
sshm
sshm --config /path/to/config
```

- `↑`/`↓` (or `k`/`j`): move the selection
- `Enter`: connect — runs `ssh <alias>`
- `Esc` / `Ctrl+C`: quit without connecting

## Parsing notes

- Wildcard `Host` patterns (e.g. `Host *`, `Host bastion-*`) are never shown as selectable rows. They only supply inherited values (`HostName`, `User`, `IdentityFile`, ...) to the literal hosts that match them.
- Value precedence follows OpenSSH's own rule: the first value found for each setting wins, so host-specific blocks should come before generic `Host *` blocks in the config file.
- `Match` blocks are not supported.

## Project layout

```
cmd/sshm/             entry point
internal/sshconfig/   SSH config parser
internal/ui/          terminal UI (bubbletea)
internal/sysinfo/     hostname/user lookup for the header
tools/genicon/        generates assets/sshm.ico from assets/sshm.png
```

## Testing

```
go test ./...
```

## License

Licensed under the [GNU General Public License v3.0](LICENSE).

You are free to use, share, and modify `sshm`, for personal or commercial purposes, at no cost. If you distribute it or a modified version, you must keep it under the same license (GPLv3) and provide the source code — it cannot be turned into closed-source or paid software. Forks should keep clear attribution to the original project and avoid presenting themselves as the official `sshm`.
