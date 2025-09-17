package api

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"dynamic-link-redirect/api/service"
	"dynamic-link-redirect/config"
	"dynamic-link-redirect/utils"

	"github.com/rs/zerolog/log"
)

type DynamicLinkHandler struct {
	service *service.DynamicLinkService
	config  *config.Config
}

func NewDynamicLinkHandler(service *service.DynamicLinkService, config *config.Config) *DynamicLinkHandler {
	return &DynamicLinkHandler{service: service, config: config}
}

func (h *DynamicLinkHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	requestedURL := utils.FullRequestURL(r)
	nonPreviewHost, err := h.service.GetNonPreviewHost(r.Host)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	requestedURL.Host = nonPreviewHost

	dynamicLinkQueryParams, err := h.service.GetQueryParamsFromURL(r.Context(), requestedURL)
	log.Debug().Str("dynamicLinkQueryParams", dynamicLinkQueryParams.Encode()).Msg("Dynamic link query params")

	if err != nil || dynamicLinkQueryParams == nil {
		http.NotFound(w, r)
		return
	}

	requestQueryParams := r.URL.Query()
	log.Debug().Str("request_query_params", requestQueryParams.Encode()).Msg("request query params")

	isPreview, err := h.service.IsPreviewHost(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check if host is preview")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if isPreview && requestQueryParams.Get("from-preview") != "true" {
		urlCopy := utils.FullRequestURL(r)
		newHost, err := h.service.GetNonPreviewHost(urlCopy.Host)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get non preview host")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		urlCopy.Host = newHost

		query := urlCopy.Query()
		query.Set("from-preview", "true")
		urlCopy.RawQuery = query.Encode()

		log.Debug().Str("urlCopy", urlCopy.String()).Msg("Handling preview page")
		handlePreviewPage(w, h.config.AppIconImageURL, h.config.AppName, *urlCopy, dynamicLinkQueryParams.Get("st"), dynamicLinkQueryParams.Get("sd"), dynamicLinkQueryParams.Get("si"))
		return
	}

	userAgent := r.Header.Get("User-Agent")
	isiPad := strings.Contains(userAgent, "iPad")
	isiPhone := strings.Contains(userAgent, "iPhone")
	isIos := isiPad || isiPhone
	isAndroid := strings.Contains(userAgent, "Android")

	if isIos {
		log.Debug().Msg("Handling iOS User Agent")
		fullURL := utils.FullRequestURL(r)
		previewURL, err := h.service.GeneratePreviewURL(fullURL)
		if err != nil {
			log.Error().Err(err).Msg("Failed to transform to preview URL")
			http.Error(w, "Invalid long link format", http.StatusInternalServerError)
			return
		}
		if requestQueryParams.Get("from-preview") != "true" {
			log.Debug().Str("previewURL", previewURL.String()).Msg("Redirecting to preview URL")
			http.Redirect(w, r, previewURL.String(), http.StatusFound)
		} else {
			handleiOSDynamicLink(w, r, dynamicLinkQueryParams, isiPad, isiPhone)
		}
		return
	} else if isAndroid {
		log.Debug().Msg("Handling Android User Agent")
		fullURL := utils.FullRequestURL(r)
		previewURL, err := h.service.GeneratePreviewURL(fullURL)
		if err != nil {
			log.Error().Err(err).Msg("Failed to transform to preview URL")
			http.Error(w, "Invalid long link format", http.StatusInternalServerError)
			return
		}
		if requestQueryParams.Get("from-preview") != "true" {
			log.Debug().Str("previewURL", previewURL.String()).Msg("Redirecting to preview URL")
			http.Redirect(w, r, previewURL.String(), http.StatusFound)
		} else {
			handleAndroidDynamicLink(w, r, dynamicLinkQueryParams, requestedURL)
		}
		return
	} else {
		handleWebUserAgent(w, r, dynamicLinkQueryParams)
		return
	}
}

func handleWebUserAgent(w http.ResponseWriter, r *http.Request, queryParams url.Values) {
	log.Debug().Msg("Handling web user agent")
	ofl := queryParams.Get("ofl")
	if ofl != "" {
		unescapedOfl, err := url.QueryUnescape(ofl)
		if err != nil {
			log.Error().Str("ofl", ofl).Msg("Failed to unescape 'ofl' parameter")
			http.Error(w, "Invalid 'ofl' link format", http.StatusInternalServerError)
			return
		}
		parsedOfl, err := url.Parse(unescapedOfl)
		if err != nil || !parsedOfl.IsAbs() {
			log.Error().Str("ofl", unescapedOfl).Msg("Invalid 'ofl' parameter")
			http.Error(w, "Invalid 'ofl' link format", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, unescapedOfl, http.StatusFound)
		return
	}

	link := queryParams.Get("link")
	if link != "" {
		unescapedLink, err := url.QueryUnescape(link)
		if err != nil {
			log.Error().Str("link", link).Msg("Failed to unescape 'link' parameter")
			http.Error(w, "Invalid 'link' link format", http.StatusInternalServerError)
			return
		}
		parsedOfl, err := url.Parse(unescapedLink)
		if err != nil || !parsedOfl.IsAbs() {
			log.Error().Str("link", unescapedLink).Msg("Invalid 'link' parameter")
			http.Error(w, "Invalid 'link' link format", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, unescapedLink, http.StatusFound)
		return
	}
	http.NotFound(w, r)
}

func handlePreviewPage(w http.ResponseWriter, appIconImageURL, appName string, dynamicLink url.URL, socialTitle, socialDescription, socialImageLink string) {
	log.Debug().Msg("Handling preview page")
	tmpl, err := template.ParseFiles("templates/preview.html")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse preview.html template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Debug().Str("dynamicLink", dynamicLink.String()).Msg("Executing template")

	err = tmpl.Execute(w, struct {
		DynamicLink       string
		AppIconImageURL   string
		AppName           string
		SocialTitle       string
		SocialDescription string
		SocialImageLink   string
	}{
		DynamicLink:       dynamicLink.String(),
		AppIconImageURL:   appIconImageURL,
		AppName:           appName,
		SocialTitle:       socialTitle,
		SocialDescription: socialDescription,
		SocialImageLink:   socialImageLink,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleiOSDynamicLink(w http.ResponseWriter, r *http.Request, queryParams url.Values, isiPad bool, isiPhone bool) {
	log.Debug().Msg("Handling iOS dynamic link")

	redirectIfValid := func(paramName string, deviceCheck bool) bool {
		link := queryParams.Get(paramName)
		if link == "" || !deviceCheck {
			return false
		}

		unescapedLink, err := url.QueryUnescape(link)
		if err != nil {
			log.Error().Str(paramName, link).Msgf("Failed to unescape '%s' parameter", paramName)
			http.Error(w, "Invalid '"+paramName+"' link format", http.StatusInternalServerError)
			return true
		}

		parsedURL, err := url.Parse(unescapedLink)
		if err != nil || !parsedURL.IsAbs() {
			log.Error().Str(paramName, unescapedLink).Msgf("Invalid '%s' link", paramName)
			http.Error(w, "Invalid '"+paramName+"' link format", http.StatusInternalServerError)
			return true
		}

		http.Redirect(w, r, unescapedLink, http.StatusFound)
		return true
	}

	if redirectIfValid("ipfl", isiPad) || redirectIfValid("ifl", isiPhone) {
		return
	}

	appStoreID := queryParams.Get("isi")
	if appStoreID != "" {
		redirectURL := fmt.Sprintf("https://apps.apple.com/app/id%s", appStoreID)
		params := []string{}
		for _, key := range []string{"at", "ct", "mt", "pt"} {
			if value := queryParams.Get(key); value != "" {
				params = append(params, key+"="+value)
			}
		}

		if len(params) > 0 {
			redirectURL += "?" + strings.Join(params, "&")
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}
}

func handleAndroidDynamicLink(w http.ResponseWriter, r *http.Request, queryParams url.Values, dynamicLink *url.URL) {
	log.Debug().Msg("Handling Android dynamic link")

	redirectIfValid := func(paramName string) bool {
		link := queryParams.Get(paramName)
		if link == "" {
			return false
		}
		unescapedLink, err := url.QueryUnescape(link)
		if err != nil {
			log.Error().Str(paramName, link).Msgf("Failed to unescape '%s' parameter", paramName)
			http.Error(w, "Invalid '"+paramName+"' link format", http.StatusInternalServerError)
			return true
		}

		parsedURL, err := url.Parse(unescapedLink)
		if err != nil || !parsedURL.IsAbs() {
			log.Error().Str(paramName, unescapedLink).Msgf("Invalid '%s' link", paramName)
			http.Error(w, "Invalid '"+paramName+"' link format", http.StatusInternalServerError)
			return true
		}

		http.Redirect(w, r, unescapedLink, http.StatusFound)
		return true
	}

	if redirectIfValid("afl") {
		return
	}

	appPackageName := queryParams.Get("apn")
	if appPackageName != "" {
		dynamicLink.RawQuery = ""
		redirectURL := fmt.Sprintf("https://play.google.com/store/apps/details?id=%s&referrer=tracking_id%%3D%s", appPackageName, dynamicLink)
		log.Debug().Str("redirectURL", redirectURL).Msg("Redirecting to Play Store")
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}
}

func (h *DynamicLinkHandler) AppleAppSiteAssociation(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Processing Apple App Site Association request")
	http.ServeFile(w, r, "static/apple-app-site-association.json")
}

func domainFallbackRedirect(cfg *config.Config) http.HandlerFunc {
	fallbackHost := strings.TrimSpace(cfg.FallbackHost)
	defaultScheme := "https"

	return func(w http.ResponseWriter, r *http.Request) {
 		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

        if r.Method != http.MethodGet {
            http.NotFound(w, r)
            return
        }

		target := url.URL{
			Scheme:   defaultScheme,
			Host:     fallbackHost,
			Path:     r.URL.EscapedPath(),
			RawQuery: r.URL.RawQuery,
		}

        log.Debug().
            Str("method", r.Method).
            Str("host", r.Host).
            Str("path", r.URL.Path).
            Str("query", r.URL.RawQuery).
            Str("target", target.String()).
            Msg("Fallback redirect")

		http.Redirect(w, r, target.String(), http.StatusPermanentRedirect)
	}
}