package main

import (
	"fmt"
	"github.com/ivahaev/go-logger"
	"os"
	"time"
	"yandex-money-go/ya-money-go"
)

func init()  {
	logger.SetLevel("DEBUG")
}

func main() {

	token := os.Getenv("YATOKEN")

	ya, _ := yamoney.NewYaMoney(token)

	t := time.Now()
	till, _ := time.Parse( "2006-Jan-02 00:00:00", t.Format("2006-Jan-02 00:00:00"))
	from, _ := time.Parse( "2006-Jan-02 00:00:00", t.AddDate(0, -1, 0).Format("2006-Jan-02 00:00:00"))

	account_info,_ := ya.Account_info()
	logger.JSON(account_info)


	param := fmt.Sprintf("from=%s&till=%s", from.Format(time.RFC3339), till.Format(time.RFC3339))
	logger.Debug(param)
	operations, _ := ya.OperationHistory(param)

	logger.JSON(operations)
}
