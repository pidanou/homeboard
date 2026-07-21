package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pidanou/homeboard/internal/service"
)

// writeError writes a stable machine-readable error code as JSON instead of
// a raw English string, so the frontend can localize it (docs/specs/i18n.md).
func writeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": code})
}

// errorCodes maps sentinel errors from the service layer to the stable code
// returned to the frontend. Populated alongside each service's sentinel vars.
var errorCodes = map[error]string{
	ErrUnauthorized:  "unauthorized",
	ErrAdminRequired: "admin_required",
	ErrForbidden:     "forbidden",

	service.ErrRegistrationClosed:     "registration_closed",
	service.ErrPasswordLoginDisabled:  "password_login_disabled",
	service.ErrInvalidResetToken:      "invalid_reset_link",
	service.ErrNoPasswordSet:          "no_password_set",
	service.ErrInvalidCurrentPassword: "invalid_current_password",

	service.ErrHouseholdAdminRequired: "admin_required",
	service.ErrTargetNotMember:        "target_not_member",
	service.ErrInvalidRole:            "invalid_role",
	service.ErrCannotChangeOwnRole:    "cannot_change_own_role",
	service.ErrCannotDemoteLastAdmin:  "cannot_demote_last_admin",
	service.ErrCannotRemoveSelf:       "cannot_remove_self",

	service.ErrInviteNotFound: "invite_not_found",
	service.ErrInviteNoEmail:  "invite_no_email",
	service.ErrInviteUsed:     "invite_already_used",
	service.ErrInviteExpired:  "invite_expired",
}

// codeFor resolves a sentinel error to its stable code, or fallback if err
// doesn't match any known sentinel (e.g. an unexpected/wrapped internal error,
// which should never leak its raw text to the client).
func codeFor(err error, fallback string) string {
	for sentinel, code := range errorCodes {
		if errors.Is(err, sentinel) {
			return code
		}
	}
	return fallback
}
