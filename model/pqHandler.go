package model

import (
	"database/sql"
	"time"

	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
)

type pqHandler struct {
	db *sql.DB
}

func (s *pqHandler) GetTodos(sessionId string) []*Todo {
	todos := []*Todo{}
	sql_string := "SELECT id, name, completed, createdAt From todos WHERE sessionId = $1"
	rows, err := s.db.Query(sql_string, sessionId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		rows.Scan(&todo.ID, &todo.Name, &todo.Completed, &todo.CreatedAt)
		todos = append(todos, &todo)
	}
	return todos
}

func (s *pqHandler) AddTodo(name, sessionId string) *Todo {
	sql_string := "INSERT INTO todos (sessionId, name, completed, createdAt) VALUES($1,$2,$3,now()) RETURNING id"
	stmt, err := s.db.Prepare(sql_string)
	if err != nil {
		panic(err)
	}

	var id int
	err = stmt.QueryRow(sessionId, name, false).Scan(&id)
	if err != nil {
		panic(err)
	}

	var todo Todo
	todo.ID = id
	todo.Name = name
	todo.Completed = false
	todo.CreatedAt = time.Now()

	return &todo
}

func (s *pqHandler) RemoveTodo(id int) bool {
	sql_string := "DELETE FROM todos WHERE id = $1"
	stmt, err := s.db.Prepare(sql_string)
	if err != nil {
		panic(err)
	}

	result, err := stmt.Exec(id)
	if err != nil {
		panic(err)
	}
	cnt, _ := result.RowsAffected()

	return cnt > 0
}

func (s *pqHandler) CompleteTodo(id int, complete bool) bool {
	sql_string := "UPDATE todos SET completed = $1 WHERE id = $2"
	stmt, err := s.db.Prepare(sql_string)
	if err != nil {
		panic(err)
	}

	result, err := stmt.Exec(complete, id)
	if err != nil {
		panic(err)
	}
	cnt, _ := result.RowsAffected()

	return cnt > 0
}

func (s *pqHandler) Close() {
	s.db.Close()
}

func newPQHandler(dbConn string) DBHandler {
	database, err := sql.Open("postgres", dbConn)
	if err != nil {
		panic(err)
	}
	statement, err := database.Prepare(
		`CREATE TABLE IF NOT EXISTS todos (
			id        SERIAL PRIMARY KEY,
			sessionId VARCHAR(256),
			name      TEXT,
			completed BOOLEAN,
			createdAt TIMESTAMP
		);`)
	if err != nil {
		panic(err)
	}
	_, err = statement.Exec()
	if err != nil {
		panic(err)
	}

	statement, err = database.Prepare(
		`CREATE INDEX IF NOT EXISTS sessionIdIndexOnTodos ON todos(sessionId ASC);`)
	if err != nil {
		panic(err)
	}
	_, err = statement.Exec()
	if err != nil {
		panic(err)
	}

	return &pqHandler{db: database}
	// return &pqHandler{}
}
