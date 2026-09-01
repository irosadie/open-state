package capability

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
)

func TestMCPProviderInvokeProviderMock(t *testing.T) {
	endpoint := startProviderMock(t, "padel.json")
	provider := NewMCPProvider(MCPProviderConfig{
		Endpoint: endpoint,
		ToolName: "padel.cek_available",
		Timeout:  3 * time.Second,
	}, nil)

	result, err := provider.Invoke(context.Background(), domaincap.Invocation{
		Name: "padel.availability.read",
		Payload: map[string]any{
			"date":     "2026-09-01",
			"venue_id": "padel-senayan",
		},
	})
	if err != nil {
		t.Fatalf("invoke provider mock: %v", err)
	}
	if result.FromMock {
		t.Fatal("MCP provider adapter must report a real provider result")
	}
	if result.CapabilityEvent == nil || *result.CapabilityEvent != "capability.success" {
		t.Fatalf("expected success event, got %v", result.CapabilityEvent)
	}
	if result.Data["text"] == "" {
		t.Fatalf("expected normalized provider data, got %#v", result.Data)
	}
}

func TestMCPProviderMockToolError(t *testing.T) {
	endpoint := startProviderMock(t, "padel-error.json")
	provider := NewMCPProvider(MCPProviderConfig{
		Endpoint: endpoint,
		ToolName: "padel.cek_available",
		Timeout:  3 * time.Second,
	}, nil)

	_, err := provider.Invoke(context.Background(), domaincap.Invocation{
		Name: "padel.availability.read",
		Payload: map[string]any{
			"date":     "2026-09-01",
			"venue_id": "padel-senayan",
		},
	})
	if err == nil {
		t.Fatal("expected provider tool error")
	}

	var capabilityErr *domaincap.CapabilityError
	if !asCapErr(err, &capabilityErr) {
		t.Fatalf("expected CapabilityError, got %T", err)
	}
	if capabilityErr.Kind != domaincap.ErrorKindBusiness {
		t.Fatalf("expected business error, got %s", capabilityErr.Kind)
	}
}

func TestMCPProviderMockTimeout(t *testing.T) {
	endpoint := startProviderMock(t, "padel-delay.json")
	provider := NewMCPProvider(MCPProviderConfig{
		Endpoint: endpoint,
		ToolName: "padel.cek_available",
		Timeout:  25 * time.Millisecond,
	}, nil)

	_, err := provider.Invoke(context.Background(), domaincap.Invocation{
		Name: "padel.availability.read",
		Payload: map[string]any{
			"date":     "2026-09-01",
			"venue_id": "padel-senayan",
		},
	})
	if err == nil {
		t.Fatal("expected provider timeout")
	}

	var capabilityErr *domaincap.CapabilityError
	if !asCapErr(err, &capabilityErr) {
		t.Fatalf("expected CapabilityError, got %T", err)
	}
	if capabilityErr.Kind != domaincap.ErrorKindTimeout {
		t.Fatalf("expected timeout error, got %s", capabilityErr.Kind)
	}
}

func startProviderMock(t *testing.T, fixture string) string {
	t.Helper()

	bunPath, err := exec.LookPath("bun")
	if err != nil {
		t.Skip("bun is required to run the MCP provider mock integration tests")
	}

	workspace := providerMockWorkspace(t)
	port := availablePort(t)
	fixturePath := filepath.Join(workspace, "fixtures", fixture)
	command := exec.Command(bunPath, "run", "start")
	command.Dir = workspace
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Env = append(os.Environ(),
		"MCP_PROVIDER_MOCK_PORT="+strconv.Itoa(port),
		"MCP_PROVIDER_MOCK_SCENARIO="+fixturePath,
	)

	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		t.Fatalf("start provider mock: %v", err)
	}
	t.Cleanup(func() {
		if command.Process != nil {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
		_ = command.Wait()
	})

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	waitForProviderMock(t, baseURL, &output)
	return baseURL + "/mcp"
}

func providerMockWorkspace(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve provider mock workspace: runtime caller unavailable")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../..", "apps", "mcp-provider-mock"))
}

func availablePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve test port: %v", err)
	}
	defer listener.Close()

	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatal("reserve test port: unexpected listener address")
	}
	return address.Port
}

func waitForProviderMock(t *testing.T, baseURL string, output *bytes.Buffer) {
	t.Helper()

	client := &http.Client{Timeout: 100 * time.Millisecond}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(baseURL + "/health/ready")
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("provider mock did not become ready: %s", output.String())
}
