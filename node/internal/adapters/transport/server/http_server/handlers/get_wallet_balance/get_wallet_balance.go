package getwalletbalance

import (
	"fmt"
	"log"
	"net/http"
	"node/internal/adapters/transport/server/http_server/msgs"
	"github.com/gin-gonic/gin"
)

type getWalletBalance interface {
	GetWalletBalance(address []byte) (int, error)
}

func NewGetWallelBalanceHandler(gwb getWalletBalance) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		publicKey := []byte(ctx.Query("key"))
		res, err := gwb.GetWalletBalance(publicKey)
		if err != nil {
			errMsg := fmt.Sprintf("Ошибка обработки запроса, не удалось получить балланс кошелька: %v", err)
			log.Println(errMsg)

			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: errMsg,
			})
			return
		}

		log.Printf("Адрес кошелька HEX: %x, Адрес STR: %v, Баланс: %d", publicKey, string(publicKey), res)

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
