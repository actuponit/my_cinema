package domain

type PaymentRequest struct {
	Amount   float64 `json:"amount" validate:"required"`
	Currency string  `json:"currency" validate:"required,eq=ETB"`
	TxRef    string  `json:"tx_ref"`
	// Customization Customization `json:"customization"`
}

type ChapaPaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    *struct {
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
}
