package handlers

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (user User) GetID() (ID int) {
	return user.ID
}
