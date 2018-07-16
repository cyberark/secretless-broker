package main

import (
    "database/sql"
)

type Store interface {
    CreateNote(note *Note) error
    GetNotes() ([]*Note, error)
}

type dbStore struct { // implements `Store` interface
    db *sql.DB // DB connection
}

func (store *dbStore) CreateNote(note *Note) error {
    _, err := store.db.Query("INSERT INTO notes(title, description) VALUES ($1,$2)", note.Title, note.Description)

    return err
}

func (store *dbStore) GetNotes() ([]*Note, error) {
    rows, err := store.db.Query("SELECT title, description from notes")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    notes := []*Note{}
    for rows.Next() {
        note := &Note{}

        if err := rows.Scan(&note.Title, &note.Description); err != nil {
            return nil, err
        }

        notes = append(notes, note)
    }
    return notes, nil
}

var store Store

func InitStore(s Store) {
    store = s
}
