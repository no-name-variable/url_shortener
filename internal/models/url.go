package models

type URL struct {
	Id    int64  `json:"id"`
	Alias string `json:"alias"`
	Url   string `json:"url"`
}
