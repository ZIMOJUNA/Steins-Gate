package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type client struct {
	baseURL string
	http    *http.Client
}

type apiResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type authData struct {
	Account accountData `json:"account"`
	Token   tokenData   `json:"token"`
}

type accountData struct {
	ID       uint64 `json:"id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type tokenData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type playerData struct {
	ID      uint64          `json:"id"`
	GameKey string          `json:"game_key"`
	SlotKey string          `json:"slot_key"`
	Data    json.RawMessage `json:"data"`
	Version uint64          `json:"version"`
}

type listData struct {
	Items []playerData `json:"items"`
}

func main() {
	baseURL := flag.String("base-url", "http://127.0.0.1:8080", "server base URL")
	email := flag.String("email", "3180615598@qq.com", "test account email")
	password := flag.String("password", "password123", "test account password")
	nickname := flag.String("nickname", "api-test", "test account nickname")
	code := flag.String("register-code", "", "register verification code; prompt if empty")
	skipRegister := flag.Bool("skip-register", false, "skip register and login directly")
	flag.Parse()

	testEmail := strings.TrimSpace(*email)
	if testEmail == "" {
		testEmail = promptRequired("enter test account email: ")
	}
	if !strings.Contains(testEmail, "@") {
		log.Fatalf("invalid email: %s", testEmail)
	}

	testNickname := strings.TrimSpace(*nickname)
	if testNickname == "" {
		testNickname = "api-test"
	}

	c := client{
		baseURL: strings.TrimRight(*baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}

	ctx := context.Background()
	must(c.health(ctx))

	var token string
	if !*skipRegister {
		registered, registerToken := runRegister(ctx, c, testEmail, *password, testNickname, *code)
		if registered {
			token = registerToken
		}
	}

	if token == "" {
		token = runLogin(ctx, c, testEmail, *password)
	}

	must(c.me(ctx, token))
	saveID := mustValue(c.upsertSave(ctx, token))
	must(c.listSaves(ctx, token))
	must(c.getSave(ctx, token, saveID))
	must(c.updateSave(ctx, token, saveID))
	must(c.deleteSave(ctx, token, saveID))
	must(c.logout(ctx, token))

	log.Println("api test completed")
}

func runRegister(ctx context.Context, c client, email string, password string, nickname string, code string) (bool, string) {
	status, resp, err := c.post(ctx, "/api/v1/auth/email-code", "", map[string]any{
		"email": email,
		"scene": "register",
	})
	must(err)
	if status == http.StatusConflict && resp.Code == "email_exists" {
		log.Println("account already exists, skip register")
		return false, ""
	}
	mustStatus(status, resp, http.StatusOK)
	log.Printf("verification code sent to %s", email)
	log.Println("if mail.provider=console, check the backend server log for: [mail:console] ... code=xxxxxx")

	if code == "" {
		code = promptCode("enter 6-digit register verification code: ")
	} else if !isSixDigitCode(code) {
		log.Fatalf("register-code must be a 6-digit code, got: %s", code)
	}

	status, resp, err = c.post(ctx, "/api/v1/auth/register", "", map[string]any{
		"email":    email,
		"password": password,
		"nickname": nickname,
		"code":     code,
	})
	must(err)
	if status == http.StatusConflict && resp.Code == "email_exists" {
		log.Println("account already exists, skip register")
		return false, ""
	}
	mustStatus(status, resp, http.StatusOK)

	var auth authData
	must(json.Unmarshal(resp.Data, &auth))
	log.Printf("registered account id=%d email=%s", auth.Account.ID, auth.Account.Email)
	return true, auth.Token.AccessToken
}

func runLogin(ctx context.Context, c client, email string, password string) string {
	status, resp, err := c.post(ctx, "/api/v1/auth/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	must(err)
	mustStatus(status, resp, http.StatusOK)

	var auth authData
	must(json.Unmarshal(resp.Data, &auth))
	log.Printf("logged in account id=%d email=%s", auth.Account.ID, auth.Account.Email)
	return auth.Token.AccessToken
}

func (c client) health(ctx context.Context) error {
	status, resp, err := c.get(ctx, "/health", "")
	if err != nil {
		return err
	}
	mustStatus(status, resp, http.StatusOK)
	log.Println("health ok")
	return nil
}

func (c client) me(ctx context.Context, token string) error {
	status, resp, err := c.get(ctx, "/api/v1/me", token)
	if err != nil {
		return err
	}
	mustStatus(status, resp, http.StatusOK)
	log.Println("me ok")
	return nil
}

func (c client) upsertSave(ctx context.Context, token string) (uint64, error) {
	status, resp, err := c.post(ctx, "/api/v1/player-data", token, map[string]any{
		"game_key": "apitest-game",
		"slot_key": "slot-1",
		"data": map[string]any{
			"chapter":   1,
			"play_time": 120,
			"items":     []string{"phone", "badge"},
		},
	})
	if err != nil {
		return 0, err
	}
	mustStatus(status, resp, http.StatusOK)

	var save playerData
	if err := json.Unmarshal(resp.Data, &save); err != nil {
		return 0, err
	}
	log.Printf("upsert save ok id=%d version=%d", save.ID, save.Version)
	return save.ID, nil
}

func (c client) listSaves(ctx context.Context, token string) error {
	status, resp, err := c.get(ctx, "/api/v1/player-data?game_key=apitest-game", token)
	if err != nil {
		return err
	}
	mustStatus(status, resp, http.StatusOK)

	var list listData
	if err := json.Unmarshal(resp.Data, &list); err != nil {
		return err
	}
	log.Printf("list saves ok count=%d", len(list.Items))
	return nil
}

func (c client) getSave(ctx context.Context, token string, id uint64) error {
	status, resp, err := c.get(ctx, fmt.Sprintf("/api/v1/player-data/%d", id), token)
	if err != nil {
		return err
	}
	mustStatus(status, resp, http.StatusOK)
	log.Println("get save ok")
	return nil
}

func (c client) updateSave(ctx context.Context, token string, id uint64) error {
	status, resp, err := c.put(ctx, fmt.Sprintf("/api/v1/player-data/%d", id), token, map[string]any{
		"data": map[string]any{
			"chapter":   2,
			"play_time": 360,
			"items":     []string{"phone", "badge", "lab-note"},
		},
	})
	if err != nil {
		return err
	}
	mustStatus(status, resp, http.StatusOK)
	log.Println("update save ok")
	return nil
}

func (c client) deleteSave(ctx context.Context, token string, id uint64) error {
	status, resp, err := c.delete(ctx, fmt.Sprintf("/api/v1/player-data/%d", id), token)
	if err != nil {
		return err
	}
	mustStatus(status, resp, http.StatusNoContent)
	log.Println("delete save ok")
	return nil
}

func (c client) logout(ctx context.Context, token string) error {
	status, resp, err := c.post(ctx, "/api/v1/auth/logout", token, nil)
	if err != nil {
		return err
	}
	mustStatus(status, resp, http.StatusNoContent)
	log.Println("logout ok")
	return nil
}

func (c client) get(ctx context.Context, path string, token string) (int, apiResponse, error) {
	return c.request(ctx, http.MethodGet, path, token, nil)
}

func (c client) post(ctx context.Context, path string, token string, body any) (int, apiResponse, error) {
	return c.request(ctx, http.MethodPost, path, token, body)
}

func (c client) put(ctx context.Context, path string, token string, body any) (int, apiResponse, error) {
	return c.request(ctx, http.MethodPut, path, token, body)
}

func (c client) delete(ctx context.Context, path string, token string) (int, apiResponse, error) {
	return c.request(ctx, http.MethodDelete, path, token, nil)
}

func (c client) request(ctx context.Context, method string, path string, token string, body any) (int, apiResponse, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, apiResponse{}, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, apiResponse{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return 0, apiResponse{}, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, apiResponse{}, err
	}
	if res.StatusCode == http.StatusNoContent {
		return res.StatusCode, apiResponse{Code: "ok", Message: "ok"}, nil
	}
	if len(raw) == 0 {
		return res.StatusCode, apiResponse{}, nil
	}

	var resp apiResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return res.StatusCode, apiResponse{}, fmt.Errorf("decode response: %w body=%s", err, string(raw))
	}
	return res.StatusCode, resp, nil
}

func mustStatus(status int, resp apiResponse, wants ...int) {
	for _, want := range wants {
		if status == want {
			return
		}
	}
	log.Fatalf("unexpected status=%d code=%s message=%s", status, resp.Code, resp.Message)
}

func prompt(label string) string {
	fmt.Print(label)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(line)
}

func promptRequired(label string) string {
	for {
		value := prompt(label)
		if value != "" {
			return value
		}
	}
}

func promptCode(label string) string {
	for {
		value := prompt(label)
		if isSixDigitCode(value) {
			return value
		}
		log.Println("please enter the 6-digit code, not the email address")
	}
}

func isSixDigitCode(value string) bool {
	if len(value) != 6 {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustValue[T any](value T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return value
}
