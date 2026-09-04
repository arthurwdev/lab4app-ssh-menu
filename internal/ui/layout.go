package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/arthurwdev/lab4app-ssh-menu/internal/sshconfig"
)

const (
	minDescriptionWidth = 15
	minVPNWidth         = 10
	maxAliasWidth       = 24
	maxDescriptionWidth = 40
	maxUserWidth        = 16
	maxHostnameWidth    = 22
	maxVPNWidth         = 34

	separator  = " │ "
	rowPrefix  = "  "
	curPrefix  = "▶ "
	keyGlyphOn = "●"
	dotGlyph   = "·"
)

type columns struct {
	alias, description, user, hostname, key, vpn int
}

// computeColumns sizes every column from the actual host data (capped at
// sensible maximums), then shrinks the most flexible columns so the table
// fits the terminal width.
func computeColumns(hosts []sshconfig.Host, termWidth int) columns {
	c := columns{
		alias:       len("ALIAS"),
		description: len("DESCRIPTION"),
		user:        len("USER"),
		hostname:    len("HOSTNAME"),
		key:         len("KEY"),
		vpn:         len("VPN"),
	}
	for _, h := range hosts {
		c.alias = max(c.alias, len(h.Alias))
		c.description = max(c.description, len(h.Description))
		c.user = max(c.user, len(h.User))
		c.hostname = max(c.hostname, len(h.HostName))
		c.vpn = max(c.vpn, vpnNaturalWidth(h))
	}
	c.alias = min(c.alias, maxAliasWidth)
	c.description = min(c.description, maxDescriptionWidth)
	c.user = min(c.user, maxUserWidth)
	c.hostname = min(c.hostname, maxHostnameWidth)
	c.vpn = min(c.vpn, maxVPNWidth)

	shrinkToFit(&c, termWidth)
	return c
}

func vpnNaturalWidth(h sshconfig.Host) int {
	if !h.VPN {
		return len(dotGlyph)
	}
	if h.VPNName == "" {
		return len(keyGlyphOn)
	}
	return len(keyGlyphOn) + 1 + len(h.VPNName)
}

// total returns the full rendered width of the table body (excluding the
// leading selection prefix), including inter-column separators.
func (c columns) total() int {
	fixed := c.alias + c.description + c.user + c.hostname + c.key + c.vpn
	return fixed + lipgloss.Width(separator)*5
}

func shrinkToFit(c *columns, termWidth int) {
	if termWidth <= 0 {
		return
	}
	budget := termWidth - len(rowPrefix)
	for c.total() > budget && (c.description > minDescriptionWidth || c.vpn > minVPNWidth) {
		switch {
		case c.description > minDescriptionWidth:
			c.description--
		case c.vpn > minVPNWidth:
			c.vpn--
		}
	}
}

func padVisible(s string, width int) string {
	if w := lipgloss.Width(s); w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

// truncateVisible shortens s to width visible cells, appending an ellipsis
// when it had to cut content. It assumes s carries no ANSI styling yet.
func truncateVisible(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return string(runes[:1])
	}
	return string(runes[:width-1]) + "…"
}

func renderHeaderRow(c columns, styles Styles) string {
	cells := []string{
		padVisible("ALIAS", c.alias),
		padVisible("DESCRIPTION", c.description),
		padVisible("USER", c.user),
		padVisible("HOSTNAME", c.hostname),
		padVisible("KEY", c.key),
		padVisible("VPN", c.vpn),
	}
	return rowPrefix + styles.ColumnHeader.Render(strings.Join(cells, separator))
}

func keyGlyph(h sshconfig.Host) string {
	if h.UseKey {
		return keyGlyphOn
	}
	return dotGlyph
}

func vpnGlyphText(h sshconfig.Host, width int) string {
	if !h.VPN {
		return dotGlyph
	}
	if h.VPNName == "" {
		return keyGlyphOn
	}
	return keyGlyphOn + " " + truncateVisible(h.VPNName, width-2)
}

func renderKeyCell(h sshconfig.Host, width int, styles Styles) string {
	if h.UseKey {
		return padVisible(styles.Good.Render(keyGlyphOn), width)
	}
	return padVisible(styles.Muted.Render(dotGlyph), width)
}

func renderVPNCell(h sshconfig.Host, width int, styles Styles) string {
	if !h.VPN {
		return padVisible(styles.Muted.Render(dotGlyph), width)
	}
	dot := styles.Good.Render(keyGlyphOn)
	if h.VPNName == "" {
		return padVisible(dot, width)
	}
	name := truncateVisible(h.VPNName, width-2)
	return padVisible(dot+" "+name, width)
}

// renderRow renders one host as a table line. Selected rows are rendered
// from plain (unstyled) cell text wrapped in a single Reverse style, since
// nesting an outer style around already-styled inner text would reset the
// outer style's attributes midway through the line.
func renderRow(h sshconfig.Host, c columns, selected bool, styles Styles) string {
	alias := padVisible(truncateVisible(h.Alias, c.alias), c.alias)
	desc := padVisible(truncateVisible(h.Description, c.description), c.description)
	user := padVisible(truncateVisible(h.User, c.user), c.user)
	hostname := padVisible(truncateVisible(h.HostName, c.hostname), c.hostname)

	if selected {
		key := padVisible(keyGlyph(h), c.key)
		vpn := padVisible(vpnGlyphText(h, c.vpn), c.vpn)
		body := strings.Join([]string{alias, desc, user, hostname, key, vpn}, separator)
		return styles.Accent.Render(curPrefix) + styles.Selected.Render(body)
	}

	key := renderKeyCell(h, c.key, styles)
	vpn := renderVPNCell(h, c.vpn, styles)
	sep := styles.Separator.Render(separator)
	body := strings.Join([]string{alias, desc, user, hostname, key, vpn}, sep)
	return rowPrefix + body
}

func renderHeader(version, username, hostname string, width int, styles Styles) string {
	left := styles.Title.Render("SSH Menu v" + version)
	right := styles.Meta.Render(username + "@" + hostname)

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	line := " " + left + strings.Repeat(" ", gap) + right

	divider := styles.Divider.Render(strings.Repeat("─", max(width-1, 1)))
	return line + "\n " + divider
}

func renderFooter(styles Styles) string {
	return styles.Muted.Render(" ↑/↓ navigate    enter connect    esc quit")
}
