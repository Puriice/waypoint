package cmd

type Login struct {
	Email    string
	Password string
}

func NewLogin(email, password string) Login {
	return Login{
		Email:    email,
		Password: password,
	}
}
