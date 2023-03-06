package controllers

import (
	"net/http"
)

// ArticlesController 文章相关页面
type ArticlesController struct {
}

// Show 文章详情页面
func (*ArticlesController) Show(w http.ResponseWriter, r *http.Request) {
	//id := getRouteVariable("id", r)
	//
	//article, err := getArticleByID(id)
	//
	//// 3. 如果出现错误
	//if err != nil {
	//	if err == sql.ErrNoRows {
	//		// 3.1 数据未找到
	//		w.WriteHeader(http.StatusNotFound)
	//		fmt.Fprint(w, "404文章未找到")
	//	} else {
	//		// 3.2 数据库错误
	//		logger.LogError(err)
	//		w.WriteHeader(http.StatusInternalServerError)
	//		fmt.Fprint(w, "500 服务器内部错误")
	//	}
	//} else {
	//	// 4. 读取成功
	//	tmpl, err := template.New("show.gohtml").Funcs(template.FuncMap{
	//		"RouteName2URL": route.RouteName2URL,
	//		"Int64ToString": types.Int64ToString,
	//	}).ParseFiles("resources/views/articles/show.gohtml")
	//	logger.LogError(err)
	//	err = tmpl.Execute(w, article)
	//	logger.LogError(err)
	//}
}
