package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// UserInfo stores the information about a user.
type UserInfo struct {
	UID        int       `json:"uid"`
	Username   *string   `json:"username"`
	Department *string   `json:"department"`
	Created    time.Time `json:"created"`
}

func connect() (db *sql.DB, err error) {
	host := os.Getenv("DB_HOST")
	password := os.Getenv("DB_PASSWORD")
	parameters := []string{"dbname=postgres user=myapp"}
	if host != "" {
		parameters = append(parameters, fmt.Sprintf("host=%s", host))
	}
	if password != "" {
		parameters = append(parameters, fmt.Sprintf("password=%s", password))
	}

	dbinfo := strings.Join(parameters, " ")

	log.Printf("Connecting to pg with parameters : %s", dbinfo)

	db, err = sql.Open("postgres", dbinfo)
	return
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	if db, err := connect(); err == nil {
		defer db.Close()

		switch r.Method {
		case http.MethodPost:
			handleInsertUser(db, w, r)
		case http.MethodGet:
			handleListUsers(db, w, r)
		default:
			handle404(w)
		}
	}
}

func handle404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func handle500(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func sendJSON(data interface{}, w http.ResponseWriter) {
	var err error
	var serialized []byte
	if serialized, err = json.Marshal(data); err != nil {
		handle500(err, w)
	}
	w.Write(serialized)
}

func handleListUsers(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var err error
	var rows *sql.Rows
	var response = make([]UserInfo, 0)

	if rows, err = db.Query("SELECT * FROM userinfo"); err != nil {
		handle500(err, w)
		return
	}
	for rows.Next() {
		entry := UserInfo{}
		if err = rows.Scan(&entry.UID, &entry.Username, &entry.Department, &entry.Created); err != nil {
			handle500(err, w)
			return
		}
		response = append(response, entry)
	}

	sendJSON(response, w)
}

func handleInsertUser(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var err error
	var lastInsertID int
	var body []byte

	userInfo := UserInfo{}
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		handle500(err, w)
		return
	}

	json.Unmarshal(body, &userInfo)
	if err = db.QueryRow("INSERT INTO userinfo(username,department) VALUES($1,$2) returning uid;", userInfo.Username, userInfo.Department).Scan(&lastInsertID); err != nil {
		handle500(err, w)
		return
	}

	response := map[string]int{"uid": lastInsertID}
	sendJSON(response, w)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("/", handleUsers)
	log.Printf("Starting myapp on :80")
	log.Fatal(http.ListenAndServe(":80", logRequest(http.DefaultServeMux)))
}
