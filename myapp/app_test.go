package myapp

import (
	"encoding/json"
	"heroku/todos/model"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodos(t *testing.T) {
	getSessionID = func(r *http.Request) string {
		return "gettestsessionId"
	}

	os.Remove("./test.db")
	assert := assert.New(t)
	appH := MakeNewHandler("./test.db")
	defer appH.Close()

	ts := httptest.NewServer(appH)
	defer ts.Close()

	// add testing - first data
	resp, err := http.PostForm(ts.URL+"/todos", url.Values{"name": {"Test todo"}})
	assert.NoError(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)

	todo := new(model.Todo)
	err = json.NewDecoder(resp.Body).Decode(&todo)
	assert.NoError(err)
	assert.Equal(todo.Name, "Test todo")
	id1 := todo.ID
	log.Println("app_test.go / add result >", *todo, id1)

	// add testing - second data
	resp, err = http.PostForm(ts.URL+"/todos", url.Values{"name": {"Test todo2"}})
	assert.NoError(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)

	// todo = new(Todo)
	err = json.NewDecoder(resp.Body).Decode(&todo)
	assert.NoError(err)
	assert.Equal(todo.Name, "Test todo2")
	id2 := todo.ID
	log.Println("app_test.go / add result >", *todo, id2)

	// get testing - getting whole data
	resp, err = http.Get(ts.URL + "/todos")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	todos := []*model.Todo{}
	err = json.NewDecoder(resp.Body).Decode(&todos)
	assert.NoError(err)
	assert.Equal(2, len(todos))
	for i := 0; i < len(todos); i++ {
		log.Println("after getting whole data >", *todos[i])
	}

	// complete testing
	resp, err = http.Get(ts.URL + "/complete-todo/" + strconv.Itoa(id2) + "?complete=true")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/todos")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	// todos := []*Todo{}
	err = json.NewDecoder(resp.Body).Decode(&todos)
	assert.NoError(err)
	assert.Equal(2, len(todos))
	for i := 0; i < len(todos); i++ {
		log.Println("after complete for >", *todos[i])
	}

	// DELETE testing
	req, _ := http.NewRequest("DELETE", ts.URL+"/todos/"+strconv.Itoa(id1), nil) // data는 필요없어서 nil 처리
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	log.Println("delete id1: ", id1)

	resp, err = http.Get(ts.URL + "/todos")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	// todos := []*Todo{}
	err = json.NewDecoder(resp.Body).Decode(&todos)
	assert.NoError(err)
	assert.Equal(1, len(todos))
	for i := 0; i < len(todos); i++ {
		log.Println("after delete for >", *todos[i])
	}

}
