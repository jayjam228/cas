package models

type JWT struct {
	Token   string    `json:"token"`
	Secret_key   string    `json:"secret_key"`
    Algorithm   string    `json:"algorithm"`
    Expire   string    `json:"expire"`
}