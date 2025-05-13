package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"dynamic-link-redirect/api/model"
	"dynamic-link-redirect/config"

	"github.com/rs/zerolog/log"
)

type DynamicLinkService struct {
	config *config.Config
}

func NewDynamicLinkService(config *config.Config) *DynamicLinkService {
	return &DynamicLinkService{config: config}
}

func (s *DynamicLinkService) GetQueryParamsFromURL(ctx context.Context, url *url.URL) (url.Values, error) {
	log.Debug().Str("url", url.String()).Msg("Getting query params from url")

	reqBody := model.ExchangeShortLinkRequest{
		RequestedLink: url.String(),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request body")
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(s.config.ExchangeShortLinkEndpoint, "application/json", strings.NewReader(string(jsonBody)))
	if err != nil {
		log.Error().Err(err).Msg("Failed to make POST request")
		return nil, fmt.Errorf("failed to make POST request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response model.LongLinkResponseModel
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.LongLink == "" {
		return nil, nil
	}

	parsedURL, err := url.Parse(response.LongLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse long link: %w", err)
	}

	return parsedURL.Query(), nil
}

func (s *DynamicLinkService) GetNonPreviewHost(host string) (string, error) {
	switch s.config.PreviewUrlStyle {
	case "hyphenated":
		hostParts := strings.SplitN(host, ".", 2)
		if len(hostParts) < 2 {
			return host, nil
		}
		subdomain := hostParts[0]
		if strings.HasSuffix(subdomain, "-preview") {
			cleanSubdomain := strings.TrimSuffix(subdomain, "-preview")
			return cleanSubdomain + "." + hostParts[1], nil
		}
		return host, nil

	case "subdomain":
		hostParts := strings.SplitN(host, ".", 2)
		if len(hostParts) < 2 {
			return host, nil
		}
		if hostParts[0] == "preview" {
			return hostParts[1], nil
		}
		return host, nil

	default:
		return "", fmt.Errorf("invalid preview URL style: %s", s.config.PreviewUrlStyle)
	}
}

func (s *DynamicLinkService) IsPreviewHost(r *http.Request) (bool, error) {
	host := r.Host

	log.Debug().
		Str("url", r.URL.String()).
		Str("host", r.Host).
		Str("referer", r.Referer()).
		Msg("Incoming request")

	log.Debug().Str("host", host).Msg("Checking if host is a preview host")

	switch s.config.PreviewUrlStyle {
	case "hyphenated":
		hostParts := strings.Split(host, ".")
		if len(hostParts) < 2 {
			log.Debug().Str("host", host).Msg("Host has insufficient parts")
			return false, nil
		}
		subdomain := hostParts[0]
		if strings.HasSuffix(subdomain, "-preview") {
			log.Debug().Str("subdomain", subdomain).Msg("Preview host detected (hyphenated style)")
			return true, nil
		}
	case "subdomain":
		hostParts := strings.Split(host, ".")
		if len(hostParts) < 2 {
			log.Debug().Str("host", host).Msg("Host has insufficient parts")
			return false, nil
		}

		if hostParts[0] == "preview" {
			log.Debug().Str("subdomain", hostParts[0]).Msg("Preview host detected (subdomain style)")
			return true, nil
		}
	default:
		log.Error().Str("style", s.config.PreviewUrlStyle).Msg("Invalid preview URL style configuration")
		return false, fmt.Errorf("invalid preview URL style: %s", s.config.PreviewUrlStyle)
	}

	log.Debug().Str("host", host).Msg("Not a preview host")
	return false, nil
}

func (s *DynamicLinkService) GeneratePreviewURL(originalURL *url.URL) (*url.URL, error) {
	if originalURL == nil {
		return nil, fmt.Errorf("original URL cannot be nil")
	}

	previewURL := *originalURL

	hostParts := strings.Split(previewURL.Host, ".")
	if len(hostParts) < 2 {
		return nil, fmt.Errorf("invalid host format: %s", previewURL.Host)
	}

	switch s.config.PreviewUrlStyle {
	case "hyphenated":
		hostParts[0] = hostParts[0] + "-preview"
	case "subdomain":
		hostParts = append([]string{"preview"}, hostParts...)
	default:
		return nil, fmt.Errorf("invalid preview URL style: %s", s.config.PreviewUrlStyle)
	}

	previewURL.Host = strings.Join(hostParts, ".")
	return &previewURL, nil
}
