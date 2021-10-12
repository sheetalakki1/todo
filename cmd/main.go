package main

import (
	"fmt"
	"go_todo/logger"
	"go_todo/router"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// main function
func main() {
	sugarLogger := logger.InitLogger()

	router.Router()
	zap.L().Info(fmt.Sprintf("Listening & Serving on : %s", os.Getenv("APP_PORT")))
	err := http.ListenAndServe(os.Getenv("SERVER_PORT"), nil)
	if err != nil {
		sugarLogger.Errorf("Failed to start server %s", err.Error())
	}
}
