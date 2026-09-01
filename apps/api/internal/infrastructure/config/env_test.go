package config

import "testing"

func setRequiredConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://localhost/openstate")
	t.Setenv("JWT_SECRET", "test-jwt-secret-that-is-at-least-32-characters")
	t.Setenv("MCP_API_KEY_PEPPER", "test-mcp-api-key-pepper-that-is-at-least-32-characters")
}

func TestLoadDefaultsToAdvisoryGateway(t *testing.T) {
	setRequiredConfigEnv(t)
	t.Setenv("MCP_GATEWAY_MODE", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.MCPGatewayMode != "advisory" {
		t.Fatalf("gateway mode = %q", cfg.MCPGatewayMode)
	}
}

func TestLoadAcceptsSecureGateway(t *testing.T) {
	setRequiredConfigEnv(t)
	t.Setenv("MCP_GATEWAY_MODE", "SECURE")
	cfg, err := Load()
	if err != nil || cfg.MCPGatewayMode != "secure" {
		t.Fatalf("config=%#v err=%v", cfg, err)
	}
}

func TestLoadRejectsUnknownGatewayMode(t *testing.T) {
	setRequiredConfigEnv(t)
	t.Setenv("MCP_GATEWAY_MODE", "passthrough")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid gateway mode error")
	}
}
