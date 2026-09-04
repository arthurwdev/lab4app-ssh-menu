// Package sysinfo resolves the local machine name and current user for the
// menu header.
package sysinfo

import (
	"os"
	"os/user"
	"strings"
)

// Hostname returns the local machine's hostname, or "unknown" if it can't
// be determined.
func Hostname() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "unknown"
	}
	return h
}

// Username returns the current OS user's login name, or "unknown" if it
// can't be determined.
func Username() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		// On Windows, user.Current().Username is often "DOMAIN\user";
		// keep just the account name.
		name := u.Username
		if idx := strings.LastIndex(name, `\`); idx != -1 {
			name = name[idx+1:]
		}
		return name
	}
	for _, envVar := range []string{"USER", "USERNAME"} {
		if v := os.Getenv(envVar); v != "" {
			return v
		}
	}
	return "unknown"
}
