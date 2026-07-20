package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users              repository.UserRepository
	jwtSecret          []byte
	mailer             *EmailService
	allowPasswordLogin bool
}

func NewAuthService(users repository.UserRepository, jwtSecret string, mailer *EmailService) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret), mailer: mailer, allowPasswordLogin: true}
}

// SetAllowPasswordLogin toggles whether password-based register/login are accepted.
// Defaults to true; set to false to run OIDC-only.
func (s *AuthService) SetAllowPasswordLogin(allow bool) {
	s.allowPasswordLogin = allow
}

var ErrRegistrationClosed = errors.New("registration is closed")
var ErrPasswordLoginDisabled = errors.New("password login is disabled")

// CreateUser creates an account without checking the registration lock.
// Used by the invite flow to allow new users to join via an invite link.
func (s *AuthService) CreateUser(ctx context.Context, email, password, name string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	hashStr := string(hash)
	user := &model.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: &hashStr,
		Name:         name,
		Locale:       "en",
		TimeFormat:   "auto",
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	s.mailer.Send(email, emailData{
		Subject: "Welcome to Family Board",
		Heading: "Welcome aboard!",
		Name:    name,
		Body:    "Your Family Board account has been created. Open the app to get started.",
	})
	return user, nil
}

// IssueToken signs a JWT for the given user ID.
func (s *AuthService) IssueToken(userID string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return t.SignedString(s.jwtSecret)
}

// CheckRegistrationOpen applies the ALLOW_REGISTRATION gate (open, or locked
// after the first user exists — the self-host single-tenant bootstrap
// pattern). Shared by password registration and OIDC first-login signup so
// both respect the same policy.
func (s *AuthService) CheckRegistrationOpen(ctx context.Context) error {
	if os.Getenv("ALLOW_REGISTRATION") == "true" {
		return nil
	}
	exists, err := s.users.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check users: %w", err)
	}
	if exists {
		return ErrRegistrationClosed
	}
	return nil
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*model.User, error) {
	if !s.allowPasswordLogin {
		return nil, ErrPasswordLoginDisabled
	}
	if err := s.CheckRegistrationOpen(ctx); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	hashStr := string(hash)
	user := &model.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: &hashStr,
		Name:         name,
		Locale:       "en",
		TimeFormat:   "auto",
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	s.mailer.Send(email, emailData{
		Subject: "Welcome to Family Board",
		Heading: "Welcome aboard!",
		Name:    name,
		Body:    "Your Family Board account has been created. Open the app to get started.",
	})
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	if !s.allowPasswordLogin {
		return "", ErrPasswordLoginDisabled
	}
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if user.PasswordHash == nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	signed, err := s.IssueToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	s.mailer.Send(email, emailData{
		Subject: "New login to your Family Board account",
		Heading: "New login detected",
		Name:    user.Name,
		Body:    fmt.Sprintf("A new login was detected on your account at %s UTC.\n\nIf this wasn't you, change your password immediately.", time.Now().UTC().Format("2006-01-02 15:04:05")),
	})
	return signed, nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *AuthService) UpdateName(ctx context.Context, userID, name string) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Name = name
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.PasswordHash == nil {
		return errors.New("account has no password set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("invalid current password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	hashStr := string(hash)
	user.PasswordHash = &hashStr
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}
	s.mailer.Send(user.Email, emailData{
		Subject: "Your Family Board password was changed",
		Heading: "Password changed",
		Name:    user.Name,
		Body:    fmt.Sprintf("Your password was changed at %s UTC.\n\nIf this wasn't you, contact your administrator immediately.", time.Now().UTC().Format("2006-01-02 15:04:05")),
	})
	return nil
}

func (s *AuthService) SetAvatar(ctx context.Context, userID string, avatarURL *string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.AvatarURL = avatarURL
	return s.users.Update(ctx, user)
}

var supportedLocales = map[string]bool{"en": true, "fr": true, "es": true, "de": true}

func (s *AuthService) SetLocale(ctx context.Context, userID, locale string) (*model.User, error) {
	if !supportedLocales[locale] {
		return nil, errors.New("unsupported locale")
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Locale = locale
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

var supportedTimeFormats = map[string]bool{"auto": true, "12": true, "24": true}

func (s *AuthService) SetTimeFormat(ctx context.Context, userID, format string) (*model.User, error) {
	if !supportedTimeFormats[format] {
		return nil, errors.New("unsupported time format")
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.TimeFormat = format
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

var supportedReminderMinutes = map[int]bool{15: true, 30: true, 60: true}

func (s *AuthService) SetReminderMinutesBefore(ctx context.Context, userID string, minutes *int) (*model.User, error) {
	if minutes != nil && !supportedReminderMinutes[*minutes] {
		return nil, errors.New("unsupported reminder offset")
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.ReminderMinutesBefore = minutes
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}
