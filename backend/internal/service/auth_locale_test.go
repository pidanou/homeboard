package service

import "testing"

// Guards against the supported-locale allowlist drifting from the frontend's
// paraglide locales (web/project.inlang/settings.json) — a mismatch here
// causes PATCH /api/v1/profile/locale to silently reject a locale the UI
// already offers.
func TestSupportedLocalesMatchesFrontend(t *testing.T) {
	frontendLocales := []string{"en", "fr", "es", "de"}
	for _, l := range frontendLocales {
		if !supportedLocales[l] {
			t.Errorf("locale %q is offered by the frontend but rejected by SetLocale", l)
		}
	}
}
