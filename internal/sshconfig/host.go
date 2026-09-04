// Package sshconfig reads and interprets an OpenSSH client config file
// (~/.ssh/config), turning each reachable Host entry into a Host value that
// the UI can render.
package sshconfig

// Host is a single connectable entry from the SSH config, with values
// already resolved according to OpenSSH's "first match wins" precedence
// across Host blocks (including wildcard blocks such as "Host *").
type Host struct {
	Alias       string // the literal alias used to run "ssh <Alias>"
	Description string // from a "# Description ..." comment inside the block
	HostName    string // resolved HostName, or Alias itself if not set
	User        string // resolved User, empty if not set
	UseKey      bool   // true if an IdentityFile is effectively set
	VPN         bool   // true if a "# VPN: true" or "# VPN Name: ..." comment is present
	VPNName     string // from a "# VPN Name: ..." comment, free text
}
