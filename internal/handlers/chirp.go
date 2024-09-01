package handlers


import (
	"strconv"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// Getter for ID
func (c Chirp) GetID() int {
	return c.ID
}

func (c Chirp) GetUniqueIdentifier() string {
	return strconv.Itoa(c.ID)
}

// Setter for ID
func (c *Chirp) SetID(id int) {
	c.ID = id
}
