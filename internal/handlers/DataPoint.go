package handlers

type DataPoint interface {
	GetID() int
	GetUniqueIdentifier() string
}