package discovery

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
)

// mockServiceError implements common.ServiceError for testing.
type mockServiceError struct {
	statusCode int
	code       string
	message    string
}

func (e *mockServiceError) GetHTTPStatusCode() int  { return e.statusCode }
func (e *mockServiceError) GetCode() string         { return e.code }
func (e *mockServiceError) GetMessage() string      { return e.message }
func (e *mockServiceError) Error() string           { return e.message }
func (e *mockServiceError) GetOpcRequestID() string { return "test-request-id" }

func TestClassifyOCIError_Nil(t *testing.T) {
	if classifyOCIError("test", nil) != nil {
		t.Error("classifyOCIError(nil) should return nil")
	}
}

func TestClassifyOCIError_401(t *testing.T) {
	err := &mockServiceError{statusCode: 401, code: "NotAuthenticated", message: "not authenticated"}
	result := classifyOCIError("tenancy details", err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	if !strings.Contains(msg, "tenancy details") {
		t.Errorf("expected resource name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "key_file") {
		t.Errorf("expected auth guidance about key_file, got: %s", msg)
	}
	if !strings.Contains(msg, "fingerprint") {
		t.Errorf("expected auth guidance about fingerprint, got: %s", msg)
	}
}

func TestClassifyOCIError_403(t *testing.T) {
	err := &mockServiceError{statusCode: 403, code: "NotAuthorized", message: "not authorized"}
	result := classifyOCIError("shapes", err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	if !strings.Contains(msg, "shapes") {
		t.Errorf("expected resource name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "permission") {
		t.Errorf("expected permission guidance, got: %s", msg)
	}
	if !strings.Contains(msg, "--compartment") {
		t.Errorf("expected --compartment suggestion, got: %s", msg)
	}
}

func TestClassifyOCIError_404(t *testing.T) {
	err := &mockServiceError{statusCode: 404, code: "NotFound", message: "not found"}
	result := classifyOCIError("compartments", err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	if !strings.Contains(msg, "compartments") {
		t.Errorf("expected resource name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "not found") {
		t.Errorf("expected not-found guidance, got: %s", msg)
	}
	if !strings.Contains(msg, "compartment OCID") {
		t.Errorf("expected compartment OCID guidance, got: %s", msg)
	}
}

func TestClassifyOCIError_429(t *testing.T) {
	err := &mockServiceError{statusCode: 429, code: "TooManyRequests", message: "too many requests"}
	result := classifyOCIError("VCN discovery", err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	if !strings.Contains(msg, "rate limit") {
		t.Errorf("expected rate limit guidance, got: %s", msg)
	}
	if !strings.Contains(msg, "VCN discovery") {
		t.Errorf("expected resource name in message, got: %s", msg)
	}
}

func TestClassifyOCIError_500(t *testing.T) {
	err := &mockServiceError{statusCode: 500, code: "InternalServerError", message: "internal error"}
	result := classifyOCIError("images", err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	if !strings.Contains(msg, "OCI service error") {
		t.Errorf("expected server error guidance, got: %s", msg)
	}
	if !strings.Contains(msg, "ocistatus.oraclecloud.com") {
		t.Errorf("expected status page link, got: %s", msg)
	}
}

func TestClassifyOCIError_503(t *testing.T) {
	err := &mockServiceError{statusCode: 503, code: "ServiceUnavailable", message: "service unavailable"}
	result := classifyOCIError("limits", err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	if !strings.Contains(msg, "OCI service error") {
		t.Errorf("expected server error guidance for 503, got: %s", msg)
	}
}

func TestClassifyOCIError_NetworkError(t *testing.T) {
	netErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: fmt.Errorf("connection refused"),
	}
	result := classifyOCIError("tenancy details", netErr)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	if !strings.Contains(msg, "Could not connect") {
		t.Errorf("expected connection guidance, got: %s", msg)
	}
	if !strings.Contains(msg, "internet connection") {
		t.Errorf("expected internet connection guidance, got: %s", msg)
	}
}

func TestClassifyOCIError_PlainError(t *testing.T) {
	err := fmt.Errorf("something unexpected happened")
	result := classifyOCIError("shapes", err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	if !strings.Contains(msg, "shapes") {
		t.Errorf("expected resource name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "something unexpected happened") {
		t.Errorf("expected original message preserved, got: %s", msg)
	}
}

func TestClassifyOCIError_Unwrap(t *testing.T) {
	original := &mockServiceError{statusCode: 403, code: "NotAuthorized", message: "not authorized"}
	result := classifyOCIError("test", original)
	if result == nil {
		t.Fatal("expected non-nil error")
	}

	var unwrapped *mockServiceError
	if !errors.As(result, &unwrapped) {
		t.Error("classified error should be unwrappable to the original error type")
	}
	if unwrapped.GetHTTPStatusCode() != 403 {
		t.Errorf("unwrapped error should preserve status code, got %d", unwrapped.GetHTTPStatusCode())
	}
}

func TestClassifyOCIError_UnknownStatusCode(t *testing.T) {
	err := &mockServiceError{statusCode: 418, code: "Teapot", message: "I'm a teapot"}
	result := classifyOCIError("shapes", err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	msg := result.Error()
	// Should contain resource context and original message but no specific guidance
	if !strings.Contains(msg, "shapes") {
		t.Errorf("expected resource name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "I'm a teapot") {
		t.Errorf("expected original message, got: %s", msg)
	}
	// Should NOT contain any of the specific guidance strings
	if strings.Contains(msg, "key_file") || strings.Contains(msg, "permission") || strings.Contains(msg, "rate limit") {
		t.Errorf("unknown status code should not produce specific guidance, got: %s", msg)
	}
}
