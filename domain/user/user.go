package user

type User struct {
	UserID   string
	Username []string
	Email    string
	Password Password
}

func NewUser(
	userID string,
	username []string,
	email string,
	password string,
) User {
	return User{
		UserID:   userID,
		Username: username,
		Email:    email,
		Password: NewPassword(password),
	}
}
