package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"webapp/db"
	"webapp/model"
)

type App struct {
	d        db.DB
	handlers map[string]http.HandlerFunc
}

////////////////////////////////////////////////////////////
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

/////////////////////////////////////////////////////////////////

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

	app.handlers["/picture/"] = http.StripPrefix("/picture/", http.FileServer(http.Dir("./picture"))).ServeHTTP
	app.handlers["/picture/upload"] = recieveImage
	app.handlers["/commodities/"] = commodityHandler
	app.handlers["/commodities"] = commoditiesHandler

	app.handlers["/users"] = getUsersInfo
	app.handlers["/users/"] = getAUserInfo
	app.handlers["/user/register"] = userRegister

	return app
}

func recieveImage(w http.ResponseWriter, r *http.Request) {

	f, h, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := h.Filename
	defer f.Close()

	t, err := os.Create("./picture/" + filename)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer t.Close()

	if _, err := io.Copy(t, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) GetCommodities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("get all commodities")
	if r.Method == "GET" {
		commodities, err := a.d.GetAllCommodity()

		comc := len(commodities)
		println(comc) /////////////////////////////

		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		err = json.NewEncoder(w).Encode(commodities)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	} else {
		var commodity model.Commodity
		r.ParseMultipartForm(512)
		for k, v := range r.MultipartForm.Value {
			fmt.Println("value,k,v = ", k, ",", v)
		}
		commodity.Introduction = r.MultipartForm.Value["introduction"][0]
		commodity.Name = r.MultipartForm.Value["name"][0]
		commodity.Picture = r.MultipartForm.Value["picture"][0]
		commodity.Price, _ = strconv.ParseFloat(r.MultipartForm.Value["price"][0], 64)
		a.d.PostCommodity(&commodity)
	}

}

func (a *App) GetCommodity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	urlStr := r.URL.String()
	println(urlStr) ///////////////
	fmt.Println("get a commodity")
	commodityname := urlStr[len("/commodities")+1:]
	if isComment(commodityname) {
		a.OperateCommentsForCM(w, r)
	} else {
		println(commodityname)
		commodity, err := a.d.GetOneCommodity(commodityname)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		err = json.NewEncoder(w).Encode(commodity)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	}
}

func (a *App) OperateCommentsForCM(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Println("get comments for a commodity")

		urlStr := r.URL.String()
		println(urlStr) ///////////////
		commodityname := urlStr[len("/commodities")+1 : len(urlStr)-len("/comments")]
		println(commodityname)
		commemts, err := a.d.GetCommentsForCM(commodityname)

		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		err = json.NewEncoder(w).Encode(commemts)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	} else if r.Method == "POST" {

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
	} else if r.Method == "DELETE" {
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
	} else if r.Method == "PATCH" {
		fmt.Println("Update a comment")
		var comment model.Comment
		r.ParseMultipartForm(128)
		comment.Username = r.MultipartForm.Value["username"][0]
		comment.Comment = r.MultipartForm.Value["comment"][0]
		comment.Commodity = r.MultipartForm.Value["commodity"][0]
		a.d.UpdateComment(&comment)
	}

}

func isComment(name string) bool {
	commentMode := regexp.MustCompile(`[A-Za-z0-9]+/comments`)
	return commentMode.MatchString(name)

}

///////////////////zjy

// GetUsersInfo ...
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
func isGetCart(urlStr string) bool {
	fmt.Println("cart url", urlStr)
	commentMode := regexp.MustCompile(`[A-Za-z0-9]+/cart`)
	return commentMode.MatchString(urlStr)

}

// GetAUserInfo ...
func (a *App) GetAUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// fmt.Println("url", r.URL.String())
	fmt.Println("Operate on a user")
	username := r.URL.String()[len("/users/"):]
	if isGetCart(username) {
		fmt.Println("Get a cart")
		a.GetAUserCart(w, r)
	} else {
		fmt.Println("Get A user Info")
		user, err := a.d.GetAUserInfo(username)
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

func (a *App) GetAUserCart(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Println("get cart for a user")

		urlStr := r.URL.String()
		println(urlStr) ///////////////
		username := urlStr[len("/users/") : len(urlStr)-len("/cart")]
		println("user:", username)
		cart, err := a.d.GetCart(username)

		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		err = json.NewEncoder(w).Encode(cart)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err.Error())
		}
	} else {
		var cart model.Cart
		r.ParseMultipartForm(512)
		for k, v := range r.MultipartForm.Value {
			fmt.Println("value,k,v = ", k, ",", v)
		}

		cart.Username = r.MultipartForm.Value["username"][0]
		coms := r.MultipartForm.Value["commodities"][0]
		//var commjs []model.Commodity
		if err := json.Unmarshal([]byte(coms), &cart.Commodities); err == nil {
			fmt.Println("change string to json :", cart.Commodities)
		} else {
			log.Fatal(err)
		}
		a.d.WriteCart(&cart)
	}
}

// UserRegister ...
func (a *App) UserRegister(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	balance := 0.0
	//balance, err := strconv.ParseFloat((r.PostFormValue("balance")), 64)

	user, err := a.d.UserRegister(username, password, balance)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}
