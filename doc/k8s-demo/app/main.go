package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "database/sql"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "os"
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
    http.ListenAndServe(":8080", r)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello World!")
}

// SECTION: NOTE

type Note struct {
    Title     string `json:"title"`
    Description string `json:"description"`
}

func getNoteHandler(w http.ResponseWriter, r *http.Request) {
    /*
        The list of notes is now taken from the store instead of the package level variable we had earlier
    */
    notes, err := store.GetNotes()
    if err != nil {
        fmt.Println(fmt.Errorf("Error: %v", err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    // Everything else is the same as before
    noteListBytes, err := json.Marshal(notes)

    if err != nil {
        fmt.Println(fmt.Errorf("Error: %v", err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    w.Write(noteListBytes)
}

func createNoteHandler(w http.ResponseWriter, r *http.Request) {
    note := Note{}

    err := r.ParseForm()

    if err != nil {
        fmt.Println(fmt.Errorf("Error: %v", err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    note.Title = r.Form.Get("title")
    note.Description = r.Form.Get("description")

    // The only change we made here is to use the `CreateNote` method instead of
    // appending to the `note` variable like we did earlier
    err = store.CreateNote(&note)
    if err != nil {
        fmt.Println(err)
    }

    http.Redirect(w, r, "/assets/", http.StatusFound)
}

// SECTION: STORE

// Our store will have two methods, to add a new note,
// and to get all existing notes
// Each method returns an error, in case something goes wrong
type Store interface {
    CreateNote(note *Note) error
    GetNotes() ([]*Note, error)
}

// The `dbStore` struct will implement the `Store` interface
// It also takes the sql DB connection object, which represents
// the database connection.
type dbStore struct {
    db *sql.DB
}

func (store *dbStore) CreateNote(note *Note) error {
    // 'Note' is a simple struct which has "title" and "description" attributes
    // THe first underscore means that we don't care about what's returned from
    // this insert query. We just want to know if it was inserted correctly,
    // and the error will be populated if it wasn't
    _, err := store.db.Query("INSERT INTO notes(title, description) VALUES ($1,$2)", note.Title, note.Description)
    return err
}

func (store *dbStore) GetNotes() ([]*Note, error) {
    // Query the database for all notes, and return the result to the
    // `rows` object
    rows, err := store.db.Query("SELECT title, description from notes")
    // We return incase of an error, and defer the closing of the row structure
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Create the data structure that is returned from the function.
    // By default, this will be an empty array of notes
    notes := []*Note{}
    for rows.Next() {
        // For each row returned by the table, create a pointer to a note,
        note := &Note{}
        // Populate the `Title` and `Description` attributes of the note,
        // and return incase of an error
        if err := rows.Scan(&note.Title, &note.Description); err != nil {
            return nil, err
        }
        // Finally, append the result to the returned array, and repeat for
        // the next row
        notes = append(notes, note)
    }
    return notes, nil
}

// The store variable is a package level variable that will be available for
// use throughout our application code
var store Store

/*
We will need to call the InitStore method to initialize the store. This will
typically be done at the beginning of our application (in this case, when the server starts up)
This can also be used to set up the store as a mock, which we will be observing
later on
*/
func InitStore(s Store) {
    store = s
}
