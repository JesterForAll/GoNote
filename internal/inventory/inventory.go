package inventory

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/JesterForAll/gonote/internal/balance"
	"github.com/JesterForAll/gonote/internal/database"
	"github.com/JesterForAll/gonote/internal/transaction"
	"gorm.io/gorm"
)

type Inventory struct {
	logger  *slog.Logger
	Db      *database.Database
	balance *balance.Balance
}

type HelpOctaveNumber struct {
	Ok     bool
	Octave string
}

type HelpNotePosition struct {
	Ok  bool
	Pos string
}

const costOfFailSafe int = 10
const costOfHelpers int = 15

func newInventory(logger *slog.Logger, balance *balance.Balance) (*Inventory, error) {
	db, err := database.New("../../static/inventory.db", logger, database.InventoryDbStruct{})
	if err != nil {
		logger.Error("failed to connect database", slog.Any("err", err))

		return nil, err
	}

	return &Inventory{logger: logger, Db: db, balance: balance}, nil
}

func (inv *Inventory) GetCurrentNumOfSafeFails(userID int) int {

	var invDB database.InventoryDbStruct

	exist := inv.Db.CheckIfExistAndGetFirst(map[string]interface{}{"user_id": userID}, &invDB)

	if !exist {
		return 0
	}

	return invDB.NumOfSafeFails
}

func (inv *Inventory) UpdateCurrentNumOfSafeFails(userID int, buy bool) (int, error) {

	var invDB database.InventoryDbStruct

	exist := inv.Db.CheckIfExistAndGetFirst(map[string]interface{}{"user_id": userID}, &invDB)

	if !exist {
		invDB.UserID = userID
	}

	oldNumOfSafeFails := invDB.NumOfSafeFails

	if buy {

		currBalance := inv.balance.GetCurrentBalance(userID)

		if currBalance < costOfFailSafe {
			inv.logger.Info("insufficient balance to buy fail safe")

			return oldNumOfSafeFails, nil
		}

		err := inv.balance.UpdateCurrentBalance(userID, -costOfFailSafe)

		if err != nil {
			inv.logger.Error("failed to deduct balance", slog.Any("err", err))

			return oldNumOfSafeFails, err
		}
	}

	val := 1

	if !buy {
		val *= -1

		if oldNumOfSafeFails == 0 {
			return oldNumOfSafeFails, nil
		}
	}

	invDB.NumOfSafeFails += val

	err := inv.Db.Upsert(&invDB)
	if err != nil {
		inv.logger.Error("error updating number of fail saves", slog.Any("err", err))

		return oldNumOfSafeFails, err
	}

	return invDB.NumOfSafeFails, nil
}

func (inv *Inventory) UpdateCurrentNumOfSafeFailsWithTx(ctx context.Context, userID int, buy bool) (int, error) {

	var invDB database.InventoryDbStruct

	exist := inv.Db.CheckIfExistAndGetFirst(map[string]interface{}{"user_id": userID}, &invDB)

	if !exist {
		invDB.UserID = userID
	}

	oldNumOfSafeFails := invDB.NumOfSafeFails

	if buy {

		currBalance := inv.balance.GetCurrentBalance(userID)

		if currBalance < costOfFailSafe {
			inv.logger.Info("insufficient balance to buy fail safe")

			return oldNumOfSafeFails, nil
		}

	}

	val := 1

	if !buy {
		val *= -1

		if oldNumOfSafeFails == 0 {
			return oldNumOfSafeFails, nil
		}

		invDB.NumOfSafeFails += val

		err := inv.Db.Upsert(&invDB)
		if err != nil {
			inv.logger.Error("error updating number of fail saves", slog.Any("err", err))

			return oldNumOfSafeFails, err
		}
	}

	invDB.NumOfSafeFails += val

	//running a transaction
	if buy {
		err := transaction.RunMulti(ctx, transaction.MultiConfig{
			Name:   "buying safe fail transaction",
			Logger: inv.logger,
			DBs:    []*database.Database{inv.Db, inv.balance.Db}},
			func(ctx context.Context, txs ...*gorm.DB) error {

				if err := inv.Db.UpsertWithTx(txs[0], &invDB); err != nil {
					return fmt.Errorf("error updating inventory: %w", err)
				}

				if err := inv.balance.UpdateCurrentBalanceWithTx(txs[1], userID, -costOfFailSafe); err != nil {
					return fmt.Errorf("error updating balance: %w", err)
				}

				return nil
			})

		if err != nil {
			inv.logger.Error("error during buying fail safe transaction", slog.Any("err", err))

			return oldNumOfSafeFails, err
		}
	}

	return invDB.NumOfSafeFails, nil
}

func (inv *Inventory) GetHelpWithOctaveNumber(userID int, currOctave string) (*HelpOctaveNumber, error) {
	currBalance := inv.balance.GetCurrentBalance(userID)

	var helpOctaveNumber HelpOctaveNumber

	if currBalance < costOfHelpers {
		inv.logger.Info("insufficient balance to buy helper")

		helpOctaveNumber.Ok = false

		return &helpOctaveNumber, nil
	}

	helpOctaveNumber.Octave = currOctave
	helpOctaveNumber.Ok = true

	err := inv.balance.UpdateCurrentBalance(userID, -costOfHelpers)

	if err != nil {
		inv.logger.Error("failed to deduct balance", slog.Any("err", err))

		return nil, err
	}

	return &helpOctaveNumber, nil
}

func (inv *Inventory) GetHelpWithNotePosition(userID int, currNote string) (*HelpNotePosition, error) {
	currBalance := inv.balance.GetCurrentBalance(userID)

	var helpNotePosition HelpNotePosition

	if currBalance < costOfHelpers {
		inv.logger.Info("insufficient balance to buy helper")

		helpNotePosition.Ok = false

		return &helpNotePosition, nil
	}

	lestSideNotes := []string{"до#, C#", "до, С", "ми, E", "ре#, D#", "ре, D", "фа, F"}

	helpNotePosition.Pos = "правая часть"

	if slices.Contains(lestSideNotes, currNote) {
		helpNotePosition.Pos = "левая часть"
	}

	helpNotePosition.Ok = true

	err := inv.balance.UpdateCurrentBalance(userID, -costOfHelpers)

	if err != nil {
		inv.logger.Error("failed to deduct balance", slog.Any("err", err))

		return nil, err
	}

	return &helpNotePosition, nil
}
