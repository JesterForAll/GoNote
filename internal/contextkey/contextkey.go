package contextkey

type contextKey string

const userIDKey contextKey = "userID"

func GetUserIDKey() contextKey {
	return userIDKey
}
