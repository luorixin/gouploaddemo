package handler

import (
  "demo/util"
  "github.com/gin-gonic/gin"
  "net/http"
)

func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
  return http.HandlerFunc(
    func(w http.ResponseWriter, r *http.Request) {
      r.ParseForm()
      username := r.Form.Get("username")
      token := r.Form.Get("token")
      if len(username) < 3 || !IsTokenValid(token) {
        w.WriteHeader(http.StatusForbidden)
        return
      }
      h(w, r)
    },
  )
}

func HTTPInterceptorGIN() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")

		if len(username) < 3 || !IsTokenValid(token) {
		  // token失败直接返回失败提示
		  c.Abort()
			resp := util.NewRespMsg(
				int(-1),
				"token无效",
				nil,
			)
			c.JSON(http.StatusOK, resp)
      return
		}
    c.Next()
	}
}
