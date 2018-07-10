package main

import (
    "database/sql"
)

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
