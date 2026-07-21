package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/oauth2"
)

const (
	oidcFlowCookieName = "oidc_flow"
	oidcFlowCookiePath = "/api/v1/auth/oidc"
	oidcFlowTTL        = 10 * time.Minute
)

type OIDCHandler struct {
	oidc       *service.OIDCService
	handoff    *service.OIDCHandoffStore
	cookieKey  []byte
	appBaseURL string
	secure     bool
}

func NewOIDCHandler(oidcSvc *service.OIDCService, handoff *service.OIDCHandoffStore, jwtSecret, appBaseURL, redirectURL string) *OIDCHandler {
	return &OIDCHandler{
		oidc:       oidcSvc,
		handoff:    handoff,
		cookieKey:  deriveCookieKey(jwtSecret),
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
		secure:     strings.HasPrefix(redirectURL, "https://"),
	}
}

func deriveCookieKey(jwtSecret string) []byte {
	key := make([]byte, 32)
	// Distinct info string keeps this key independent of the raw JWT signing key,
	// even though both are derived from the same JWT_SECRET input.
	kdf := hkdf.New(sha256.New, []byte(jwtSecret), nil, []byte("homeboard-oidc-flow-cookie"))
	io.ReadFull(kdf, key) //nolint:errcheck // hkdf.Read from a fixed-size hash never fails
	return key
}

func (h *OIDCHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/login", h.login)
	r.Get("/callback", h.callback)
	r.Post("/exchange", h.exchange)
	return r
}

type flowCookiePayload struct {
	State    string `json:"state"`
	Nonce    string `json:"nonce"`
	Verifier string `json:"verifier"`
	Expires  int64  `json:"expires"`
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (h *OIDCHandler) signCookie(payload flowCookiePayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, h.cookieKey)
	mac.Write(raw)
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

var errFlowCookieInvalid = errors.New("oidc: flow cookie missing, tampered, or expired")

func (h *OIDCHandler) verifyCookie(cookieValue string) (flowCookiePayload, error) {
	var payload flowCookiePayload
	parts := strings.SplitN(cookieValue, ".", 2)
	if len(parts) != 2 {
		return payload, errFlowCookieInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return payload, errFlowCookieInvalid
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return payload, errFlowCookieInvalid
	}
	mac := hmac.New(sha256.New, h.cookieKey)
	mac.Write(raw)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return payload, errFlowCookieInvalid
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, errFlowCookieInvalid
	}
	if time.Now().Unix() > payload.Expires {
		return payload, errFlowCookieInvalid
	}
	return payload, nil
}

func (h *OIDCHandler) clearFlowCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oidcFlowCookieName,
		Value:    "",
		Path:     oidcFlowCookiePath,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (h *OIDCHandler) login(w http.ResponseWriter, r *http.Request) {
	state, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "oidc_start_failed")
		return
	}
	nonce, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "oidc_start_failed")
		return
	}
	verifier := oauth2.GenerateVerifier()

	cookieValue, err := h.signCookie(flowCookiePayload{
		State:    state,
		Nonce:    nonce,
		Verifier: verifier,
		Expires:  time.Now().Add(oidcFlowTTL).Unix(),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "oidc_start_failed")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oidcFlowCookieName,
		Value:    cookieValue,
		Path:     oidcFlowCookiePath,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(oidcFlowTTL.Seconds()),
	})

	http.Redirect(w, r, h.oidc.AuthCodeURL(state, nonce, verifier), http.StatusFound)
}

func (h *OIDCHandler) callback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(oidcFlowCookieName)
	if err != nil {
		writeError(w, http.StatusBadRequest, "oidc_flow_expired")
		return
	}
	payload, err := h.verifyCookie(cookie.Value)
	h.clearFlowCookie(w)
	if err != nil {
		writeError(w, http.StatusBadRequest, "oidc_flow_expired")
		return
	}

	// CSRF: state must match what we handed the IdP (double-submit cookie pattern).
	if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("state")), []byte(payload.State)) != 1 {
		writeError(w, http.StatusBadRequest, "oidc_invalid_state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "oidc_missing_code")
		return
	}

	idToken, err := h.oidc.Exchange(r.Context(), code, payload.Verifier)
	if err != nil {
		writeError(w, http.StatusBadGateway, "oidc_exchange_failed")
		return
	}

	// Replay defense: go-oidc verifies signature/issuer/audience/expiry but does
	// NOT check the nonce for us — that's on the caller.
	if subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(payload.Nonce)) != 1 {
		writeError(w, http.StatusBadRequest, "oidc_invalid_nonce")
		return
	}

	var rawClaims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&rawClaims); err != nil {
		writeError(w, http.StatusBadGateway, "oidc_invalid_claims")
		return
	}

	user, err := h.oidc.FindOrCreateUser(r.Context(), service.OIDCClaims{
		Issuer:        idToken.Issuer,
		Subject:       idToken.Subject,
		Email:         rawClaims.Email,
		EmailVerified: rawClaims.EmailVerified,
		Name:          rawClaims.Name,
	})
	if err != nil {
		h.redirectWithError(w, r, err)
		return
	}

	token, err := h.oidc.IssueToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issue_token_failed")
		return
	}

	handoffCode, err := h.handoff.Put(token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "complete_login_failed")
		return
	}

	http.Redirect(w, r, h.appBaseURL+"/callback?code="+handoffCode, http.StatusFound)
}

// redirectWithError sends the user back to the frontend's callback page with a
// generic error code instead of a raw error message, and never with a caller-
// supplied redirect target — appBaseURL is fixed server-side config.
func (h *OIDCHandler) redirectWithError(w http.ResponseWriter, r *http.Request, err error) {
	reason := "oidc_failed"
	if errors.Is(err, service.ErrOIDCEmailNotVerified) {
		reason = "email_not_verified"
	} else if errors.Is(err, service.ErrRegistrationClosed) {
		reason = "registration_closed"
	}
	http.Redirect(w, r, h.appBaseURL+"/callback?error="+reason, http.StatusFound)
}

func (h *OIDCHandler) exchange(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	token, err := h.handoff.Consume(body.Code)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_or_expired_code")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
