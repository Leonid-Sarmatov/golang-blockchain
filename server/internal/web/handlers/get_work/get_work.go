package getwork

import (
	"encoding/base64"
	"fmt"
	"golang_blockchain/internal/web/msgs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type getWork interface {
	GetWorkForMining(rewardAddress []byte) ([]byte, []byte, error)
}

func NewGetWorkHandler(gw getWork) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rewardAddress := []byte(ctx.Query("reward_address"))
		log.Printf("reward_address: STR = %s, BYTE: %v", string(rewardAddress), rewardAddress)
		work1, work2, err := gw.GetWorkForMining(rewardAddress)
		if err != nil {
			errMsg := fmt.Sprintf("Ошибка обработки запроса, не удалось выдать работу: %v", err)
			log.Println(errMsg)

			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: errMsg,
			})
			return
		}

		log.Printf("Первый блок: %v", base64.StdEncoding.EncodeToString(work1))
		log.Printf("Второй блок: %v", base64.StdEncoding.EncodeToString(work2))

		ctx.JSON(http.StatusOK, &msgs.ResponseWork{
			BaseResponse: msgs.BaseResponse{
				Status: "OK!",
			},
			RewardBlock: base64.StdEncoding.EncodeToString(work1),
			MainBlock: base64.StdEncoding.EncodeToString(work2),
		})
	}
}