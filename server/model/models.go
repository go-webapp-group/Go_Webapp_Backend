package model

type Commodity struct {
	Name         string  `json:"name"`
	Introduction string  `json:"introduction"`
	Picture      string  `json:"picture"`
	Price        float64 `json:"price"`
}

/*
- shopping cart
    - username (string)
	- commodity list (list[commodity, commodity, ...])*/

type Cart struct {
	Username    string      `json:"username"`
	Commodities []Commodity `json:"commodities"`
}

// User ...
type User struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Balance  float64 `json:"balance"`
}

type Comment struct {
	Username  string `json:"username"`
	Commodity string `json:"commodity"`
	Comment   string `json:"comment"`
}
