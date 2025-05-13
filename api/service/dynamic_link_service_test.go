package service

import (
	"net/http"
	"net/url"
	"testing"

	"dynamic-link-service/config"
)

func TestIsPreviewHost(t *testing.T) {
	tests := []struct {
		name           string
		previewStyle   string
		host           string
		expectedResult bool
		expectError    bool
	}{
		// Hyphenated style tests
		{
			name:           "hyphenated style - valid preview host",
			previewStyle:   "hyphenated",
			host:           "myapp-preview.example.com",
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "hyphenated style - non-preview host",
			previewStyle:   "hyphenated",
			host:           "myapp.example.com",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "hyphenated style - invalid host format",
			previewStyle:   "hyphenated",
			host:           "example",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "hyphenated style - preview in wrong position",
			previewStyle:   "hyphenated",
			host:           "preview-myapp.example.com",
			expectedResult: false,
			expectError:    false,
		},

		// Subdomain style tests
		{
			name:           "subdomain style - valid preview host",
			previewStyle:   "subdomain",
			host:           "preview.example.com",
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "subdomain style - non-preview host",
			previewStyle:   "subdomain",
			host:           "myapp.example.com",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "subdomain style - invalid host format",
			previewStyle:   "subdomain",
			host:           "example",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "subdomain style - preview in wrong position",
			previewStyle:   "subdomain",
			host:           "myapp.preview.example.com",
			expectedResult: false,
			expectError:    false,
		},

		// Invalid style tests
		{
			name:           "invalid style - should return error",
			previewStyle:   "invalid",
			host:           "preview.example.com",
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://"+tt.host, nil)
			if err != nil {
				t.Fatalf("Failed to create test request: %v", err)
			}

			service := &DynamicLinkService{
				config: &config.Config{
					PreviewUrlStyle: tt.previewStyle,
				},
			}

			result, err := service.IsPreviewHost(req)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expectedResult {
				t.Errorf("IsPreviewHost() = %v, want %v for host %s with style %s",
					result, tt.expectedResult, tt.host, tt.previewStyle)
			}
		})
	}
}

func TestGeneratePreviewURL(t *testing.T) {
	tests := []struct {
		name         string
		previewStyle string
		originalURL  string
		expectedURL  string
		expectError  bool
	}{
		// Hyphenated style tests
		{
			name:         "hyphenated style - simple domain",
			previewStyle: "hyphenated",
			originalURL:  "http://myapp.example.com",
			expectedURL:  "http://myapp-preview.example.com",
			expectError:  false,
		},
		{
			name:         "hyphenated style - with path and query",
			previewStyle: "hyphenated",
			originalURL:  "https://myapp.example.com/path?query=value",
			expectedURL:  "https://myapp-preview.example.com/path?query=value",
			expectError:  false,
		},

		// Subdomain style tests
		{
			name:         "subdomain style - simple domain",
			previewStyle: "subdomain",
			originalURL:  "http://myapp.example.com",
			expectedURL:  "http://preview.myapp.example.com",
			expectError:  false,
		},
		{
			name:         "subdomain style - with path and query",
			previewStyle: "subdomain",
			originalURL:  "https://myapp.example.com/path?query=value",
			expectedURL:  "https://preview.myapp.example.com/path?query=value",
			expectError:  false,
		},

		// Error cases
		{
			name:         "invalid style",
			previewStyle: "invalid",
			originalURL:  "http://myapp.example.com",
			expectedURL:  "",
			expectError:  true,
		},
		{
			name:         "invalid URL format",
			previewStyle: "hyphenated",
			originalURL:  "not-a-url",
			expectedURL:  "",
			expectError:  true,
		},
		{
			name:         "empty host",
			previewStyle: "hyphenated",
			originalURL:  "http://",
			expectedURL:  "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalURL, err := url.Parse(tt.originalURL)
			if err != nil && !tt.expectError {
				t.Fatalf("Failed to parse original URL: %v", err)
			}

			service := &DynamicLinkService{
				config: &config.Config{
					PreviewUrlStyle: tt.previewStyle,
				},
			}

			previewURL, err := service.GeneratePreviewURL(originalURL)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if previewURL.String() != tt.expectedURL {
				t.Errorf("GeneratePreviewURL() = %v, want %v", previewURL.String(), tt.expectedURL)
			}
		})
	}
}

func TestGetNonPreviewHost(t *testing.T) {
	tests := []struct {
		name         string
		previewStyle string
		host         string
		expectedHost string
		expectError  bool
	}{
		// Hyphenated style tests
		{
			name:         "hyphenated style - valid preview host",
			previewStyle: "hyphenated",
			host:         "myapp-preview.example.com",
			expectedHost: "myapp.example.com",
			expectError:  false,
		},
		{
			name:         "hyphenated style - non-preview host",
			previewStyle: "hyphenated",
			host:         "myapp.example.com",
			expectedHost: "myapp.example.com",
			expectError:  false,
		},
		{
			name:         "hyphenated style - single part host",
			previewStyle: "hyphenated",
			host:         "example",
			expectedHost: "example",
			expectError:  false,
		},

		// Subdomain style tests
		{
			name:         "subdomain style - valid preview host",
			previewStyle: "subdomain",
			host:         "preview.example.com",
			expectedHost: "example.com",
			expectError:  false,
		},
		{
			name:         "subdomain style - non-preview host",
			previewStyle: "subdomain",
			host:         "myapp.example.com",
			expectedHost: "myapp.example.com",
			expectError:  false,
		},
		{
			name:         "subdomain style - single part host",
			previewStyle: "subdomain",
			host:         "example",
			expectedHost: "example",
			expectError:  false,
		},

		// Invalid style tests
		{
			name:         "invalid style - should return error",
			previewStyle: "invalid",
			host:         "preview.example.com",
			expectedHost: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &DynamicLinkService{
				config: &config.Config{
					PreviewUrlStyle: tt.previewStyle,
				},
			}

			result, err := service.GetNonPreviewHost(tt.host)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expectedHost {
				t.Errorf("GetNonPreviewHost() = %v, want %v for host %s with style %s",
					result, tt.expectedHost, tt.host, tt.previewStyle)
			}
		})
	}
}
