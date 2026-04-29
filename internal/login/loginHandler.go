package login

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/JesterForAll/gonote/internal/database"
	"github.com/JesterForAll/gonote/internal/jwt"
)

type LoginHandler struct {
	loginStruct *loginStruct
	logger      *slog.Logger
	jwtManager  *jwt.Manager
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

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

func New(logger *slog.Logger, jwtManager *jwt.Manager) (*LoginHandler, error) {
	loginStruct, err := newLogin(logger)
	if err != nil {
		logger.Error("failed create login struct", slog.Any("err", err))

		return nil, err
	}

	return &LoginHandler{loginStruct: loginStruct, logger: logger, jwtManager: jwtManager}, nil
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

	tokenString, err := loginHand.jwtManager.GenerateToken(int(userDB.ID))
	if err != nil {
		loginHand.logger.Error("error generating JWT token", slog.Any("err", err))
		http.Error(w, "Internal server error while generating token", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse{
		Token:     tokenString,
		ExpiresAt: 24 * 60 * 60,
	})

	loginHand.logger.Info("user logined", "user_name", userDB.UserName, "user id", userDB.ID)
}
