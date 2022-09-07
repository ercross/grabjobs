package models

type Job struct {
	Title    string   `json:"title"`
	Location Location `json:"location"`
}
