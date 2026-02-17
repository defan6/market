package domain

const (
	RoleUser    = "user"
	RoleManager = "manager"
	RoleAdmin   = "admin"
)

type User struct {
	ID           int64  `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password"`
	Role         string `db:"role"`
}

type UserDetails struct {
	ID    int64
	Email string
	Role  string
}

func NewUserDetails(id int64, email string, role string) UserDetails {
	return UserDetails{
		ID:    id,
		Email: email,
		Role:  role,
	}
}
