package model

// Commodity define a commodity
type Commodity struct {
	Id 			 string `json:"itemId"`
	Name         string  `json:"itemName"`
	Introduction string  `json:"itemDetails"`
	Picture      string  `json:itemImage"`
	Price        float64 `json:"itemPrice"`
}

// Cart define a shopping cart
type Cart struct {
	Username    string      `json:"username"`
	Commodities []Commodity `json:"commodities"`
}

// User define a user
type User struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Balance  float64 `json:"balance"`
}

// Comment define a comment
type Comment struct {
	Username  string `json:"username"`
	Commodity string `json:"commodity"`
	Comment   string `json:"comment"`
}

//Token define a user's token
type Token struct {
	Username string `json:"username"`
	TokenStr string `json:"tokenstr"`
}

//TokenKey use to save the key in mongoDB
type TokenKey struct {
	Username string `json:"username"`
	Key      string `json:"key"`
}
