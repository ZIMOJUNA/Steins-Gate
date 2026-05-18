package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Future-Game-Laboratory/Steins-Gate/mysql"
)

type PlayerData struct {
	ID        uint64          `json:"id"`
	AccountID uint64          `json:"account_id"`
	GameKey   string          `json:"game_key"`
	SlotKey   string          `json:"slot_key"`
	Data      json.RawMessage `json:"data"`
	Version   uint64          `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type PlayerDataSummary struct {
	ID        uint64    `json:"id"`
	GameKey   string    `json:"game_key"`
	SlotKey   string    `json:"slot_key"`
	Version   uint64    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PlayerDataService struct{}

func NewPlayerDataService() *PlayerDataService {
	return &PlayerDataService{}
}

func (s *PlayerDataService) List(ctx context.Context, accountID uint64, gameKey string) ([]PlayerDataSummary, error) {
	gameKey = strings.TrimSpace(gameKey)

	query := `
		SELECT id, game_key, slot_key, version, created_at, updated_at
		FROM player_data
		WHERE account_id = ?`
	args := []any{accountID}
	if gameKey != "" {
		query += " AND game_key = ?"
		args = append(args, gameKey)
	}
	query += " ORDER BY updated_at DESC, id DESC"

	rows, err := mysql.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PlayerDataSummary, 0)
	for rows.Next() {
		var item PlayerDataSummary
		if err := rows.Scan(&item.ID, &item.GameKey, &item.SlotKey, &item.Version, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (s *PlayerDataService) Get(ctx context.Context, accountID uint64, id uint64) (*PlayerData, error) {
	return getPlayerData(ctx, accountID, id)
}

func (s *PlayerDataService) Upsert(ctx context.Context, accountID uint64, gameKey string, slotKey string, data json.RawMessage) (*PlayerData, error) {
	gameKey, slotKey, err := normalizeDataKeys(gameKey, slotKey)
	if err != nil {
		return nil, err
	}
	if err := validateJSONData(data); err != nil {
		return nil, err
	}

	_, err = mysql.Exec(ctx, `
		INSERT INTO player_data (account_id, game_key, slot_key, data)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			data = VALUES(data),
			version = version + 1,
			updated_at = CURRENT_TIMESTAMP`,
		accountID, gameKey, slotKey, string(data),
	)
	if err != nil {
		return nil, err
	}

	var id uint64
	if err := mysql.QueryRow(ctx, `
		SELECT id
		FROM player_data
		WHERE account_id = ? AND game_key = ? AND slot_key = ?`,
		accountID, gameKey, slotKey,
	).Scan(&id); err != nil {
		return nil, err
	}

	return getPlayerData(ctx, accountID, id)
}

func (s *PlayerDataService) Update(ctx context.Context, accountID uint64, id uint64, data json.RawMessage) (*PlayerData, error) {
	if err := validateJSONData(data); err != nil {
		return nil, err
	}

	result, err := mysql.Exec(ctx, `
		UPDATE player_data
		SET data = ?, version = version + 1
		WHERE id = ? AND account_id = ?`,
		string(data), id, accountID,
	)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrSaveNotFound
	}

	return getPlayerData(ctx, accountID, id)
}

func (s *PlayerDataService) Delete(ctx context.Context, accountID uint64, id uint64) error {
	result, err := mysql.Exec(ctx, "DELETE FROM player_data WHERE id = ? AND account_id = ?", id, accountID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrSaveNotFound
	}
	return nil
}

func getPlayerData(ctx context.Context, accountID uint64, id uint64) (*PlayerData, error) {
	var item PlayerData
	var data []byte
	err := mysql.QueryRow(ctx, `
		SELECT id, account_id, game_key, slot_key, data, version, created_at, updated_at
		FROM player_data
		WHERE id = ? AND account_id = ?`,
		id, accountID,
	).Scan(&item.ID, &item.AccountID, &item.GameKey, &item.SlotKey, &data, &item.Version, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSaveNotFound
		}
		return nil, err
	}
	item.Data = json.RawMessage(data)
	return &item, nil
}

func normalizeDataKeys(gameKey string, slotKey string) (string, string, error) {
	gameKey = strings.TrimSpace(gameKey)
	slotKey = strings.TrimSpace(slotKey)
	if gameKey == "" {
		gameKey = "default"
	}
	if slotKey == "" {
		slotKey = "default"
	}
	if len(gameKey) > 64 || len(slotKey) > 64 {
		return "", "", ErrInvalidInput
	}
	return gameKey, slotKey, nil
}

func validateJSONData(data json.RawMessage) error {
	if len(data) == 0 || !json.Valid(data) {
		return ErrInvalidInput
	}
	return nil
}
