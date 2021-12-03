package myapp

import (
	"heroku/todos/model"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var rd *render.Render = render.New()

// 인터페이스 외부로 공개를 해야함, close 함수의 권한을 main.go 에 넘겨주기위해서
// app.go 에도 언제 close 할지 모르므로 main.go 로 이관.
// 그래서 AppHandler 만들어서 넘겨줌.
type AppHandler struct {
	// 임베디드 처리함. 이유는 ??? 상속과 비슷한데 조금 다름 is 관계는 아니고, has 관계임.
	// 이름을 암시적으로 생략함. handler 생략함.
	http.Handler
	db model.DBHandler
}

// func getSessionID(r *http.Request) string {
var getSessionID = func(r *http.Request) string {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	// Set some session values.
	val := session.Values["id"]
	if val == nil {
		return ""
	}
	return val.(string)
}

func (a *AppHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/todo.html", http.StatusTemporaryRedirect)
}

func (a *AppHandler) getTodoListHandler(w http.ResponseWriter, r *http.Request) {
	sessionsId := getSessionID(r)
	list := a.db.GetTodos(sessionsId)
	rd.JSON(w, http.StatusOK, list)
}

func (a *AppHandler) addTodoHandler(w http.ResponseWriter, r *http.Request) {
	sessionsId := getSessionID(r)
	name := r.FormValue("name")
	todo := a.db.AddTodo(name, sessionsId)
	rd.JSON(w, http.StatusCreated, todo)
}

type Success struct {
	Success bool `json:"success"`
}

func (a *AppHandler) removeTodoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	ok := a.db.RemoveTodo(id)
	if ok {
		rd.JSON(w, http.StatusOK, Success{Success: true})
	} else {
		rd.JSON(w, http.StatusOK, Success{Success: false})
	}
}

func (a *AppHandler) completeTodoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	complete := r.FormValue("complete") == "true"
	ok := a.db.CompleteTodo(id, complete)
	if ok {
		rd.JSON(w, http.StatusOK, Success{Success: true})
	} else {
		rd.JSON(w, http.StatusOK, Success{Success: false})
	}
}

func (a *AppHandler) Close() {
	a.db.Close()
}

func CheckSignin(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// if request RRL is /singin.html, then next()
	if strings.Contains(r.URL.Path, "/signin") || strings.Contains(r.URL.Path, "/auth") {
		next(rw, r)
		return
	}

	// if user already signed in
	sessionID := getSessionID(r)
	if sessionID != "" {
		next(rw, r)
		return
	}
	// if user not sign in
	// redirect signin.html
	http.Redirect(rw, r, "/signin.html", http.StatusTemporaryRedirect)
}

// 리턴변경 http.Handler -> AppHandler
func MakeNewHandler(filepath string) *AppHandler {

	mux := mux.NewRouter()
	// Classic() *Negroni => return New(NewRecovery(), NewLogger(), NewStatic(http.Dir("public")))
	// ng := negroni.Classic()
	ng := negroni.New(negroni.NewRecovery(), negroni.NewLogger(), negroni.HandlerFunc(CheckSignin), negroni.NewStatic(http.Dir("public")))
	ng.UseHandler(mux)

	a := &AppHandler{
		Handler: ng, // mux->ng
		db:      model.NewDBHandler(filepath),
	}
	mux.HandleFunc("/", a.indexHandler)
	mux.HandleFunc("/todos", a.getTodoListHandler).Methods("GET")
	mux.HandleFunc("/todos", a.addTodoHandler).Methods("POST")
	mux.HandleFunc("/todos/{id:[0-9]+}", a.removeTodoHandler).Methods("DELETE")
	mux.HandleFunc("/complete-todo/{id:[0-9]+}", a.completeTodoHandler).Methods("GET")
	mux.HandleFunc("/auth/google/login", googleLoginHandler)
	mux.HandleFunc("/auth/google/callback", googleAuthCallback)

	return a
}
