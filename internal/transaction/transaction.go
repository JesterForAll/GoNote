package transaction

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/JesterForAll/gonote/internal/database"
)

type MultiFunc func(ctx context.Context, txs ...*gorm.DB) error

type MultiConfig struct {
	Name   string
	Logger *slog.Logger
	DBs    []*database.Database
}

func RunMulti(ctx context.Context, cfg MultiConfig, fn MultiFunc) error {
	var txs []*gorm.DB

	for i, db := range cfg.DBs {
		tx := db.WithContext(ctx).Begin()

		err := tx.Error
		if err != nil {
			rollbackAll(txs)

			cfg.Logger.Error(fmt.Sprintf("transaction %s failed to begin on db %d", cfg.Name, i), slog.Any("err", err))

			return fmt.Errorf("failed to begin transaction on db %d: %w", i, err)
		}

		txs = append(txs, tx)
	}

	if err := fn(ctx, txs...); err != nil {
		rollbackAll(txs)

		cfg.Logger.Error(fmt.Sprintf("transaction %s failed", cfg.Name), slog.Any("err", err))

		return err
	}

	for i, tx := range txs {
		if err := tx.Commit().Error; err != nil {
			rollbackAll(txs)

			cfg.Logger.Error(fmt.Sprintf("transaction %s failed to commit on db %d", cfg.Name, i), slog.Any("err", err))

			return fmt.Errorf("failed to commit on db %d: %w", i, err)
		}
	}

	return nil
}

func rollbackAll(txs []*gorm.DB) {
	for _, tx := range txs {
		_ = tx.Rollback()
	}
}

func MultiWithConfig(name string, logger *slog.Logger, dbs ...*database.Database) MultiConfig {
	return MultiConfig{
		Name:   name,
		Logger: logger,
		DBs:    dbs,
	}
}
