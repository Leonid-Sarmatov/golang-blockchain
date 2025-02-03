package msgs

type BaseResponse struct {
	Status string
	ErrorMessage string
}

type RequestWork struct {
	ConbaseTransaction []byte
	Transaction []byte
}

type ResponseWork struct {
	BaseResponse
	RequestWork
	ConbaseTransactionPOW int
	TransactionPOW int
}

type RequestWalletBalance struct {
	PublicKey []byte
}

type ResponseWalletBalance struct {
	BaseResponse
	RequestWalletBalance
	Balance int
}
