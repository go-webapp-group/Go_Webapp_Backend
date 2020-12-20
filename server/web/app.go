package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"
	"webapp/db"
	"webapp/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

//App define a app
type App struct {
	d        db.DB
	handlers map[string]http.HandlerFunc
}

//Serve start the webapp server
func (a *App) Serve() error {
	for path, handler := range a.handlers {
		http.Handle(path, handler)
	}

	log.Println("Web server is available on port 8080")
	return http.ListenAndServe(":8080", nil)
}

func sendErr(w http.ResponseWriter, code int, message string) {
	resp, _ := json.Marshal(map[string]string{"error": message})
	http.Error(w, string(resp), code)
}

// Needed in order to disable CORS for local development
func disableCors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		h(w, r)
	}
}

// NewApp init the webapp routes
func NewApp(d db.DB, cors bool) App {
	app := App{
		d:        d,
		handlers: make(map[string]http.HandlerFunc),
	}

	commodityHandler := app.GetCommodity
	commoditiesHandler := app.GetCommodities
	userRegister := app.UserRegister
	getAUserInfo := app.GetAUserInfo
	getUsersInfo := app.GetUsersInfo

	if !cors {
		commodityHandler = disableCors(commodityHandler)
		commoditiesHandler = disableCors(commoditiesHandler)
		///
		getUsersInfo = disableCors(getUsersInfo)
		getAUserInfo = disableCors(getAUserInfo)
		userRegister = disableCors(userRegister)
	}
	//分配路径
	app.handlers["/picture/"] = http.StripPrefix("/picture/", http.FileServer(http.Dir("./picture"))).ServeHTTP
	app.handlers["/picture/upload"] = recieveImage
	app.handlers["/commodities/"] = commodityHandler
	app.handlers["/commodities"] = commoditiesHandler

	app.handlers["/users"] = getUsersInfo
	app.handlers["/users/"] = getAUserInfo
	app.handlers["/users/register"] = userRegister
	app.handlers["/"] = writeApiRoot
	return app
}

//writeApiRoot return all the api server list
func writeApiRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type Url struct {
		urlName  string
		urlValue string
	}
	type UForm struct {
		URLFORM []Url `json:"urlform"`
	}
	//Api信息
	apiStr := make(map[string]string)
	apiStr["all_commodities_url"] = "http://localhost:8080/commodities"
	apiStr["post_commoditie_url:http"] = "//localhost:8080/commodities"
	apiStr["get_commoditie_info_url"] = "http://localhost:8080/commodities/{commodity}"
	apiStr["get_comment_url"] = "http://localhost:8080/commodities/{commodity}/comments"
	apiStr["post_comment_url"] = "http://localhost:8080/commodities"
	apiStr["delete_comment_url"] = "http://localhost:8080/commodities"
	apiStr["get_alluser_url"] = "http://localhost:8080/users"
	apiStr["user_register_url"] = "http://localhost:8080/users/register"
	apiStr["get_a_user_url"] = "http://localhost:8080/users/{user}"
	apiStr["get_user_cart"] = "http://localhost:8080/users/{user}/cart"
	apiStr["update_comment_url"] = "http://localhost:8080/users/{user}/cart"
	apiStr["get_picture"] = "http://localhost:8080/picture/{picture}"
	apiStr["post_picture"] = "http://localhost:8080/picture/upload"
	//发送到根root
	err := json.NewEncoder(w).Encode(apiStr)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}

//revcieveImage  recieve the upload image
func recieveImage(w http.ResponseWriter, r *http.Request) {
	//获取图片文件
	f, h, err := r.FormFile("image")
	fmt.Println("recieve a img")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//在路径./picture/下创建新文件
	filename := h.Filename
	defer f.Close()
	t, err := os.Create("./picture/" + filename)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer t.Close()
	//将图片复制到同名新文件
	if _, err := io.Copy(t, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetCommodities if r.Method is GET it will get all commodities, if r.Method is POST will add or update a commodity
func (a *App) GetCommodities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("get all commodities")
	//获取所有商品信息
	if r.Method == "GET" {
		//从数据库取所有商品信息的数据
		commodities, err := a.d.GetAllCommodity()
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		//将信息写入response
		err = json.NewEncoder(w).Encode(commodities)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	} else { //为商店添加新商品
		var commodity model.Commodity
		r.ParseMultipartForm(512)
		fmt.Println("Add a new commodity")
		for k, v := range r.MultipartForm.Value {
			fmt.Println("value,k,v = ", k, ",", v)
		}
		//将信息写入数据库
		commodity.Introduction = r.MultipartForm.Value["introduction"][0]
		commodity.Name = r.MultipartForm.Value["name"][0]
		commodity.Picture = r.MultipartForm.Value["picture"][0]
		//string转float64
		commodity.Price, _ = strconv.ParseFloat(r.MultipartForm.Value["price"][0], 64)
		a.d.PostCommodity(&commodity)
	}

}

// GetCommodity define all the api in the form as /commodities/
func (a *App) GetCommodity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//获取url“commodities/”后面的字段
	urlStr := r.URL.String()
	fmt.Println("get a commodity")
	commodityname := urlStr[len("/commodities")+1:]
	//判断是否是操作商品（/commodities/{}/comments评论的url
	if isComment(urlStr) {
		fmt.Println("Operate a commodity's comment")
		a.OperateCommentsForCM(w, r)
	} else { //请求单个商品的详细信息
		//将url编码解编码
		cName, _ := url.QueryUnescape(commodityname)
		println(cName)
		//从数据库获取信息
		commodity, err := a.d.GetOneCommodity(cName)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		//写信息
		err = json.NewEncoder(w).Encode(commodity)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	}
}

// OperateCommentsForCM get post patch delete the coments of a commodity
func (a *App) OperateCommentsForCM(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" { //获取某商品的评论
		w.Header().Set("Content-Type", "application/json")
		//获取商品名
		urlStr := r.URL.String()
		println(urlStr)
		commodityname := urlStr[len("/commodities")+1 : len(urlStr)-len("/comments")]
		cName, _ := url.QueryUnescape(commodityname)
		fmt.Println("get comments for a commodity", cName)
		//从数据库取数据
		commemts, err := a.d.GetCommentsForCM(cName)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		//写数据
		err = json.NewEncoder(w).Encode(commemts)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	} else if r.Method == "POST" { //发布新评论
		var comment model.Comment
		r.ParseMultipartForm(128)
		for k, v := range r.MultipartForm.Value {
			fmt.Println("value,k,v = ", k, ",", v)
		}
		comment.Username = r.MultipartForm.Value["username"][0]
		comment.Comment = r.MultipartForm.Value["comment"][0]
		comment.Commodity = r.MultipartForm.Value["commodity"][0]
		a.d.WriteComment(&comment)
		fmt.Println(comment)
	} else if r.Method == "DELETE" { //删除评论
		fmt.Println("Delete a commet")
		var comment model.Comment
		r.ParseMultipartForm(128)
		for k, v := range r.MultipartForm.Value {
			fmt.Println("value,k,v = ", k, ",", v)
		}
		comment.Username = r.MultipartForm.Value["username"][0]
		comment.Comment = r.MultipartForm.Value["comment"][0]
		comment.Commodity = r.MultipartForm.Value["commodity"][0]
		a.d.DeleteComment(&comment)
	} else if r.Method == "PATCH" { //修改某评论
		fmt.Println("Update a comment")
		var comment model.Comment
		r.ParseMultipartForm(128)
		comment.Username = r.MultipartForm.Value["username"][0]
		comment.Comment = r.MultipartForm.Value["comment"][0]
		comment.Commodity = r.MultipartForm.Value["commodity"][0]
		a.d.UpdateComment(&comment)
	}

}

//isComment return true if name is a comment or false is not a comment
func isComment(name string) bool {
	commentMode := regexp.MustCompile(`[A-Za-z0-9%]+/comments`)
	return commentMode.MatchString(name)

}

// GetUsersInfo get all usersinfo
func (a *App) GetUsersInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user, err := a.d.GetUsersInfo()
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}

//isGetCart :判断url是否是请求用户的购物车
func isGetCart(urlStr string) bool {
	commentMode := regexp.MustCompile(`[A-Za-z0-9%]+/cart`)
	return commentMode.MatchString(urlStr)

}

// GetAUserInfo get a userinfo获取某用户信息，以及用户的购物车信息获取和修改
func (a *App) GetAUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := r.URL.String()[len("/users/"):]
	cUsername, _ := url.QueryUnescape(username)
	if isGetCart(username) { //购物车信息的获取和修改
		fmt.Println("Get a cart")
		a.GetAUserCart(w, r)
	} else { //获取用户的详细信息
		fmt.Println("Get A user Info")
		//从数据库取信息
		user, err := a.d.GetAUserInfo(cUsername)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		err = json.NewEncoder(w).Encode(user)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	}
}

// GetAUserCart if r.Method is GET it will get the cartInfo of a user or update the cartInfo if the r.Method is POST
//该Api需要token认证，检验客户端发送过来的token字段
func (a *App) GetAUserCart(w http.ResponseWriter, r *http.Request) {
	//获取用户名
	urlStr := r.URL.String()
	username := urlStr[len("/users/") : len(urlStr)-len("/cart")]
	cUsername, _ := url.QueryUnescape(username)
	//从数据库获取该用户的密钥
	akey, _ := a.d.GetAToken(cUsername)
	//进行token认证
	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(akey.Key), nil
		})
	if err == nil { //如果token字段正确，允许对用户的购物车信息进行申请和修改
		if token.Valid {
			fmt.Println("token is valid")
			w.WriteHeader(http.StatusOK)

			if r.Method == "GET" { //获取购物车信息
				w.Header().Set("Content-Type", "application/json")
				fmt.Println("get cart for a user")
				//从数据库取数据
				cart, err := a.d.GetCart(cUsername)
				if err != nil {
					sendErr(w, http.StatusInternalServerError, err.Error())
					return
				}
				//写数据
				err = json.NewEncoder(w).Encode(cart)
				if err != nil {
					sendErr(w, http.StatusInternalServerError, err.Error())
				}
			} else { //修改用户购物车，即对数据进行如果存在则更新，如果不存在则添加的操作
				var cart model.Cart
				r.ParseMultipartForm(512)
				for k, v := range r.MultipartForm.Value {
					fmt.Println("value,k,v = ", k, ",", v)
				}

				cart.Username = r.MultipartForm.Value["username"][0]
				//coms是一个字符串，需要进行反序列化为json格式结构体
				coms := r.MultipartForm.Value["commodities"][0]
				if err := json.Unmarshal([]byte(coms), &cart.Commodities); err == nil {
					fmt.Println("change string to json :", cart.Commodities)
				} else {
					log.Fatal(err)
				}
				//写数据
				a.d.WriteCart(&cart)
				w.Write([]byte("Successfully updated shopping cart information"))
			}
		} else { //token已经失效，可能是过了有效期
			fmt.Println("Token is not valid")
			w.WriteHeader(http.StatusUnauthorized)
		}
	} else { //token字段错误，无权访问
		fmt.Println("Unauthorized access to this resource")
		w.WriteHeader(http.StatusUnauthorized)
	}

}

// UserRegister register
//用户注册，将当前时间，有效时间，和密钥（用户名，密码）生成token返回给客户端
func (a *App) UserRegister(w http.ResponseWriter, r *http.Request) {
	//从request获取用户的信息
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	balance1 := r.PostFormValue("balance")
	balance, _ := strconv.ParseFloat(balance1, 64)
	//往数据库添加用户
	_, err := a.d.UserRegister(username, password, balance)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	//生成Token：
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	//失效时间
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(2)).Unix()
	//生效时间
	claims["iat"] = time.Now().Unix()
	token.Claims = claims
	//密钥字段
	secretKey := username + password
	//生成token字符串
	tokenStr, _ := token.SignedString([]byte(secretKey))
	/*
		var user model.User
		user.Balance = balance
		user.Password = password
		user.Username = username
	*/
	w.Header().Set("Content-Type", "application/json")
	/*
		err = json.NewEncoder(w).Encode(user)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}*/
	//生成结构体Token,包括用户名和token字符串，写进response
	var tokenStruct model.Token
	tokenStruct.Username = username
	tokenStruct.TokenStr = tokenStr
	err = json.NewEncoder(w).Encode(tokenStruct)
	//生成TokenKey结构体，包括用户名和密钥，储存入数据库
	var tkey model.TokenKey
	tkey.Key = secretKey
	tkey.Username = username
	a.d.AddToken(&tkey)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}
