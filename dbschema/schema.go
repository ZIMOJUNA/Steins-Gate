package dbschema

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed schema.sql
var schemaSQL string

const (
	dropStart   = "-- steins-gate:drop:start"
	dropEnd     = "-- steins-gate:drop:end"
	createStart = "-- steins-gate:create:start"
	createEnd   = "-- steins-gate:create:end"
)

func Ensure(ctx context.Context, db *sql.DB) error {
	return execSection(ctx, db, createStart, createEnd)
}

func Reset(ctx context.Context, db *sql.DB) error {
	if err := execSection(ctx, db, dropStart, dropEnd); err != nil {
		return err
	}
	return Ensure(ctx, db)
}

func execSection(ctx context.Context, db *sql.DB, startMarker string, endMarker string) error {
	section, err := readSection(startMarker, endMarker)
	if err != nil {
		return err
	}

	for _, stmt := range splitStatements(section) {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec schema statement: %w", err)
		}
	}
	return nil
}

func readSection(startMarker string, endMarker string) (string, error) {
	start := strings.Index(schemaSQL, startMarker)
	if start < 0 {
		return "", fmt.Errorf("schema marker not found: %s", startMarker)
	}
	start += len(startMarker)

	end := strings.Index(schemaSQL[start:], endMarker)
	if end < 0 {
		return "", fmt.Errorf("schema marker not found: %s", endMarker)
	}

	return schemaSQL[start : start+end], nil
}

func splitStatements(sqlText string) []string {
	parts := strings.Split(sqlText, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		statements = append(statements, stmt)
	}
	return statements
}
