package coinstransfer

import (
	"golang_blockchain/internal/web/msgs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createCoinTransfer interface {
	CreateCoinTransfer(amount int, recipientAddress, senderAddress []byte) error
}

func NewCoinTransferHandler(cct createCoinTransfer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := &msgs.RequestCoinsTransfer{}

		if err := ctx.ShouldBindJSON(req); err != nil {
			// Если JSON отсутствует или неверный, возвращаем ошибку
			ctx.JSON(http.StatusBadRequest, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: "Некорректный запрос, ошибка JSON парсинга",
			})
			return
		}

		err := cct.CreateCoinTransfer(req.Amount, []byte(req.RecipientKey), []byte(req.SenderKey))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: "Ошибка обработки запроса, перевод средств прерван",
			})
			return
		}

		ctx.JSON(http.StatusOK, &msgs.ResponseCoinsTransfer{
			BaseResponse: msgs.BaseResponse{
				Status: "OK!",
			},
			RequestCoinsTransfer: *req,
		})
	}
}
