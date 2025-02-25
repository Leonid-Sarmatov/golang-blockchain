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
	GetWorkForMining(rewardAddress []byte) ([]byte, error)
}

func NewGetWorkHandler(gw getWork) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rewardAddress := []byte(ctx.Query("reward_address"))
		log.Printf("<get_work.go> Адрес вознаграждения майнера: STR = %s, BYTE: %v", string(rewardAddress), rewardAddress)
		work, err := gw.GetWorkForMining(rewardAddress)
		if err != nil {
			errMsg := fmt.Sprintf("Ошибка обработки запроса, не удалось выдать работу: %v", err)
			log.Println(errMsg)

			ctx.JSON(http.StatusInternalServerError, &msgs.BaseResponse{
				Status:       "Error",
				ErrorMessage: errMsg,
			})
			return
		}

		log.Printf("<get_work.go> Блок выданный майнеру (для передачи по сети): %v", base64.StdEncoding.EncodeToString(work))

		ctx.JSON(http.StatusOK, &msgs.ResponseWork{
			BaseResponse: msgs.BaseResponse{
				Status: "OK!",
			},
			Block: base64.StdEncoding.EncodeToString(work),
		})
	}
}