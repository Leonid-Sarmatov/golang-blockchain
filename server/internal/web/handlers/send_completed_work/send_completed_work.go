package sendcompletedwork

import (
	"encoding/base64"
	"fmt"
	"golang_blockchain/internal/web/msgs"
	hashcalulator "golang_blockchain/pkg/hash_calulator"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type acceptWork interface {
	SendCompletedWork(bytesRewardBlock, bytesMainBlock []byte, rewardBlockPOW, mainBlockPOW int) error
}

func NewSendCompletedWorkHandler(aw acceptWork) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := &msgs.RequestCompletedWork{}

		if err := ctx.ShouldBindJSON(req); err != nil {
			// Если JSON отсутствует или неверный, возвращаем ошибку
			ctx.JSON(http.StatusBadRequest, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: "Некорректный запрос, ошибка JSON парсинга",
			})
			return
		}

		d := hashcalulator.NewHashCalculator()
		decodedBlock1, _ := base64.StdEncoding.DecodeString(req.RewardBlock)
		decodedBlock2, _ := base64.StdEncoding.DecodeString(req.MainBlock)
		fmt.Println(">>>", d.HashCalculate(decodedBlock1), "\n>>>", d.HashCalculate(decodedBlock2))
		fmt.Printf("Блок 1 = %x\n, Блок 2 = %x\n", decodedBlock1, decodedBlock2)
		
		err := aw.SendCompletedWork(decodedBlock1, decodedBlock2, req.RewardBlockPOW, req.MainBlockPOW)
		if err != nil {
			errMsg := fmt.Sprintf("Ошибка обработки запроса, не удалось принять работу: %v", err)
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