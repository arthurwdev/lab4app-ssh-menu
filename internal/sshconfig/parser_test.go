package sshconfig

import (
	"testing"
)

func TestLoad_BuildsExpectedHosts(t *testing.T) {
	hosts, err := Load("testdata/main_config")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	wantAliases := []string{"included-host", "vuecorevm", "no-comments", "web-a", "web-b", "bastion-01"}
	if len(hosts) != len(wantAliases) {
		t.Fatalf("got %d hosts, want %d: %+v", len(hosts), len(wantAliases), hosts)
	}
	for i, want := range wantAliases {
		if hosts[i].Alias != want {
			t.Errorf("hosts[%d].Alias = %q, want %q", i, hosts[i].Alias, want)
		}
	}

	byAlias := map[string]Host{}
	for _, h := range hosts {
		byAlias[h.Alias] = h
	}

	t.Run("included host merges Include and inherits Host * defaults", func(t *testing.T) {
		h := byAlias["included-host"]
		if h.HostName != "172.16.0.5" {
			t.Errorf("HostName = %q, want 172.16.0.5", h.HostName)
		}
		if h.User != "includeduser" {
			t.Errorf("User = %q, want includeduser (own block should win over Host *)", h.User)
		}
		if h.Description != "Comes from an included file" {
			t.Errorf("Description = %q", h.Description)
		}
		if h.UseKey {
			t.Errorf("UseKey = true, want false")
		}
	})

	t.Run("vuecorevm reads comments and identity file", func(t *testing.T) {
		h := byAlias["vuecorevm"]
		if h.HostName != "10.10.2.4" {
			t.Errorf("HostName = %q", h.HostName)
		}
		if h.User != "vueadmin" {
			t.Errorf("User = %q, want vueadmin (own block should win over Host *)", h.User)
		}
		if !h.UseKey {
			t.Errorf("UseKey = false, want true")
		}
		if !h.VPN {
			t.Errorf("VPN = false, want true")
		}
		wantDesc := "Servidor Docker Core (Keycloak, DNS, Nginx e Grafana)"
		if h.Description != wantDesc {
			t.Errorf("Description = %q, want %q", h.Description, wantDesc)
		}
		wantVPNName := "OpenVPN Vuepay | Azure VPN PD Case | OpenVPN PD Case"
		if h.VPNName != wantVPNName {
			t.Errorf("VPNName = %q, want %q", h.VPNName, wantVPNName)
		}
	})

	t.Run("host without HostName falls back to alias", func(t *testing.T) {
		h := byAlias["no-comments"]
		if h.HostName != "10.10.2.9" {
			t.Errorf("HostName = %q", h.HostName)
		}
		if h.User != "defaultuser" {
			t.Errorf("User = %q, want defaultuser (inherited from Host *)", h.User)
		}
		if h.UseKey {
			t.Errorf("UseKey = true, want false")
		}
		if h.Description != "" || h.VPN {
			t.Errorf("expected no description/VPN, got %+v", h)
		}
	})

	t.Run("shared block applies to every alias, HostName falls back per-alias", func(t *testing.T) {
		a := byAlias["web-a"]
		b := byAlias["web-b"]
		if a.HostName != "web-a" || b.HostName != "web-b" {
			t.Errorf("HostName fallback wrong: a=%q b=%q", a.HostName, b.HostName)
		}
		if a.User != "www" || b.User != "www" {
			t.Errorf("User not applied to both aliases: a=%q b=%q", a.User, b.User)
		}
		if a.Description != "Shared block with two aliases" {
			t.Errorf("Description = %q", a.Description)
		}
	})

	t.Run("wildcard pattern block (not just Host *) is inherited and not listed", func(t *testing.T) {
		h := byAlias["bastion-01"]
		if h.HostName != "10.10.3.1" {
			t.Errorf("HostName = %q", h.HostName)
		}
		if !h.UseKey {
			t.Errorf("UseKey = false, want true (inherited from Host bastion-*)")
		}
		if _, ok := byAlias["bastion-*"]; ok {
			t.Errorf("wildcard pattern must not appear as a selectable host")
		}
	})
}

func TestLoad_MissingFile(t *testing.T) {
	if _, err := Load("testdata/does-not-exist"); err == nil {
		t.Fatal("expected an error for a missing config file, got nil")
	}
}

func TestSplitDirective(t *testing.T) {
	cases := []struct {
		line      string
		wantKey   string
		wantValue string
	}{
		{"HostName 10.10.2.4", "hostname", "10.10.2.4"},
		{"User=vueadmin", "user", "vueadmin"},
		{"User = vueadmin", "user", "vueadmin"},
		{"IdentityFile  ~/.ssh/key.pem", "identityfile", "~/.ssh/key.pem"},
	}
	for _, c := range cases {
		key, value := splitDirective(c.line)
		if key != c.wantKey || value != c.wantValue {
			t.Errorf("splitDirective(%q) = (%q, %q), want (%q, %q)", c.line, key, value, c.wantKey, c.wantValue)
		}
	}
}

func TestApplyComments_ToleratesMissingColon(t *testing.T) {
	h := Host{}
	applyComments(&h, []string{
		"# Description Servidor sem dois-pontos",
		"# VPN: true",
	})
	if h.Description != "Servidor sem dois-pontos" {
		t.Errorf("Description = %q", h.Description)
	}
	if !h.VPN {
		t.Errorf("VPN = false, want true")
	}
}

func TestApplyComments_VPNNameImpliesVPN(t *testing.T) {
	h := Host{}
	applyComments(&h, []string{
		"# VPN Name: OpenVPN Vuepay",
	})
	if !h.VPN {
		t.Errorf("VPN = false, want true (implied by VPN Name)")
	}
	if h.VPNName != "OpenVPN Vuepay" {
		t.Errorf("VPNName = %q", h.VPNName)
	}
}
