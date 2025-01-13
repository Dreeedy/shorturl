package common

type SetURLData []SetURLItem

type SetURLItem struct {
	UUID        string
	ShortURL    string
	OriginalURL string
}

// type GetURLData []GetURLItem

// type GetURLItem struct {
// 	ShortURL string
// }

// type GetURLResult []GetURLItemResult

// type GetURLItemResult struct {
// 	OriginalURL string
// 	Found       bool
// }
