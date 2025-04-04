package createwallet

import (
	"fmt"
	"mem_pool/internal/adapters/transport/server/http_server/msgs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createWallet interface {
	CreateWallet(address []byte) error
}

func NewCreateWalletHandler(cw createWallet) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := &msgs.RequestWallet{}

		if err := ctx.ShouldBindJSON(req); err != nil {
			// Если JSON отсутствует или неверный, возвращаем ошибку
			ctx.JSON(http.StatusBadRequest, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: "Некорректный запрос, ошибка JSON парсинга",
			})
			return
		}
		
		err := cw.CreateWallet([]byte(req.Address))
		if err != nil {
			errMsg := fmt.Sprintf("Ошибка обработки запроса, не удалось создать кошелек: %v", err)
			log.Println(errMsg)

			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: errMsg,
			})
			return
		}

		ctx.JSON(http.StatusOK, &msgs.BaseResponse{
			Status: "OK!",
		})
	}
}