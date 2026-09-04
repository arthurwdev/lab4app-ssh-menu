package sshconfig

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	descriptionRe = regexp.MustCompile(`(?i)^#\s*description\s*:?\s*(.+)$`)
	vpnFlagRe     = regexp.MustCompile(`(?i)^#\s*vpn\s*:?\s*(true|yes|1)\s*$`)
	vpnNameRe     = regexp.MustCompile(`(?i)^#\s*vpn[\s_-]*name\s*:?\s*(.+)$`)
)

// directive is a single "Key value" line captured from inside a Host block.
type directive struct {
	key   string // lower-cased
	value string
}

// rawBlock is one "Host <pattern...>" section as it appears in the file,
// before resolving inheritance between blocks.
type rawBlock struct {
	patterns   []string
	directives []directive
	comments   []string
}

// Load reads and parses the SSH config at path (following any Include
// directives) and returns the list of connectable hosts, in the order their
// aliases first appear in the file.
func Load(path string) ([]Host, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("SSH config not found at %s", path)
		}
		return nil, err
	}

	blocks, err := parseFile(path, map[string]bool{})
	if err != nil {
		return nil, err
	}
	return buildHosts(blocks), nil
}

// DefaultPath returns the platform's default SSH config location (~/.ssh/config).
func DefaultPath() (string, error) {
	u, err := user.Current()
	if err != nil || u.HomeDir == "" {
		return "", fmt.Errorf("could not resolve home directory: %w", err)
	}
	return filepath.Join(u.HomeDir, ".ssh", "config"), nil
}

func parseFile(filePath string, visited map[string]bool) ([]*rawBlock, error) {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		abs = filePath
	}
	if visited[abs] {
		return nil, nil
	}
	visited[abs] = true

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var blocks []*rawBlock
	var current *rawBlock

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if current != nil {
				current.comments = append(current.comments, line)
			}
			continue
		}

		key, value := splitDirective(line)
		switch key {
		case "host":
			current = &rawBlock{patterns: strings.Fields(value)}
			blocks = append(blocks, current)
		case "match":
			// Match blocks are not supported; stop attributing directives
			// to a host until the next "Host" line.
			current = nil
		case "include":
			included, err := resolveIncludes(value, filepath.Dir(filePath), visited)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, included...)
		default:
			if current != nil {
				current.directives = append(current.directives, directive{key: key, value: value})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return blocks, nil
}

func resolveIncludes(value, baseDir string, visited map[string]bool) ([]*rawBlock, error) {
	var result []*rawBlock
	for _, pattern := range strings.Fields(value) {
		resolved := expandPath(pattern, baseDir)
		matches, err := filepath.Glob(resolved)
		if err != nil {
			return nil, err
		}
		sort.Strings(matches)
		for _, m := range matches {
			blocks, err := parseFile(m, visited)
			if err != nil {
				// A file referenced by Include that can't be read is skipped
				// rather than failing the whole menu.
				continue
			}
			result = append(result, blocks...)
		}
	}
	return result, nil
}

func expandPath(p, baseDir string) string {
	if strings.HasPrefix(p, "~/") || p == "~" {
		if u, err := user.Current(); err == nil {
			p = filepath.Join(u.HomeDir, strings.TrimPrefix(p, "~"))
		}
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(baseDir, p)
	}
	return p
}

// splitDirective splits a config line into its lower-cased key and value,
// accepting "Key value", "Key=value" and "Key = value" forms.
func splitDirective(line string) (key, value string) {
	idx := strings.IndexFunc(line, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '='
	})
	if idx == -1 {
		return strings.ToLower(line), ""
	}
	key = strings.ToLower(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	value = strings.TrimPrefix(value, "=")
	value = strings.TrimSpace(value)
	return key, value
}

func isWildcard(pattern string) bool {
	return strings.ContainsAny(pattern, "*?")
}

func matchesAny(alias string, patterns []string) bool {
	for _, p := range patterns {
		if ok, err := path.Match(p, alias); err == nil && ok {
			return true
		}
	}
	return false
}

// buildHosts resolves effective values for every literal (non-wildcard)
// alias, applying the first-value-wins precedence OpenSSH uses across all
// matching blocks (including wildcard blocks like "Host *").
func buildHosts(blocks []*rawBlock) []Host {
	var hosts []Host
	seen := map[string]bool{}

	for _, block := range blocks {
		for _, alias := range block.patterns {
			if isWildcard(alias) || seen[alias] {
				continue
			}
			seen[alias] = true

			h := Host{Alias: alias}
			applyComments(&h, block.comments)

			effective := map[string]string{}
			for _, other := range blocks {
				if !matchesAny(alias, other.patterns) {
					continue
				}
				for _, d := range other.directives {
					if _, exists := effective[d.key]; !exists {
						effective[d.key] = d.value
					}
				}
			}

			h.HostName = effective["hostname"]
			if h.HostName == "" {
				h.HostName = alias
			}
			h.User = effective["user"]
			if _, ok := effective["identityfile"]; ok {
				h.UseKey = true
			}

			hosts = append(hosts, h)
		}
	}
	return hosts
}

func applyComments(h *Host, comments []string) {
	for _, c := range comments {
		if m := vpnNameRe.FindStringSubmatch(c); m != nil {
			h.VPNName = strings.TrimSpace(m[1])
			h.VPN = true
			continue
		}
		if m := vpnFlagRe.FindStringSubmatch(c); m != nil {
			h.VPN = true
			continue
		}
		if m := descriptionRe.FindStringSubmatch(c); m != nil {
			h.Description = strings.TrimSpace(m[1])
			continue
		}
	}
}
