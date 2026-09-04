# CLAUDE.md

Instructions and context for Claude Code (or any agent) working on this repository.

## Project

`sshm` (SSH Menu) is an interactive terminal menu that reads `~/.ssh/config` and lets the user pick a host to connect to. Enter runs `ssh <alias>`; Esc quits without connecting. See [README.md](README.md) for user-facing docs.

This project is intended to go open source. Keep that in mind for licensing headers, comment tone, and anything that assumes a private/internal context.

## Language rules

- **All `.md` documentation is written in English**, objectively (no filler). This includes README, CLAUDE.md, AGENTS.md, and any future docs.
- **All code, identifiers, and code comments are in English.**
- **The interface (UI strings shown in the terminal) is always in English.** Review terminology before adding new UI strings — the project owner is not highly fluent in English, so prefer plain, common words over idioms.
- Conversation with the user in Claude Code happens in Portuguese (per the user's global config), but that never leaks into files.

## Stack and key decisions

- **Language**: Go (module `github.com/arthurwdev/lab4app-ssh-menu`, binary name `sshm`). Chosen for single static binaries and easy cross-compilation.
- **Target platforms**: Windows, Linux, macOS (multiplatform). The dev machine is Windows; do not assume Unix-only behavior (no `syscall.Exec`, path separators, etc.).
- **`go` directive**: kept at whatever `go mod tidy` resolves as the minimum required by dependencies (currently `1.25.0`), not pinned to the locally installed toolchain version — keep it as low as the dependency graph allows for wider compatibility once open-sourced.
- **TUI libs**: `charmbracelet/bubbletea` (event loop) + `charmbracelet/lipgloss` (styling). Deliberately **not** using `charmbracelet/bubbles/table`: a hand-rolled renderer in `internal/ui/layout.go` gives full control over column widths, truncation, and the colored KEY/VPN indicator glyphs without fighting the component's own styling model.
- **No search/filter in v1** — explicitly deferred to a future version. Don't add it unless asked.
- **`Match` blocks are not supported** by the parser — explicitly out of scope.

## `internal/sshconfig` parsing rules (already implemented, keep consistent)

- `Host` blocks with wildcard patterns (`*`, `?`) are never listed as selectable hosts; they only contribute inherited values.
- Effective values (`HostName`, `User`, `IdentityFile` presence) are resolved by scanning **all** blocks in file order and applying OpenSSH's real precedence: **first value found for a key wins**. This means a `Host *` block placed *before* a specific host in the file will shadow that host's own directive — this is correct, intentional, and mirrors real `ssh` behavior. Test fixtures in `internal/sshconfig/testdata/` deliberately put `Host *` last to reflect best practice.
- `Description`, `VPN`, and `VPN Name` are read **only** from comments inside the literal host's own block — never inherited from wildcard blocks.
- Comment formats are matched case-insensitively with an optional colon: `# Description ...`, `# VPN: true` (or `yes`/`1`), `# VPN Name: ...`. A `VPN Name` comment implies `VPN: true` even without an explicit `# VPN: true` line.
- `Include` is supported (glob + `~` expansion), expanded in place at the point it appears, with cycle protection via a visited-files set.
- `HostName` falls back to the alias itself when not explicitly set (matches real `ssh` behavior).
- Multiple aliases on one `Host` line (`Host a b`) produce one row per alias, sharing the same effective values.

## `internal/ui` rendering notes

- Row rendering avoids nesting one lipgloss `Style.Render()` call inside another: an inner style's trailing reset code would also clear the outer style's attributes mid-line. Selected rows are built from **plain, unstyled** cell text and wrapped in a single `Reverse(true)` style; unstyled columns are never pre-styled before that wrap. See the comment on `renderRow` in `internal/ui/layout.go` before changing this.
- `ui.New()` seeds a `defaultWidth` (80) and computes columns immediately, so the very first frame isn't blank before bubbletea's first real `WindowSizeMsg` arrives.
- Colors use `lipgloss.AdaptiveColor` so the menu stays readable in both light and dark terminal themes.

## Windows icon pipeline

The Windows binary embeds `assets/sshm.ico` as its executable icon. This is a two-step, mostly-manual pipeline; both generated files are committed so a normal `go build` (local or in CI) needs nothing extra:

1. `assets/sshm.ico` is generated from `assets/sshm.png` (the project logo) by `tools/genicon` — a small internal tool (module-local, `package main`) using `golang.org/x/image/draw` to render the 256/48/32/16 px frames and pack them into a standard multi-resolution ICO with PNG-compressed frames (the Vista+ format; validated by manually parsing the `ICONDIR`/`ICONDIRENTRY` structure and decoding each embedded frame, not just trusting `System.Drawing.Icon` which has known quirks with PNG-compressed ICO frames). Regenerate with:
   ```
   go run ./tools/genicon
   ```
   `golang.org/x/image` is a direct dependency of the module for this reason, but only `tools/genicon` imports it — it is never linked into the shipped `sshm`/`sshm.exe` binary.
2. `cmd/sshm/rsrc_windows_amd64.syso` is generated from that `.ico` using [`akavel/rsrc`](https://github.com/akavel/rsrc) (a one-time external dev tool, not a module dependency):
   ```
   go install github.com/akavel/rsrc@latest
   rsrc -ico=assets/sshm.ico -arch=amd64 -o=cmd/sshm/rsrc_windows_amd64.syso
   ```
   Go's toolchain auto-links any `*_windows_amd64.syso` file found next to `package main` when `GOOS=windows GOARCH=amd64` — including when cross-compiling from Linux/macOS, since it's just a prebuilt COFF object with no host-tool dependency. That's why the release workflow needs no extra step to get the icon into `sshm-windows-amd64.exe`.

Regenerate both files whenever `assets/sshm.png` changes. There is no `.syso` for other GOARCH values (only `amd64` is built/released for Windows); add one the same way if that ever changes.

## Dev environment gotcha

On the primary dev machine, Go is installed at `C:\Program Files\Go\bin\go.exe` but **is not on PATH** in fresh shells. Prefix commands with:

```
export PATH="/c/Program Files/Go/bin:$PATH"
```

(Bash tool) — this does not persist between tool calls, so repeat it per command or chain it in the same invocation.

## Testing

- `internal/sshconfig`: table/fixture-driven tests in `parser_test.go` against `testdata/main_config` + `testdata/extra_config` (the latter exercises `Include`). Extend the fixture rather than creating parallel ones when adding parser behavior.
- `internal/ui`: `model_test.go` drives the bubbletea `Model` directly (`Update`/`View`) without a real TTY — no need for manual interactive testing for logic changes. Don't leave ad-hoc print-only preview tests in the repo; use them locally then delete them.
- Run everything with:
  ```
  go build ./... && go vet ./... && go test ./...
  ```

## CI/CD (`.github/workflows/`)

- **`build.yml`** — runs on push to `master` and on PRs targeting `master`. Checks `gofmt`, runs `go vet` and `go test`, then does a cross-compile sanity build (`go build` to `/dev/null`, no execution) for windows/linux/darwin amd64 to back up the multiplatform claim in the README.
- **`release.yml`** — runs on pushing a tag matching `v*`. Cross-compiles `sshm` for windows/amd64 and linux/amd64 only (per explicit scope — no darwin release binaries), strips debug info (`-ldflags "-s -w"`), stamps `main.version` from the tag, generates `checksums.txt`, and publishes everything to a GitHub Release via `gh release create` (relies on the runner's preinstalled `gh` CLI and the automatic `GITHUB_TOKEN`, no third-party release action).
- Both workflows use `actions/setup-go@v5` with `go-version-file: go.mod`, so they always match the module's declared minimum Go version — no version number to keep in sync by hand.

## Out of scope for now (don't add unless the user asks)

- Search/filter in the host list.
- `Match` block support.
- Editing the SSH config from within the app.
- Non-amd64 architectures, and macOS release binaries, in `release.yml`.
- A dedicated `goversioninfo`-style version-info resource (product name, company, file version) in the Windows binary — only the icon is embedded today.
