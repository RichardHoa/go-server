package handlers

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
}

func (user User) GetID() (ID int) {
	return user.ID
}

func (user User) GetUniqueIdentifier() (uniqueIdentifier string) {
	return user.Email
}