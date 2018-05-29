package main

// Task object
type Task struct {
	Title       string `json:"title"`
	Assigned    string `json:"assigned"`
	Deadline    string `json:"deadline"`
	Status      string `json:"status"`
	Discussion  string `json:"discussion"`
	Description string `json:"description"`
}

// Project object
type Project struct {
	Title   string `json:"title"`
	Creator string `json:"creator"`
	Status  string `json:"status"`
}
