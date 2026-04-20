package login

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/JesterForAll/gonote/internal/database"
	"github.com/JesterForAll/gonote/internal/session"
)

type LoginHandler struct {
	loginStruct  *loginStruct
	logger       *slog.Logger
	tokenManager *session.TokenManager
}

type createUserRequest struct {
	Name string `json:"name"`
}

type userData struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

type userID struct {
	ID string `json:"id"`
}

func New(logger *slog.Logger, tokenManager *session.TokenManager) (*LoginHandler, error) {
	loginStruct, err := newLogin(logger)
	if err != nil {
		logger.Error("failed create login struct", slog.Any("err", err))

		return nil, err
	}

	return &LoginHandler{loginStruct: loginStruct, logger: logger, tokenManager: tokenManager}, nil
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

	var userID userID

	err := json.NewDecoder(r.Body).Decode(&userID)
	if err != nil {
		loginHand.logger.Error("error decoding create user request", slog.Any("err", err))
		http.Error(w, "Bad request, error while decoding body", http.StatusBadRequest)

		return
	}

	userIDint, err := strconv.Atoi(userID.ID)
	if err != nil {
		loginHand.logger.Error("error converting user id in create user request", slog.Any("err", err))
		http.Error(w, "Bad request, error converting user id", http.StatusBadRequest)

		return
	}

	var userDB database.LoginDBStruct

	exist := loginHand.loginStruct.DB.CheckIfExistAndGetFirst(map[string]interface{}{"id": userIDint}, &userDB)
	if !exist {
		loginHand.logger.Error("login error, selected user doesnt exist", slog.Any("err", err))
		http.Error(w, "Bad request, selected user doesnt exist", http.StatusBadRequest)

		return
	}

	cookieVal := strconv.Itoa(int(userDB.ID)) + ":" + loginHand.tokenManager.GetToken()

	cookie := &http.Cookie{
		Name:     "user_id",
		Value:    cookieVal,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	}
	http.SetCookie(w, cookie)

	loginHand.logger.Info("user logined", "user_name", userDB.UserName, "user id", userDB.ID)

	http.Redirect(w, r, "/main", http.StatusSeeOther)

}
