package main

type Task struct {
	Title      string `json:"title"`
	Assigned   string `json:"assigned"`
	Deadline   uint64 `json:"deadline"`
	Status     string `json:"status"`
	Discussion string `json:"discussion"`
}

type Issue struct {
	Title string `json:"title"`
	Description string `json:"description"`
}

type Project struct {
	Title   string `json:"title"`
	Creator string `json:"creator"`
	Status  string `json:"status"`
}

type User struct {
	TelegramID string `json:"telegram_id"`
	TrelloID   string `json:"trello_id"`
}
