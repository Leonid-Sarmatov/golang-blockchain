package coinstransfer

import (
	"mem_pool/internal/transaction"
	"mem_pool/internal/adapters/transport/server/http_server/msgs"
	"net/http"
	"log"
	"fmt"

	"github.com/gin-gonic/gin"
)

type createCoinTransfer interface {
	CreateCoinTransferTransaction(amount int, recipientAddress, senderAddress []byte) (*transaction.Transaction, error)
	AddTransactionToProcessing(t *transaction.Transaction) error
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

		t, err := cct.CreateCoinTransferTransaction(req.Amount, []byte(req.RecipientKey), []byte(req.SenderKey))
		if err != nil {
			errMsg := fmt.Sprintf("Ошибка обработки запроса, не удалось сформировать транзакцию: %v", err)
			log.Println(errMsg)

			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: errMsg,
			})
			return
		}

		err = cct.AddTransactionToProcessing(t)
		if err != nil {
			errMsg := fmt.Sprintf("Ошибка обработки запроса, не удалось добавить транзакцию в очередь на обработку: %v", err)
			log.Println(errMsg)

			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: errMsg,
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
