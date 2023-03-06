package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"goblog/bootstrap"
	"goblog/pkg/database"
	"goblog/pkg/logger"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

var router *mux.Router
var db *sql.DB

// Article 对应一条文章数据
type Article struct {
	ID          int64
	Title, Body string
}

// Link 方法用来生成文章链接
func (a Article) Link() string {
	showURL, err := router.Get("articles.show").URL("id", strconv.FormatInt(a.ID, 10))
	if err != nil {
		logger.LogError(err)
		return ""
	}
	return showURL.String()
}

// Delete 方法用以从数据库中删除单条记录
func (a Article) Delete() (int64, error) {
	rs, err := db.Exec("delete from articles where id = " + strconv.FormatInt(a.ID, 10))

	if err != nil {
		return 0, err
	}

	if n, _ := rs.RowsAffected(); n > 0 {
		return n, nil
	}

	return 0, nil
}

func getArticleByID(id string) (Article, error) {
	article := Article{}
	query := "SELECT * FROM articles WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	return article, err
}

func articlesEditHandler(w http.ResponseWriter, r *http.Request) {
	id := getRouteVariable("id", r)

	article, err := getArticleByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		updateURL, _ := router.Get("articles.update").URL("id", id)
		data := ArticlesFormData{
			Title:  article.Title,
			Body:   article.Body,
			URL:    updateURL,
			Errors: nil,
		}
		tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
		logger.LogError(err)

		err = tmpl.Execute(w, data)
		logger.LogError(err)
	}
}

func validateArticleFormData(title string, body string) map[string]string {
	errors := make(map[string]string)
	// 验证标题
	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需介于 3-40"
	}

	// 验证内容
	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于或等于 10 个字节"
	}

	return errors
}

func articlesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := getRouteVariable("id", r)

	_, err := getArticleByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		title := r.PostFormValue("title")
		body := r.PostFormValue("body")

		errors := validateArticleFormData(title, body)

		if len(errors) == 0 {
			query := "update articles set title = ?, body = ? where id = ?"
			rs, err := db.Exec(query, title, body, id)

			if err != nil {
				logger.LogError(err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "500 服务器内部错误")
			}

			if n, _ := rs.RowsAffected(); n > 0 {
				showURL, _ := router.Get("articles.show").URL("id", id)
				http.Redirect(w, r, showURL.String(), http.StatusFound)
			} else {
				fmt.Fprint(w, "您没有做任何更改！")
			}
		} else {
			updateURL, _ := router.Get("articles.update").URL("id", id)
			data := ArticlesFormData{
				Title:  title,
				Body:   body,
				URL:    updateURL,
				Errors: errors,
			}
			tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
			logger.LogError(err)

			err = tmpl.Execute(w, data)
			logger.LogError(err)
		}
	}
}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select * from articles")
	logger.LogError(err)
	defer rows.Close()

	var articles []Article

	for rows.Next() {
		var article Article
		err = rows.Scan(&article.ID, &article.Title, &article.Body)
		logger.LogError(err)
		articles = append(articles, article)
	}

	err = rows.Err()
	logger.LogError(err)

	tmpl, err := template.ParseFiles("resources/views/articles/index.gohtml")
	logger.LogError(err)

	err = tmpl.Execute(w, articles)
	logger.LogError(err)
}

// ArticlesFormData 创建博文表单数据
type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

// 创建文章
func articleCreateHandler(w http.ResponseWriter, r *http.Request) {
	parseArticleForm(&w, "", "", make(map[string]string))
}

// 修改文章
func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

	errors := make(map[string]string)

	// 检查是否有错误
	if len(errors) == 0 {
		lastInsertID, err := saveArticleToDB(title, body)
		if lastInsertID > 0 {
			fmt.Fprint(w, "插入成功，ID为"+strconv.FormatInt(lastInsertID, 10))
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		parseArticleForm(&w, title, body, errors)
	}
}

func articlesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := getRouteVariable("id", r)

	article, err := getArticleByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		rowsAffected, err := article.Delete()
		if err != nil {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		} else {
			if rowsAffected > 0 {
				indexURL, _ := router.Get("articles.index").URL()
				http.Redirect(w, r, indexURL.String(), http.StatusFound)
			} else {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, "404 文章未找到")
			}
		}
	}
}

func saveArticleToDB(title, body string) (int64, error) {
	// 变量初始化
	var (
		id   int64
		err  error
		rs   sql.Result
		stmt *sql.Stmt
	)

	// 1. 获取一个 prepare 声明语句
	stmt, err = db.Prepare("insert into articles(title,body) values(?,?)")

	// 错误检测
	if err != nil {
		return 0, err
	}

	// 2. 在此函数运行结束后关闭语句，防止占用 SQL 连接
	defer stmt.Close()

	// 3. 执行请求，传参进入绑定的内容
	rs, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}

	// 4. 插入成功的话，会返回自增ID
	if id, err = rs.LastInsertId(); id > 0 {
		return id, nil
	}
	return 0, err
}

// 解析创建文章、修改文章模版
func parseArticleForm(w *http.ResponseWriter, title, body string, errors map[string]string) {
	storeURL, _ := router.Get("articles.store").URL()
	data := ArticlesFormData{
		Title:  title,
		Body:   body,
		URL:    storeURL,
		Errors: errors,
	}
	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(*w, data)
	if err != nil {
		panic(err)
	}
}

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

// getRouteVariable 获取 URI 路由参数
func getRouteVariable(parameterName string, r *http.Request) string {
	vars := mux.Vars(r)
	return vars[parameterName]
}

func main() {

	database.Initialize()
	db = database.DB

	bootstrap.SetupDB()
	router = bootstrap.SetupRoute()

	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articleCreateHandler).Methods("GET").Name("articles.create")
	router.HandleFunc("/articles/{id:[0-9]+}/edit", articlesEditHandler).Methods("GET").Name("articles.edit")
	router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")
	router.HandleFunc("/articles/{id:[0-9]+}/delete", articlesDeleteHandler).Methods("POST").Name("articles.delete")

	// 中间件：强制内容类型为 HTML
	router.Use(forceHTMLMiddleware)

	// 通过命名路由获取 URL 示例
	homeUrl, _ := router.Get("home").URL()
	fmt.Println("homeURL:", homeUrl)

	articleURL, _ := router.Get("articles.show").URL("id", "23")
	fmt.Println("articleURL:", articleURL)

	_ = http.ListenAndServe(":3000", removeTrailingSlash(router))
}
