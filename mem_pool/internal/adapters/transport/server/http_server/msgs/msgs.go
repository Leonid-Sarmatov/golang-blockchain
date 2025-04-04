package msgs

/* Основа ответа */
type BaseResponse struct {
	Status       string
	ErrorMessage string `json:"ErrorMessage,omitempty"`
}

/* Работа с кошельком */
type RequestWallet struct {
	Address string
}

type RequestWalletBalance struct {
	PublicKey string
}

type ResponseWalletBalance struct {
	BaseResponse
	RequestWalletBalance
	Balance int
}

/* Работа с майнером */
type RequestWork struct {
	RewardAddress string
}

type ResponseWork struct {
	BaseResponse
	Block string
}

type RequestCompletedWork struct {
	ResponseWork
	BlockPOW int
}

/* Переводы коинов */
type RequestCoinsTransfer struct {
	Amount       int
	SenderKey    string
	RecipientKey string
}

type ResponseCoinsTransfer struct {
	BaseResponse
	RequestCoinsTransfer
}
