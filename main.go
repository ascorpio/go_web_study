package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"goblog/bootstrap"
	"net/http"
	"strings"
)

var router *mux.Router

// 设置标头中间件
func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 在设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 2. 继续执行请求
		next.ServeHTTP(w, r)
	})
}

// 移除url尾部多余斜杠
func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 除首页以外，移除所有请求路径后面的斜杆
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	bootstrap.SetupDB()
	router = bootstrap.SetupRoute()

	// 中间件：强制内容类型为 HTML
	router.Use(forceHTMLMiddleware)

	// 通过命名路由获取 URL 示例
	homeUrl, _ := router.Get("home").URL()
	fmt.Println("homeURL:", homeUrl)

	articleURL, _ := router.Get("articles.show").URL("id", "23")
	fmt.Println("articleURL:", articleURL)

	_ = http.ListenAndServe(":3000", removeTrailingSlash(router))
}
