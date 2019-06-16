package model

// UsersWallet defines the info of each user's wallet
type UsersWallet struct {
	UserID  string `json:"userID"`
	Balance int64  `json:"balance"`
}
