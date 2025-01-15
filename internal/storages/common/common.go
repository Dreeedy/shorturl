package common

type SetURLData []SetURLItem

type SetURLItem struct {
	UUID          string
	Hash          string
	OriginalURL   string
	OperationType string
	CorrelationId string
	ShortURL      string
}
