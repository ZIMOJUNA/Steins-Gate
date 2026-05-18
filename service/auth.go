package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/Future-Game-Laboratory/Steins-Gate/config"
	"github.com/Future-Game-Laboratory/Steins-Gate/mailer"
	"github.com/Future-Game-Laboratory/Steins-Gate/mysql"
	cache "github.com/Future-Game-Laboratory/Steins-Gate/redis"
	mysqldriver "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const (
	CodeSceneRegister      = "register"
	CodeSceneResetPassword = "reset_password"
	accountStatusActive    = 1
)

type AuthService struct {
	sender mailer.Sender
}

type Account struct {
	ID          uint64     `json:"id"`
	Email       string     `json:"email"`
	Nickname    string     `json:"nickname"`
	Status      int        `json:"status"`
	LoginCount  uint64     `json:"login_count"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type AuthResult struct {
	Account Account   `json:"account"`
	Token   AuthToken `json:"token"`
}

func NewAuthService(sender mailer.Sender) *AuthService {
	return &AuthService{sender: sender}
}

func (s *AuthService) SendEmailCode(ctx context.Context, email string, scene string) error {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return err
	}
	scene, err := normalizeScene(scene)
	if err != nil {
		return err
	}

	exists, err := accountExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	if scene == CodeSceneRegister && exists {
		return ErrEmailExists
	}
	if scene == CodeSceneResetPassword && !exists {
		return ErrAccountNotFound
	}

	if err := ensureEmailCodeRateLimit(ctx, email, scene); err != nil {
		return err
	}

	code, err := randomNumericCode(6)
	if err != nil {
		return err
	}

	ttl := config.Conf.Auth.EmailCodeDuration()
	hash := hashEmailCode(email, scene, code)
	codeKey := emailCodeKey(scene, email)
	failKey := emailCodeFailKey(scene, email)
	if err := cache.Set(ctx, codeKey, []byte(hash), ttl); err != nil {
		return err
	}
	_ = cache.Del(ctx, failKey)

	if err := s.sender.SendVerificationCode(ctx, email, scene, code, ttl); err != nil {
		_ = cache.Del(ctx, codeKey)
		_ = cache.Del(ctx, emailCodeCooldownKey(scene, email))
		return err
	}

	return nil
}

func (s *AuthService) Register(ctx context.Context, email string, password string, nickname string, code string) (*AuthResult, error) {
	email = normalizeEmail(email)
	nickname = strings.TrimSpace(nickname)

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	if err := verifyEmailCode(ctx, email, CodeSceneRegister, code); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := mysql.Exec(ctx, `
		INSERT INTO user_accounts (email, nickname, password_hash, status)
		VALUES (?, ?, ?, ?)`,
		email, nickname, string(passwordHash), accountStatusActive,
	)
	if err != nil {
		if isDuplicateEntry(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	accountID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	account, err := GetAccountByID(ctx, uint64(accountID))
	if err != nil {
		return nil, err
	}

	token, err := issueToken(uint64(accountID))
	if err != nil {
		return nil, err
	}

	return &AuthResult{Account: *account, Token: token}, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (*AuthResult, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	account, passwordHash, err := getAccountWithPasswordByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if account.Status != accountStatusActive {
		return nil, ErrAccountDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	if _, err := mysql.Exec(ctx, `
		UPDATE user_accounts
		SET last_login_at = ?, login_count = login_count + 1
		WHERE id = ?`,
		now, account.ID,
	); err != nil {
		return nil, err
	}
	account.LastLoginAt = &now
	account.LoginCount++

	token, err := issueToken(account.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Account: *account, Token: token}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email string, code string, newPassword string) (*AuthResult, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(newPassword); err != nil {
		return nil, err
	}
	if err := verifyEmailCode(ctx, email, CodeSceneResetPassword, code); err != nil {
		return nil, err
	}

	account, err := getAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	if account.Status != accountStatusActive {
		return nil, ErrAccountDisabled
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if _, err := mysql.Exec(ctx, "UPDATE user_accounts SET password_hash = ? WHERE id = ?", string(passwordHash), account.ID); err != nil {
		return nil, err
	}

	token, err := issueToken(account.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Account: *account, Token: token}, nil
}

func (s *AuthService) Logout(_ context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrUnauthorized
	}
	return cache.DeleteToken(token)
}

func AuthenticateToken(ctx context.Context, token string) (*Account, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrUnauthorized
	}

	accountIDText, err := cache.GetTokenBizID(token)
	if err != nil {
		return nil, ErrUnauthorized
	}

	accountID, err := strconv.ParseUint(accountIDText, 10, 64)
	if err != nil {
		return nil, ErrUnauthorized
	}

	account, err := GetAccountByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	if account.Status != accountStatusActive {
		return nil, ErrAccountDisabled
	}

	return account, nil
}

func GetAccountByID(ctx context.Context, accountID uint64) (*Account, error) {
	var account Account
	var lastLoginAt sql.NullTime
	err := mysql.QueryRow(ctx, `
		SELECT id, email, nickname, status, login_count, last_login_at, created_at, updated_at
		FROM user_accounts
		WHERE id = ?`,
		accountID,
	).Scan(
		&account.ID,
		&account.Email,
		&account.Nickname,
		&account.Status,
		&account.LoginCount,
		&lastLoginAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastLoginAt.Valid {
		account.LastLoginAt = &lastLoginAt.Time
	}
	return &account, nil
}

func issueToken(accountID uint64) (AuthToken, error) {
	ttl := config.Conf.Auth.TokenDuration()
	token, err := cache.GenerateToken(strconv.FormatUint(accountID, 10), ttl)
	if err != nil {
		return AuthToken{}, err
	}
	return AuthToken{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(ttl.Seconds()),
	}, nil
}

func verifyEmailCode(ctx context.Context, email string, scene string, inputCode string) error {
	scene, err := normalizeScene(scene)
	if err != nil {
		return err
	}
	inputCode = strings.TrimSpace(inputCode)
	if len(inputCode) != 6 {
		return ErrInvalidCode
	}

	codeKey := emailCodeKey(scene, email)
	failKey := emailCodeFailKey(scene, email)
	storedHash, err := cache.Get(ctx, codeKey)
	if err != nil {
		return err
	}
	if len(storedHash) == 0 {
		return ErrCodeExpired
	}

	failCount, err := cache.GetInt(ctx, failKey)
	if err != nil {
		return err
	}
	if failCount >= int64(config.Conf.Auth.MaxVerifyAttempts()) {
		_ = cache.Del(ctx, codeKey)
		_ = cache.Del(ctx, failKey)
		return ErrCodeTooManyAttempts
	}

	inputHash := hashEmailCode(email, scene, inputCode)
	if subtle.ConstantTimeCompare([]byte(inputHash), storedHash) != 1 {
		_, _ = cache.IncrWithExpire(ctx, failKey, config.Conf.Auth.EmailCodeDuration())
		return ErrInvalidCode
	}

	_ = cache.Del(ctx, codeKey)
	_ = cache.Del(ctx, failKey)
	return nil
}

func ensureEmailCodeRateLimit(ctx context.Context, email string, scene string) error {
	cooldownKey := emailCodeCooldownKey(scene, email)
	cooldown, err := cache.Get(ctx, cooldownKey)
	if err != nil {
		return err
	}
	if len(cooldown) > 0 {
		return ErrTooManyRequests
	}

	countKey := emailCodeSendCountKey(scene, email)
	count, err := cache.IncrWithExpire(ctx, countKey, time.Hour)
	if err != nil {
		return err
	}
	if count > int64(config.Conf.Auth.SendLimitPerHour()) {
		return ErrTooManyRequests
	}

	return cache.Set(ctx, cooldownKey, []byte("1"), config.Conf.Auth.ResendInterval())
}

func accountExistsByEmail(ctx context.Context, email string) (bool, error) {
	var id uint64
	err := mysql.QueryRow(ctx, "SELECT id FROM user_accounts WHERE email = ? LIMIT 1", email).Scan(&id)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func getAccountByEmail(ctx context.Context, email string) (*Account, error) {
	var account Account
	var lastLoginAt sql.NullTime
	err := mysql.QueryRow(ctx, `
		SELECT id, email, nickname, status, login_count, last_login_at, created_at, updated_at
		FROM user_accounts
		WHERE email = ?`,
		email,
	).Scan(
		&account.ID,
		&account.Email,
		&account.Nickname,
		&account.Status,
		&account.LoginCount,
		&lastLoginAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastLoginAt.Valid {
		account.LastLoginAt = &lastLoginAt.Time
	}
	return &account, nil
}

func getAccountWithPasswordByEmail(ctx context.Context, email string) (*Account, string, error) {
	var account Account
	var passwordHash string
	var lastLoginAt sql.NullTime
	err := mysql.QueryRow(ctx, `
		SELECT id, email, password_hash, nickname, status, login_count, last_login_at, created_at, updated_at
		FROM user_accounts
		WHERE email = ?`,
		email,
	).Scan(
		&account.ID,
		&account.Email,
		&passwordHash,
		&account.Nickname,
		&account.Status,
		&account.LoginCount,
		&lastLoginAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, "", err
	}
	if lastLoginAt.Valid {
		account.LastLoginAt = &lastLoginAt.Time
	}
	return &account, passwordHash, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateEmail(email string) error {
	if len(email) > 255 {
		return ErrInvalidInput
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidInput
	}
	if addr.Address != email {
		return ErrInvalidInput
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < config.Conf.Auth.MinPasswordLength() || len(password) > 72 {
		return ErrInvalidInput
	}
	return nil
}

func normalizeScene(scene string) (string, error) {
	scene = strings.ToLower(strings.TrimSpace(scene))
	switch scene {
	case CodeSceneRegister, CodeSceneResetPassword:
		return scene, nil
	default:
		return "", ErrInvalidInput
	}
}

func randomNumericCode(length int) (string, error) {
	var builder strings.Builder
	builder.Grow(length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteString(n.String())
	}
	return builder.String(), nil
}

func hashEmailCode(email string, scene string, code string) string {
	secret := config.Conf.Auth.EmailCodeHashSecret
	if secret == "" {
		secret = "steins-gate-email-code"
	}
	key := []byte(secret)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(scene))
	mac.Write([]byte(":"))
	mac.Write([]byte(email))
	mac.Write([]byte(":"))
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

func emailCodeKey(scene string, email string) string {
	return fmt.Sprintf("email_code:%s:%s", scene, email)
}

func emailCodeFailKey(scene string, email string) string {
	return fmt.Sprintf("email_code_fail:%s:%s", scene, email)
}

func emailCodeCooldownKey(scene string, email string) string {
	return fmt.Sprintf("email_code_cooldown:%s:%s", scene, email)
}

func emailCodeSendCountKey(scene string, email string) string {
	return fmt.Sprintf("email_code_send_count:%s:%s", scene, email)
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
