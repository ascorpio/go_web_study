package main

import (
	_ "github.com/go-sql-driver/mysql"
	"goblog/app/http/middlewares"
	"goblog/bootstrap"
	"net/http"
)

func main() {
	bootstrap.SetupDB()
	router := bootstrap.SetupRoute()

	_ = http.ListenAndServe(":3000", middlewares.RemoveTrailingSlash(router))
}
