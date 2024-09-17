package domain

type UserDto struct {
	User        User   `json:"user"`
	AccessToken string `json:"access_token"`
}
