package main

import (
    "fmt"
    "net/http"
    "database/sql"
    "os"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

// SECTION: MAIN

// The new router function creates the router and
// returns it to us. We can now use this function
// to instantiate and test the router outside of the main function
func newRouter() *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/hello", handler).Methods("GET")

    r.HandleFunc("/note", getNoteHandler).Methods("GET")
    r.HandleFunc("/note", createNoteHandler).Methods("POST")
    return r
}

func main() {
    fmt.Println("Starting server...")
    connString := os.Getenv("DATABASE_URL")

    db, err := sql.Open("postgres", connString)
    if err != nil {
        panic(err)
    }
    err = db.Ping()

    if err != nil {
        panic(err)
    }

    InitStore(&dbStore{db: db})

    // The router is now formed by calling the `newRouter` constructor function
    // that we defined above. The rest of the code stays the same
    r := newRouter()
    fmt.Println("Serving on port 8080")
    http.NewServeMux()
    http.ListenAndServe(":8080", r)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello World!")
}
