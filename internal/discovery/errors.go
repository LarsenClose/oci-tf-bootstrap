package discovery

import (
	"errors"
	"fmt"
	"net"

	"github.com/oracle/oci-go-sdk/v65/common"
)

// classifiedError wraps an OCI error with user-friendly guidance while preserving
// the original error for programmatic inspection via errors.Unwrap.
type classifiedError struct {
	msg   string
	cause error
}

func (e *classifiedError) Error() string { return e.msg }
func (e *classifiedError) Unwrap() error { return e.cause }

// classifyOCIError examines an OCI SDK error and returns a user-friendly message
// with actionable guidance appended to the original error.
func classifyOCIError(resource string, err error) error {
	if err == nil {
		return nil
	}

	var svcErr common.ServiceError
	if errors.As(err, &svcErr) {
		return classifyServiceError(resource, svcErr, err)
	}

	// Check for network-level errors (connection refused, timeout, DNS failure).
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return &classifiedError{
			msg:   fmt.Sprintf("%s: %v. Could not connect to OCI API. Check your internet connection and that your region is correct.", resource, err),
			cause: err,
		}
	}

	// Default: wrap with resource context only.
	return &classifiedError{
		msg:   fmt.Sprintf("%s: %v", resource, err),
		cause: err,
	}
}

func classifyServiceError(resource string, svcErr common.ServiceError, original error) error {
	code := svcErr.GetHTTPStatusCode()
	var guidance string

	switch {
	case code == 401:
		guidance = "Check your OCI config file (~/.oci/config): verify the key_file path exists and the fingerprint matches your API key in the OCI Console (Identity > Users > API Keys)."
	case code == 403:
		guidance = fmt.Sprintf("Your OCI user does not have permission to access %s. Check your IAM policies in the OCI Console, or try running with --compartment to target a specific compartment you have access to.", resource)
	case code == 404:
		guidance = fmt.Sprintf("The %s was not found. Verify your compartment OCID and region are correct. Current region can be checked with `oci iam region-subscription list`.", resource)
	case code == 429:
		guidance = fmt.Sprintf("OCI API rate limit exceeded while discovering %s. Wait a few minutes and try again, or use --compartment to reduce the scope of discovery.", resource)
	case code >= 500:
		guidance = fmt.Sprintf("OCI service error while discovering %s. This is an OCI-side issue. Check https://ocistatus.oraclecloud.com/ and retry later.", resource)
	default:
		// Unknown status code: wrap with resource context only.
		return &classifiedError{
			msg:   fmt.Sprintf("%s: %v", resource, original),
			cause: original,
		}
	}

	return &classifiedError{
		msg:   fmt.Sprintf("%s: %v. %s", resource, original, guidance),
		cause: original,
	}
}
