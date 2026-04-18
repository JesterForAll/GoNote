package login

import (
	"errors"
	"log/slog"

	"github.com/JesterForAll/gonote/internal/database"
)

type loginStruct struct {
	logger *slog.Logger
	DB     *database.Database
}

func newLogin(logger *slog.Logger) (*loginStruct, error) {

	db, err := database.New("../../static/login.db", logger, &database.LoginDBStruct{})
	if err != nil {
		logger.Error("failed to connect database", slog.Any("err", err))

		return nil, err
	}

	return &loginStruct{DB: db, logger: logger}, nil
}

func (login *loginStruct) createUser(name string) error {

	var nameData database.LoginDBStruct

	exist := login.DB.CheckIfExist(map[string]interface{}{"user_name": name}, &nameData)

	if exist {
		return errors.New("This user is already exist")
	}

	nameData.UserName = name

	login.DB.Upsert(&nameData)

	return nil
}

func (login *loginStruct) getUsers() (*[]database.LoginDBStruct, error) {

	var usersList []database.LoginDBStruct

	err := login.DB.GetAll(&usersList)
	if err != nil {
		login.logger.Error("error getting from database", slog.Any("err", err))

		return nil, err
	}

	return &usersList, nil
}
