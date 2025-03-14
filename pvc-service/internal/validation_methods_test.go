package internal

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsValidKubernetesName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		errorCode codes.Code
	}{
		{"ValidName", "valid-name", false, codes.OK},
		{"EmptyName", "", true, codes.InvalidArgument},
		{"TooLongName", string(make([]byte, 254)), true, codes.InvalidArgument},
		{"InvalidCharacter", "Invalid*Name", true, codes.InvalidArgument},
		{"EndsWithDash", "invalid-name-", true, codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsValidKubernetesName(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none for input: %s", tt.input)
				} else {
					if st, ok := status.FromError(err); ok {
						if st.Code() != tt.errorCode {
							t.Errorf("Expected error code %v, but got %v", tt.errorCode, st.Code())
						}
					} else {
						t.Errorf("Expected grpc status error but got %v", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got %v for input: %s", err, tt.input)
				}
			}
		})
	}
}

func TestIsValidSize(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string // Changed to string to match the return type of IsValidSize
	}{
		{"ValidSize", "10", "10"},
		{"ZeroSize", "0", ""},
		{"NegativeSize", "-5", ""},
		{"EmptySize", "", "2"}, // Assuming empty should default to "2"
		{"NonNumericSize", "abc", ""},
		{"LargeSize", "999999", "999999"},
		{"Whitespace", " 5 ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidSize(&tt.input); got != tt.expect {
				t.Errorf("IsValidSize(%q) = %v, expect %v", tt.input, got, tt.expect)
			}
		})
	}
}

