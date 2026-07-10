package config

import (
	"os"
	"strings"
)

type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	ProviderName string
}

type AuthConfig struct {
	JWTSecret          string
	AllowRegistration  bool
	AllowPasswordLogin bool
	OIDC               *OIDCConfig // nil if OIDC env vars are not fully set
}

// LoadAuth reads auth-related configuration from the environment.
// OIDC is considered "configured" (non-nil) only when issuer/client id/secret
// are all set — whether it actually becomes usable also depends on discovery
// succeeding at startup, which is checked separately.
func LoadAuth() AuthConfig {
	cfg := AuthConfig{
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AllowRegistration:  os.Getenv("ALLOW_REGISTRATION") == "true",
		AllowPasswordLogin: os.Getenv("ALLOW_PASSWORD_LOGIN") != "false",
	}

	issuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER_URL"))
	clientID := strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID"))
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	if issuer == "" || clientID == "" || clientSecret == "" {
		return cfg
	}

	redirectURL := strings.TrimSpace(os.Getenv("OIDC_REDIRECT_URL"))
	if redirectURL == "" {
		if apiBase := strings.TrimRight(os.Getenv("API_BASE_URL"), "/"); apiBase != "" {
			redirectURL = apiBase + "/api/v1/auth/oidc/callback"
		}
	}

	providerName := strings.TrimSpace(os.Getenv("OIDC_PROVIDER_NAME"))
	if providerName == "" {
		providerName = "SSO"
	}

	cfg.OIDC = &OIDCConfig{
		IssuerURL:    issuer,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		ProviderName: providerName,
	}
	return cfg
}
