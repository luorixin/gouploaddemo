package route

import (
  "demo/handler"
  "github.com/gin-gonic/gin"
)

func Router() *gin.Engine{
  router := gin.Default()

  router.Static("/static/", "./static")

  router.GET("/user/signup", handler.SignupHandlerGIN)
  router.POST("/user/signup", handler.DoSignUploadHandler)
  router.GET("/user/signin", handler.SignInHandlerGIN)
  router.POST("/user/signin", handler.DoSignInHandler)
  router.POST("/user/info", handler.UserInfoHandlerGIN)

  router.GET("/file/upload", handler.UploadHandlerGIN)
  router.POST("/file/upload", handler.DoUploadHandler)
  router.GET("/file/upload/suc", handler.UploadSucHandlerGIN)
  router.GET("/file/meta", handler.GetFileMetaHandlerGIN)
  router.POST("/file/query", handler.FileQueryHandlerGIN)
  router.GET("/file/download", handler.DownloadHandlerGIN)
  //router.GET("/file/update", handler.FileUpdateHandler)
  //router.GET("/file/delete", handler.FileDeleteHandler)

  router.Use(handler.HTTPInterceptorGIN())

  return router
}

