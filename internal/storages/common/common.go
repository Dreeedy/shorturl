package common

type URLData []URLItem

type URLItem struct {
	UUID          string
	Hash          string
	OriginalURL   string
	OperationType string
	CorrelationID string
	ShortURL      string
}
