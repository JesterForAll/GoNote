package balance

import (
	"log/slog"

	"github.com/JesterForAll/gonote/internal/database"
	"gorm.io/gorm"
)

type Balance struct {
	logger *slog.Logger
	Db     *database.Database
}

const BalancePlusForWin int = 10

func newBalance(logger *slog.Logger) (*Balance, error) {
	db, err := database.New("../../static/balance.db", logger, database.BalanceDbStruct{})
	if err != nil {
		logger.Error("failed to connect database", slog.Any("err", err))

		return nil, err
	}

	return &Balance{logger: logger, Db: db}, nil
}

func (balance *Balance) GetCurrentBalance(userID int) int {

	var balanceDB database.BalanceDbStruct

	exist := balance.Db.CheckIfExistAndGetFirst(map[string]interface{}{"user_id": userID}, &balanceDB)

	if !exist {
		return 0
	}

	return balanceDB.Balance
}

func (balance *Balance) UpdateCurrentBalance(userID int, val int) error {

	var balanceDB database.BalanceDbStruct

	exist := balance.Db.CheckIfExistAndGetFirst(map[string]interface{}{"user_id": userID}, &balanceDB)

	if !exist {
		balanceDB.UserID = userID
	}

	balanceDB.Balance += val

	if balanceDB.Balance < 0 {
		balanceDB.Balance = 0
	}

	err := balance.Db.Upsert(&balanceDB)
	if err != nil {
		balance.logger.Error("error updating balance", slog.Any("err", err))

		return err
	}

	return nil
}

func (balance *Balance) UpdateCurrentBalanceWithTx(tx *gorm.DB, userID int, val int) error {

	var balanceDB database.BalanceDbStruct

	exist := balance.Db.CheckIfExistAndGetFirst(map[string]interface{}{"user_id": userID}, &balanceDB)

	if !exist {
		balanceDB.UserID = userID
	}

	balanceDB.Balance += val

	if balanceDB.Balance < 0 {
		balanceDB.Balance = 0
	}

	err := tx.Save(&balanceDB).Error
	if err != nil {
		balance.logger.Error("error updating balance", slog.Any("err", err))

		return err
	}

	return nil
}
