package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirewallRule_UnmarshalFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/firewall_rules.json")
	require.NoError(t, err)

	var resp APIResponse[FirewallRule]
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 3)

	// First rule: rule_index as string "2000"
	allow := resp.Data[0]
	assert.Equal(t, "a1b2c3d4e5f6a7b8c9d0e1f2", allow.ID)
	assert.Equal(t, "Allow Established", allow.Name)
	assert.Equal(t, "accept", allow.Action)
	assert.True(t, allow.Enabled)
	assert.Equal(t, 2000, allow.RuleIndex.Int())
	assert.Equal(t, "LAN_IN", allow.Ruleset)
	assert.True(t, allow.StateEstablished)
	assert.True(t, allow.StateRelated)
	assert.False(t, allow.StateNew)
	assert.False(t, allow.StateInvalid)
	assert.False(t, allow.Logging)

	// Second rule: rule_index as integer 2001
	drop := resp.Data[1]
	assert.Equal(t, "Drop Invalid", drop.Name)
	assert.Equal(t, "drop", drop.Action)
	assert.Equal(t, 2001, drop.RuleIndex.Int())
	assert.True(t, drop.StateInvalid)
	assert.True(t, drop.Logging)

	// Third rule: has network IDs
	block := resp.Data[2]
	assert.Equal(t, "Block Guest to LAN", block.Name)
	assert.Equal(t, "reject", block.Action)
	assert.Equal(t, 3000, block.RuleIndex.Int())
	assert.Equal(t, "6a3b4c5d6e7f8a9b0c1d2e3f", block.SrcNetworkID)
	assert.Equal(t, "5f2e4a3b1c9d8e7f6a5b4c3d", block.DstNetworkID)
	assert.True(t, block.StateNew)
}

func TestFirewallRule_RuleIndexStringVsInt(t *testing.T) {
	// Explicitly test that FlexInt handles both formats in the
	// context of a FirewallRule struct
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "rule_index as string",
			input: `{"_id":"test1","name":"test","action":"accept","enabled":true,"rule_index":"2000","ruleset":"LAN_IN","state_new":false,"state_established":false,"state_invalid":false,"state_related":false,"logging":false,"site_id":"site1"}`,
			want:  2000,
		},
		{
			name:  "rule_index as integer",
			input: `{"_id":"test2","name":"test","action":"drop","enabled":true,"rule_index":3001,"ruleset":"LAN_IN","state_new":false,"state_established":false,"state_invalid":false,"state_related":false,"logging":false,"site_id":"site1"}`,
			want:  3001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rule FirewallRule
			err := json.Unmarshal([]byte(tt.input), &rule)
			require.NoError(t, err)
			assert.Equal(t, tt.want, rule.RuleIndex.Int())
		})
	}
}

func TestFirewallRule_RoundTrip(t *testing.T) {
	original := FirewallRule{
		ID:               "abc123",
		Name:             "Test Rule",
		Action:           "accept",
		Enabled:          true,
		RuleIndex:        FlexInt(5000),
		Ruleset:          "LAN_IN",
		Protocol:         "tcp",
		SrcAddress:       "10.0.0.0/8",
		DstPort:          "443",
		StateNew:         true,
		StateEstablished: true,
		Logging:          true,
		SiteID:           "site123",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FirewallRule
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestFirewallRule_OmitEmptyFields(t *testing.T) {
	rule := FirewallRule{
		ID:        "abc123",
		Name:      "Minimal",
		Action:    "accept",
		Enabled:   true,
		RuleIndex: FlexInt(1000),
		Ruleset:   "LAN_IN",
		SiteID:    "site123",
	}

	data, err := json.Marshal(rule)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	// omitempty fields should be absent
	assert.NotContains(t, raw, "protocol")
	assert.NotContains(t, raw, "src_address")
	assert.NotContains(t, raw, "dst_address")
	assert.NotContains(t, raw, "src_port")
	assert.NotContains(t, raw, "dst_port")
	assert.NotContains(t, raw, "src_networkconf_id")
	assert.NotContains(t, raw, "dst_networkconf_id")
	assert.NotContains(t, raw, "src_networkconf_type")
	assert.NotContains(t, raw, "dst_networkconf_type")
	assert.NotContains(t, raw, "icmp_typename")
}

func TestFirewallGroup_UnmarshalFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/firewall_groups.json")
	require.NoError(t, err)

	var resp APIResponse[FirewallGroup]
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 2)

	// Address group
	addrGroup := resp.Data[0]
	assert.Equal(t, "d4e5f6a7b8c9d0e1f2a3b4c5", addrGroup.ID)
	assert.Equal(t, "RFC1918 Addresses", addrGroup.Name)
	assert.Equal(t, "address-group", addrGroup.GroupType)
	assert.Equal(t, []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}, addrGroup.GroupMembers)
	assert.Equal(t, "6a1b2c3d4e5f6a7b8c9d0e1f", addrGroup.SiteID)

	// Port group
	portGroup := resp.Data[1]
	assert.Equal(t, "Web Ports", portGroup.Name)
	assert.Equal(t, "port-group", portGroup.GroupType)
	assert.Equal(t, []string{"80", "443", "8080", "8443"}, portGroup.GroupMembers)
}

func TestFirewallGroup_RoundTrip(t *testing.T) {
	original := FirewallGroup{
		ID:           "abc123",
		Name:         "Test Group",
		GroupType:    "address-group",
		GroupMembers: []string{"10.0.0.0/8", "172.16.0.0/12"},
		SiteID:       "site123",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FirewallGroup
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}
