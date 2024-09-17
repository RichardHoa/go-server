package handlers

type DataPoint interface {
	GetID() int
	GetUniqueIdentifier() string
}

type Database struct {
	Chirps map[string]Chirp `json:"chirps"`
	Users  map[string]User  `json:"users"`
}
