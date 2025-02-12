package msgs

type BaseResponse struct {
	Status       string
	ErrorMessage string `json:"ErrorMessage,omitempty"`
}

type RequestWallet struct {
	Address string
}

type RequestWork struct {
	RewardAddress string
}

type ResponseWork struct {
	BaseResponse
	RewardBlock string
	MainBlock   string
}

type RequestCompletedWork struct {
	ResponseWork
	RewardBlockPOW int
	MainBlockPOW   int
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
