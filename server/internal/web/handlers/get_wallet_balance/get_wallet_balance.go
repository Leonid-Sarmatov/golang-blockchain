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
		// Данные из тела запроса
		req := &msgs.RequestWalletBalance{}

		// Получение запроса с публичным адресом кошелька
		if err := ctx.ShouldBindJSON(req); err != nil {
			// Если JSON отсутствует или неверный, возвращаем ошибку
			ctx.JSON(http.StatusBadRequest, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: "Некорректный запрос, ошибка JSON парсинга",
			})
			return
		}

		res, err := gwb.GetWalletBalance(req.PublicKey)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: "Ошибка обработки запроса, не удалось получить балланс кошелька",
			})
			return
		}

		log.Printf("Адрес HEX: %x, Адрес STR: %v, Баланс: %d", req.PublicKey, string(req.PublicKey), res)

		ctx.JSON(http.StatusOK, &msgs.ResponseWalletBalance{
			BaseResponse: msgs.BaseResponse{
				Status: "OK!",
			},
			RequestWalletBalance: *req,
			Balance:              res,
		})
	}
}
