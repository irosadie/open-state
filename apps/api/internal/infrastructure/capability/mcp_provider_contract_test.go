package capability

import (
	"context"
	"testing"
)

func TestValidateMCPProviderContractRejectsMissingConfiguration(t *testing.T) {
	if err := ValidateMCPProviderContract(context.Background(), "", "provider-mock", "padel.cek_available"); err == nil {
		t.Fatal("expected missing provider configuration error")
	}
}

func TestValidateMCPProviderContractChecksProviderIdentityAndTool(t *testing.T) {
	endpoint := startProviderMock(t, "padel.json")

	if err := ValidateMCPProviderContract(context.Background(), endpoint, "padel-provider-mock", "padel.cek_available"); err != nil {
		t.Fatalf("expected declared provider contract to pass: %v", err)
	}
	if err := ValidateMCPProviderContract(context.Background(), endpoint, "doctor-provider-mock", "padel.cek_available"); err == nil {
		t.Fatal("expected provider alias isolation failure")
	}
	if err := ValidateMCPProviderContract(context.Background(), endpoint, "padel-provider-mock", "padel.book"); err == nil {
		t.Fatal("expected missing provider tool failure")
	}
}

func TestValidateMCPProviderContractReportsDiscoveryFailure(t *testing.T) {
	if err := ValidateMCPProviderContract(context.Background(), "http://127.0.0.1:1/mcp", "padel-provider-mock", "padel.cek_available"); err == nil {
		t.Fatal("expected provider discovery failure")
	}
}
