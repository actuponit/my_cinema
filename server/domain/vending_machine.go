package domain

type VendingMachineItem struct {
	ID     int     `json:"id" graphql:"id"`
	Amount int     `json:"amount" graphql:"amount"`
	Price  float64 `json:"price" graphql:"price"`
}

type Transaction struct {
	ID            string  `json:"id" graphql:"id"`
	CombinationID int64   `json:"combination_id" graphql:"combination_id"`
	Price         float64 `json:"price" graphql:"price"`
	Status        bool    `json:"status" graphql:"status"`
	TexRef        string  `json:"tex_ref" graphql:"tex_ref"`
}
