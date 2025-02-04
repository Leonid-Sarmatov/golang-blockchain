package msgs

type BaseResponse struct {
	Status       string
	ErrorMessage string `json:"ErrorMessage,omitempty"`
}

type RequestWork struct {
	ConbaseTransaction string
	Transaction        string
}

type ResponseWork struct {
	BaseResponse
	RequestWork
	ConbaseTransactionPOW int
	TransactionPOW        int
}

type RequestWalletBalance struct {
	PublicKey string
}

type ResponseWalletBalance struct {
	BaseResponse
	RequestWalletBalance
	Balance int
}

type RequestCoinsTransfer struct {
	Amount       int
	SenderKey    string
	RecipientKey string
}

type ResponseCoinsTransfer struct {
	BaseResponse
	RequestCoinsTransfer
}
