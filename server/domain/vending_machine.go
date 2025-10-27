package domain

type ItemRequest struct {
	ItemID int `json:"item_id"`
	Amount int `json:"amount"`
}

type VendingMachineItem struct {
	ID        int     `json:"id" graphql:"id"`
	Amount    int     `json:"amount" graphql:"amount"`
	Price     float64 `json:"price" graphql:"price"`
	MotorCode string  `json:"motor_code" graphql:"motor_code"`
}

type Transaction struct {
	ID            string  `json:"id" graphql:"id"`
	CombinationID int     `json:"combination_id" graphql:"combination_id"`
	Price         float64 `json:"price" graphql:"price"`
	Status        bool    `json:"status" graphql:"status"`
	TexRef        string  `json:"tex_ref" graphql:"tex_ref"`
}
