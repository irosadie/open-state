package capability

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestEgressPolicyBlocksLocalAndPrivateNetworksByDefault(t *testing.T) {
	policy, err := NewEgressPolicy("production", []string{"https"}, []int{443}, nil, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"http://127.0.0.1:8031/mcp", "https://10.0.0.5/mcp", "https://169.254.169.254/latest/meta-data"} {
		if err := policy.ValidateURL(context.Background(), raw); err == nil {
			t.Fatalf("egress unexpectedly allowed %s", raw)
		}
	}
}

func TestEgressPolicyAllowsOnlyExplicitLoopbackDevelopmentPort(t *testing.T) {
	policy, err := NewEgressPolicy("development", []string{"http"}, []int{8031}, nil, nil, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := policy.ValidateURL(context.Background(), "http://127.0.0.1:8031/mcp"); err != nil {
		t.Fatalf("local development endpoint blocked: %v", err)
	}
	if err := policy.ValidateURL(context.Background(), "http://127.0.0.1:8040/mcp"); err == nil {
		t.Fatal("unlisted local port allowed")
	}
}

func TestEgressPolicyRejectsReboundLoopbackAddress(t *testing.T) {
	policy, err := NewEgressPolicy("production", []string{"https"}, []int{443}, nil, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if policy.allowedIP("provider.example.com", net.ParseIP("127.0.0.1")) {
		t.Fatal("DNS-rebound loopback address was accepted")
	}
}

func TestEgressPolicyRedirectCheckRevalidatesTarget(t *testing.T) {
	policy, err := NewEgressPolicy("production", []string{"https"}, []int{443}, nil, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	client := policy.HTTPClient(context.Background())
	u, _ := url.Parse("https://127.0.0.1/mcp")
	req := &http.Request{URL: u}
	if err := client.CheckRedirect(req, nil); err == nil {
		t.Fatal("redirect to loopback unexpectedly allowed")
	}
}

func TestEgressPolicyBoundsProviderResponseBodies(t *testing.T) {
	body := &boundedResponseBody{ReadCloser: io.NopCloser(strings.NewReader("1234")), remaining: 3}
	if _, err := io.ReadAll(body); err == nil { t.Fatal("oversized provider response was accepted") }
}
