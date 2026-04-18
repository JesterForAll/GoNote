package login

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type LoginHandler struct {
	loginStruct *loginStruct
	logger      *slog.Logger
}

type createUserRequest struct {
	Name string `json:"name"`
}

type userData struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

func New(logger *slog.Logger) (*LoginHandler, error) {
	loginStruct, err := newLogin(logger)
	if err != nil {
		logger.Error("failed create login struct", slog.Any("err", err))

		return nil, err
	}

	return &LoginHandler{loginStruct: loginStruct, logger: logger}, nil
}

func (loginHand *LoginHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {

	var createUserRequest createUserRequest

	err := json.NewDecoder(r.Body).Decode(&createUserRequest)
	if err != nil {
		loginHand.logger.Error("error decoding create user request", slog.Any("err", err))
		http.Error(w, "Bad request, error while decoding body", http.StatusBadRequest)

		return
	}

	loginHand.logger.Info("got input\n", "createUserRequest", createUserRequest)

	err = loginHand.loginStruct.createUser(createUserRequest.Name)
	if err != nil {
		loginHand.logger.Error("error creating user", slog.Any("err", err))
		http.Error(w, "Bad request, error while creating user", http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

func (loginHand *LoginHandler) HandleGetUsers(w http.ResponseWriter, _ *http.Request) {

	usersListDB, err := loginHand.loginStruct.getUsers()
	if err != nil {
		loginHand.logger.Error("error getting users from database", slog.Any("err", err))
		http.Error(w, "Internal  server error, rror getting users from database", http.StatusInternalServerError)

		return
	}

	usersList := make([]userData, 0, len(*usersListDB))

	for _, data := range *usersListDB {
		usersList = append(usersList, userData{Name: data.UserName, ID: data.ID})
	}

	data, err := json.Marshal(usersList)
	if err != nil {
		loginHand.logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	loginHand.logger.Info("response for get users\n", "data", data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(data)
}

func (loginHand *LoginHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {

	// data, err := os.ReadFile("../../static/index.html")

	// if err != nil {
	// 	loginHand.logger.Error("error reading main page", slog.Any("err", err))
	// 	http.Error(w, "Internal server error while reading main page", http.StatusInternalServerError)

	// 	return
	// }

	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.WriteHeader(http.StatusOK)

	// w.Write(data)

	http.Redirect(w, r, "/main", http.StatusSeeOther)

}
