package handlers


type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// Getter for ID
func (c Chirp) GetID() int {
	return c.ID
}

// Setter for ID
func (c *Chirp) SetID(id int) {
	c.ID = id
}
