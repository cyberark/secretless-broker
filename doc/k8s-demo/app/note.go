package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

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
    decoder := json.NewDecoder(r.Body)
    var note Note
    err := decoder.Decode(&note)
    if err != nil {
        fmt.Println(fmt.Errorf("Error: %v", err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    // The only change we made here is to use the `CreateNote` method instead of
    // appending to the `note` variable like we did earlier
    err = store.CreateNote(&note)
    if err != nil {
        fmt.Println(err)
    }

    http.Redirect(w, r, "/assets/", http.StatusFound)
}
