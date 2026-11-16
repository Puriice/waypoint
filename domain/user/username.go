package user

type Username string

func NewUsername(username string) (Username, error) {
	return Username(username), nil
}

func (u Username) String() string {
	return string(u)
}
