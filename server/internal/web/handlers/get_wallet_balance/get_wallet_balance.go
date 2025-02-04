package getwalletbalance

import (
	"golang_blockchain/internal/web/msgs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type getWalletBalance interface {
	GetWalletBalance(address []byte) (int, error)
}

func NewGetWallelBalanceHandler(gwb getWalletBalance) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		/*// Данные из тела запроса
		req := &msgs.RequestWalletBalance{}

		// Получение запроса с публичным адресом кошелька
		if err := ctx.ShouldBindJSON(req); err != nil {
			// Если JSON отсутствует или неверный, возвращаем ошибку
			ctx.JSON(http.StatusBadRequest, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: "Некорректный запрос, ошибка JSON парсинга",
			})
			return
		}*/

		publicKey := []byte(ctx.Query("key"))
		res, err := gwb.GetWalletBalance(publicKey)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: "Ошибка обработки запроса, не удалось получить балланс кошелька",
			})
			return
		}

		log.Printf("Адрес HEX: %x, Адрес STR: %v, Баланс: %d", publicKey, string(publicKey), res)

		ctx.JSON(http.StatusOK, &msgs.ResponseWalletBalance{
			BaseResponse: msgs.BaseResponse{
				Status: "OK!",
			},
			RequestWalletBalance: msgs.RequestWalletBalance{
				PublicKey: string(publicKey),
			},
			Balance: res,
		})
	}
}
