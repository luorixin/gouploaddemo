package handler

import (
	dblayer "demo/db"
	"demo/util"
  "fmt"
  "io/ioutil"
	"net/http"
  "time"
)

const (
	pwd_salt = "*#890"
	token_salt = "_tokensalt"
)

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")
	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}
	enc_passwd := util.Sha1([]byte(passwd + pwd_salt))
	suc := dblayer.UserSignup(username, enc_passwd)
	if suc {
		w.Write([]byte("Success"))
	} else {
		w.Write([]byte("Failed"))
	}
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  username := r.Form.Get("username")
  password := r.Form.Get("password")

  encPasswd := util.Sha1([]byte(password + pwd_salt))


  pwdChecked := dblayer.UserSignin(username, encPasswd)
  if !pwdChecked {
    w.Write([]byte("Failed"))
    return
  }
  //生产token
  token := GenToekn(username)
  upRes := dblayer.UpdateToken(username, token)
  if !upRes {
    w.Write([]byte("Failed"))
    return
  }

  //w.Write([]byte("http://"+r.Host +"/static/view/home.html"))
  resp := util.RespMsg{
    Code: 0,
    Msg:  "OK",
    Data: struct {
      Location string
      Username string
      Token string
    }{
      Location:"http://"+r.Host+"/static/view/home.html",
      Username:username,
      Token:token,
    },
  }
  w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  username := r.Form.Get("username")
  // 放入拦截器处理
  //token := r.Form.Get("token")
  //isValidToken := IsTokenValid(token)
  //if !isValidToken {
  //  w.WriteHeader(http.StatusForbidden)
  //  return
  //}

  user, err := dblayer.GetUserInfo(username)
  if err != nil {
    w.WriteHeader(http.StatusForbidden)
    return
  }
  resp := util.RespMsg{
    Code: 0,
    Msg:  "OK",
    Data: user,
  }
  w.Write(resp.JSONBytes())

}

func GenToekn(username string) string {
  //username+timestamp+tokenSalt md5
  ts:=fmt.Sprintf("%x", time.Now().Unix())
  tokenPrefix := util.MD5([]byte(username+ts+token_salt))
  return tokenPrefix + ts[:8]
}

func IsTokenValid(token string) bool {
  if len(token) != 40 {
    return false
  }
  // TODO: 判断token的时效性，是否过期
  // TODO: 从数据库表tbl_user_token查询username对应的token信息
  // TODO: 对比两个token是否一致
  return true
}
