package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/proyaai/instantgate/internal/config"
	"github.com/proyaai/instantgate/internal/database"

	_ "github.com/go-sql-driver/mysql"
)

type Driver struct {
	config *config.DatabaseConfig
}

func NewDriver(cfg *config.DatabaseConfig) *Driver {
	return &Driver{
		config: cfg,
	}
}

func (d *Driver) Connect(ctx context.Context) (*sql.DB, error) {
	dsn := d.config.DSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	db.SetMaxOpenConns(d.config.MaxOpenConns)
	db.SetMaxIdleConns(d.config.MaxIdleConns)
	db.SetConnMaxLifetime(d.config.ConnMaxLifetime)

	return db, nil
}

func (d *Driver) Ping(ctx context.Context, db *sql.DB) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

func (d *Driver) GetTables(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := db.QueryContext(ctx, query, d.config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	return tables, nil
}

func (d *Driver) GetColumns(ctx context.Context, db *sql.DB, table string) ([]database.ColumnInfo, error) {
	query := `
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY,
			EXTRA,
			CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.QueryContext(ctx, query, d.config.Name, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for table %s: %w", table, err)
	}
	defer rows.Close()

	var columns []database.ColumnInfo
	for rows.Next() {
		var col database.ColumnInfo
		var nullable string
		var columnKey sql.NullString
		var extra sql.NullString

		if err := rows.Scan(
			&col.Name,
			&col.Type,
			&nullable,
			&col.DefaultValue,
			&columnKey,
			&extra,
			&col.MaxLength,
		); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col.Nullable = IsColumnNullable(nullable)
		col.IsPrimaryKey = columnKey.Valid && columnKey.String == "PRI"
		col.IsAutoIncrement = extra.Valid && IsAutoIncrement(extra.String)

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	return columns, nil
}

func (d *Driver) GetPrimaryKey(ctx context.Context, db *sql.DB, table string) (string, error) {
	query := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
		AND COLUMN_KEY = 'PRI'
		ORDER BY ORDINAL_POSITION
		LIMIT 1
	`

	var pkColumn string
	err := db.QueryRowContext(ctx, query, d.config.Name, table).Scan(&pkColumn)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get primary key for table %s: %w", table, err)
	}

	return pkColumn, nil
}

func (d *Driver) GetRelationships(ctx context.Context, db *sql.DB, table string) ([]database.RelationshipInfo, error) {
	query := `
		SELECT
			kcu.COLUMN_NAME,
			kcu.REFERENCED_TABLE_NAME,
			kcu.REFERENCED_COLUMN_NAME,
			kcu.CONSTRAINT_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
		WHERE kcu.TABLE_SCHEMA = ?
		AND kcu.TABLE_NAME = ?
		AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY kcu.ORDINAL_POSITION
	`

	rows, err := db.QueryContext(ctx, query, d.config.Name, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships for table %s: %w", table, err)
	}
	defer rows.Close()

	var relationships []database.RelationshipInfo
	for rows.Next() {
		var rel database.RelationshipInfo
		if err := rows.Scan(
			&rel.ColumnName,
			&rel.ReferencedTable,
			&rel.ReferencedColumn,
			&rel.ConstraintName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		relationships = append(relationships, rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating relationships: %w", err)
	}

	return relationships, nil
}
