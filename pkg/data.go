package pkg

import "time"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	RefreshHash  string `json:"-"`
}

type User struct {
	UserId string
	IP     string
}

type Session struct {
	RefreshHash string
	UserId      string
	UserIP      string
	UserEmail   string
	CreatedAt   time.Time
}
