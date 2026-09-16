package bcrypt

import "golang.org/x/crypto/bcrypt"

type Hasher struct {
	cost int
}

func NewHasher(cost int) *Hasher {
	return &Hasher{
		cost: cost,
	}
}

func (h *Hasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (h *Hasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}
