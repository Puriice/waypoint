package port

type Hasher interface {
	Hash(input string) (string, error)
	Verify(input string, hash string) bool
}
