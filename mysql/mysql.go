package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Future-Game-Laboratory/Steins-Gate/config"
	"github.com/Future-Game-Laboratory/Steins-Gate/dbschema"
	mysqldriver "github.com/go-sql-driver/mysql"
)

var db *sql.DB

var ErrNotInitialized = errors.New("mysql not initialized")

func Init() error {
	if db != nil {
		return nil
	}

	cfg := config.Conf.MySQL
	conn, err := sql.Open("mysql", buildDSN(cfg))
	if err != nil {
		return fmt.Errorf("MySQL 初始化失败：%w", err)
	}

	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return fmt.Errorf("MySQL 连接失败：%w", err)
	}

	if err := dbschema.Ensure(ctx, conn); err != nil {
		_ = conn.Close()
		return fmt.Errorf("MySQL 表结构初始化失败：%w", err)
	}

	db = conn
	log.Println("MySQL 全局初始化完成！")
	return nil
}

func buildDSN(cfg config.MySQLConfig) string {
	if cfg.DSN != "" {
		return cfg.DSN
	}

	driverCfg := mysqldriver.NewConfig()
	driverCfg.User = cfg.User
	driverCfg.Passwd = cfg.Password
	driverCfg.Net = "tcp"
	driverCfg.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	driverCfg.DBName = cfg.DBName
	driverCfg.ParseTime = true
	driverCfg.Loc = time.Local
	driverCfg.Collation = "utf8mb4_unicode_ci"

	return driverCfg.FormatDSN()
}

// GetDB 获取全局 MySQL 连接池。
func GetDB() *sql.DB {
	return db
}

// Close 关闭 MySQL 连接池。
func Close() error {
	if db == nil {
		return nil
	}
	return db.Close()
}

// Exec 执行 insert/update/delete 等写操作。
func Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.ExecContext(ctx, query, args...)
}

// Query 执行多行查询，调用方负责关闭返回的 rows。
func Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.QueryContext(ctx, query, args...)
}

// QueryRow 执行单行查询；调用前需要确保 Init 已成功执行。
func QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return db.QueryRowContext(ctx, query, args...)
}
