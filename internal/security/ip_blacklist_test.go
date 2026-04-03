package security

import (
	"net"
	"testing"
)

func TestBlacklistIP_Validation(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"not-an-ip", false},
		{"", false},
		{"999.999.999.999", false},
		{"192.168.1", false},
	}

	for _, tt := range tests {
		isValid := net.ParseIP(tt.ip) != nil
		if isValid != tt.valid {
			t.Errorf("IP %s: expected valid=%v, got %v", tt.ip, tt.valid, isValid)
		}
	}
}

func TestBlacklistedIP_Struct(t *testing.T) {
	ip := BlacklistedIP{
		ID:        "uuid-001",
		IPAddress: "192.168.1.100",
		Reason:    "Suspicious login attempts",
		AddedBy:   "admin-001",
	}

	if ip.ID != "uuid-001" {
		t.Errorf("Expected ID uuid-001, got: %s", ip.ID)
	}
	if ip.IPAddress != "192.168.1.100" {
		t.Errorf("Expected IPAddress 192.168.1.100, got: %s", ip.IPAddress)
	}
	if ip.Reason != "Suspicious login attempts" {
		t.Errorf("Expected reason, got: %s", ip.Reason)
	}
	if ip.AddedBy != "admin-001" {
		t.Errorf("Expected AddedBy admin-001, got: %s", ip.AddedBy)
	}
}
