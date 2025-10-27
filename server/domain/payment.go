package domain

type PaymentRequest struct {
	Amount           string `json:"amount" validate:"required"`
	Currency         string `json:"currency" validate:"required,eq=ETB"`
	TxRef            string `json:"tx_ref"`
	VendingMachineID []int  `json:"vending_machine_id"`
	// Customization Customization `json:"customization"`
}

type ChapaPaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    *struct {
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
}

type PaymentResponse struct {
	CheckoutURL string `json:"checkOutUrl"`
	TextRef     string `json:"textRef"`
}
