package config

import "testing"

func clearOIDCEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"JWT_SECRET", "ALLOW_REGISTRATION", "ALLOW_PASSWORD_LOGIN",
		"OIDC_ISSUER_URL", "OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET",
		"OIDC_REDIRECT_URL", "OIDC_PROVIDER_NAME", "API_BASE_URL",
	} {
		t.Setenv(k, "")
	}
}

func TestLoadAuthDefaults(t *testing.T) {
	clearOIDCEnv(t)

	cfg := LoadAuth()
	if !cfg.AllowPasswordLogin {
		t.Error("want AllowPasswordLogin true by default")
	}
	if cfg.OIDC != nil {
		t.Error("want OIDC nil when unset")
	}
}

func TestLoadAuthPasswordLoginDisabled(t *testing.T) {
	clearOIDCEnv(t)
	t.Setenv("ALLOW_PASSWORD_LOGIN", "false")

	cfg := LoadAuth()
	if cfg.AllowPasswordLogin {
		t.Error("want AllowPasswordLogin false")
	}
}

func TestLoadAuthOIDCIncomplete(t *testing.T) {
	clearOIDCEnv(t)
	t.Setenv("OIDC_ISSUER_URL", "https://idp.example.com")
	t.Setenv("OIDC_CLIENT_ID", "client-id")
	// client secret intentionally left unset

	cfg := LoadAuth()
	if cfg.OIDC != nil {
		t.Error("want OIDC nil when any of issuer/client id/secret is missing")
	}
}

func TestLoadAuthOIDCComplete(t *testing.T) {
	clearOIDCEnv(t)
	t.Setenv("OIDC_ISSUER_URL", "https://idp.example.com")
	t.Setenv("OIDC_CLIENT_ID", "client-id")
	t.Setenv("OIDC_CLIENT_SECRET", "client-secret")
	t.Setenv("API_BASE_URL", "https://api.example.com")

	cfg := LoadAuth()
	if cfg.OIDC == nil {
		t.Fatal("want OIDC set")
	}
	if cfg.OIDC.RedirectURL != "https://api.example.com/api/v1/auth/oidc/callback" {
		t.Errorf("want derived redirect url, got %q", cfg.OIDC.RedirectURL)
	}
	if cfg.OIDC.ProviderName != "SSO" {
		t.Errorf("want default provider name SSO, got %q", cfg.OIDC.ProviderName)
	}
}

func TestLoadAuthOIDCExplicitRedirectAndProviderName(t *testing.T) {
	clearOIDCEnv(t)
	t.Setenv("OIDC_ISSUER_URL", "https://idp.example.com")
	t.Setenv("OIDC_CLIENT_ID", "client-id")
	t.Setenv("OIDC_CLIENT_SECRET", "client-secret")
	t.Setenv("OIDC_REDIRECT_URL", "https://custom.example.com/callback")
	t.Setenv("OIDC_PROVIDER_NAME", "Authentik")

	cfg := LoadAuth()
	if cfg.OIDC == nil {
		t.Fatal("want OIDC set")
	}
	if cfg.OIDC.RedirectURL != "https://custom.example.com/callback" {
		t.Errorf("want explicit redirect url to win, got %q", cfg.OIDC.RedirectURL)
	}
	if cfg.OIDC.ProviderName != "Authentik" {
		t.Errorf("want provider name Authentik, got %q", cfg.OIDC.ProviderName)
	}
}
