package walletcontroller

type WalletController struct {
}

func NewWalletController() (*WalletController, error) {
	return &WalletController{}, nil
}

/*
AddBlock добавляет новый блок, и проверяет proof-of-work

Аргументы:
  - []byte: data данные блока (полезная нагрузка в виде транзакции)
  - int: pwValue доказательство работы

Возвращает:
  - error: ошибка
*/
func CreateNewWallet() error {
	return nil
}
