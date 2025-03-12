package lora_auth

import (
	"errors"
	"github.com/LorraineWen/lorago/lora_router"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

const JWTToken = " lora_router_token"

type JwtAuth struct {
	//jwt的算法
	Alg string
	//过期时间
	TimeOut        time.Duration
	RefreshTimeOut time.Duration
	//时间函数
	TimeFuc func() time.Time
	//Key
	Key []byte
	//刷新key
	RefreshKey string
	//私钥
	PrivateKey     string
	SendCookie     bool
	Authenticator  func(ctx *lora_router.Context) (map[string]any, error)
	CookieName     string
	CookieMaxAge   int64
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool
	Header         string
	AuthHandler    func(ctx *lora_router.Context, err error)
}

type JwtResponse struct {
	Token        string
	RefreshToken string
}

// 登录  用户认证（用户名密码） -> 用户id 将id生成jwt，并且保存到cookie或者进行返回
func (auth *JwtAuth) LoginHandler(ctx *lora_router.Context) (*JwtResponse, error) {
	data, err := auth.Authenticator(ctx)
	if err != nil {
		return nil, err
	}
	if auth.Alg == "" {
		auth.Alg = "HS256"
	}
	//A部分
	signingMethod := jwt.GetSigningMethod(auth.Alg)
	token := jwt.New(signingMethod)
	//B部分
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for key, value := range data {
			claims[key] = value
		}
	}
	if auth.TimeFuc == nil {
		auth.TimeFuc = func() time.Time {
			return time.Now()
		}
	}
	expire := auth.TimeFuc().Add(auth.TimeOut)
	//过期时间
	claims["exp"] = expire.Unix()
	claims["iat"] = auth.TimeFuc().Unix()
	var tokenString string
	var tokenErr error
	//C部分 secret
	if auth.usingPublicKeyAlgo() {
		tokenString, tokenErr = token.SignedString(auth.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(auth.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	jr := &JwtResponse{
		Token: tokenString,
	}
	//refreshToken
	refreshToken, err := auth.refreshToken(token)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken
	//发送存储cookie
	if auth.SendCookie {
		if auth.CookieName == "" {
			auth.CookieName = JWTToken
		}
		if auth.CookieMaxAge == 0 {
			auth.CookieMaxAge = expire.Unix() - auth.TimeFuc().Unix()
		}
		ctx.SetCookie(auth.CookieName, tokenString, int(auth.CookieMaxAge), "/", auth.CookieDomain, auth.SecureCookie, auth.CookieHTTPOnly)
	}

	return jr, nil
}

func (auth *JwtAuth) usingPublicKeyAlgo() bool {
	switch auth.Alg {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

// refreshToken和Token先比，过期时间更长，当Token过期之后，会使用refreshToken进行验证
func (auth *JwtAuth) refreshToken(token *jwt.Token) (string, error) {

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = auth.TimeFuc().Add(auth.RefreshTimeOut).Unix()
	var tokenString string
	var tokenErr error
	if auth.usingPublicKeyAlgo() {
		tokenString, tokenErr = token.SignedString(auth.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(auth.Key)
	}
	if tokenErr != nil {
		return "", tokenErr
	}
	return tokenString, nil
}

// LogoutHandler 退出登录
func (auth *JwtAuth) LogoutHandler(ctx *lora_router.Context) error {
	if auth.SendCookie {
		if auth.CookieName == "" {
			auth.CookieName = JWTToken
		}
		ctx.SetCookie(auth.CookieName, "", -1, "/", auth.CookieDomain, auth.SecureCookie, auth.CookieHTTPOnly)
		return nil
	}
	return nil
}

// RefreshHandler 刷新token
func (auth *JwtAuth) RefreshHandler(ctx *lora_router.Context) (*JwtResponse, error) {
	rToken, ok := ctx.BasicGet(auth.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}
	if auth.Alg == "" {
		auth.Alg = "HS256"
	}
	//解析token
	t, err := jwt.Parse(rToken.(string), func(token *jwt.Token) (interface{}, error) {
		if auth.usingPublicKeyAlgo() {
			return auth.PrivateKey, nil
		} else {
			return auth.Key, nil
		}
	})
	if err != nil {
		return nil, err
	}
	//B部分
	claims := t.Claims.(jwt.MapClaims)

	if auth.TimeFuc == nil {
		auth.TimeFuc = func() time.Time {
			return time.Now()
		}
	}
	expire := auth.TimeFuc().Add(auth.TimeOut)
	//过期时间
	claims["exp"] = expire.Unix()
	claims["iat"] = auth.TimeFuc().Unix()
	var tokenString string
	var tokenErr error
	//C部分 secret
	if auth.usingPublicKeyAlgo() {
		tokenString, tokenErr = t.SignedString(auth.PrivateKey)
	} else {
		tokenString, tokenErr = t.SignedString(auth.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	jr := &JwtResponse{
		Token: tokenString,
	}
	//refreshToken
	refreshToken, err := auth.refreshToken(t)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken
	//发送存储cookie
	if auth.SendCookie {
		if auth.CookieName == "" {
			auth.CookieName = JWTToken
		}
		if auth.CookieMaxAge == 0 {
			auth.CookieMaxAge = expire.Unix() - auth.TimeFuc().Unix()
		}
		ctx.SetCookie(auth.CookieName, tokenString, int(auth.CookieMaxAge), "/", auth.CookieDomain, auth.SecureCookie, auth.CookieHTTPOnly)
	}

	return jr, nil
}

//jwt登录中间件
//header token 是否

func (auth *JwtAuth) JwtAuthMiddleware(next lora_router.HandleFunc) lora_router.HandleFunc {
	return func(ctx *lora_router.Context) {
		if auth.Header == "" {
			auth.Header = "Authorization"
		}
		token := ctx.R.Header.Get(auth.Header)
		if token == "" {
			if auth.SendCookie {
				cookie, err := ctx.R.Cookie(auth.CookieName)
				if err != nil {
					auth.AuthErrorHandler(ctx, err)
					return
				}
				token = cookie.String()
			}
		}
		if token == "" {
			auth.AuthErrorHandler(ctx, errors.New("token is null"))
			return
		}
		//解析token
		t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if auth.usingPublicKeyAlgo() {
				return auth.PrivateKey, nil
			} else {
				return auth.Key, nil
			}
		})
		if err != nil {
			auth.AuthErrorHandler(ctx, err)
			return
		}
		claims := t.Claims.(jwt.MapClaims)
		ctx.BasicSet("jwt_claims", claims)
		next(ctx)
	}
}

func (auth *JwtAuth) AuthErrorHandler(ctx *lora_router.Context, err error) {
	if auth.AuthHandler == nil {
		ctx.W.WriteHeader(http.StatusUnauthorized)
	} else {
		auth.AuthHandler(ctx, err)
	}
}
