package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"dump"

	"golang.org/x/net/html"
)

//go:embed post_table.html
var post_table string

//go:embed type-here.svg
var typeHere string

// provides a https form field
func editor(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, post_table)
}

func parsed(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	doc, _ := html.Parse(strings.NewReader(r.Form["table"][0]))
	database := parseTable(doc)
	fmt.Fprint(w, dump.Dump(dump.Section{
		Title:   "SQL",
		Content: "Generation code writen by Oliver Day\n--note to self: should we create a database and have a `use` statment",
		Sub: []dump.Section{
			{
				Title:   "Create",
				Content: handleEmptySql(databaseFrom(database)),
				Sub:     nil,
			},
			{
				Title:   "Insert",
				Content: handleEmptySql(insertionFrom(database, func(s string) int { return 5 })),
				Sub:     nil,
			},
		}},
	))
}

func empty(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "image/svg+xml")
	fmt.Fprint(w, typeHere)
}

const port = "5090"

func main() {
	http.HandleFunc("/", editor)
	http.HandleFunc("/parsed", parsed)
	http.HandleFunc("/type-here", empty)
	// http.HandleFunc("/htmx.js", htmx)
	// http.HandleFunc("/", err404)

	fmt.Println("Starting server at port", port)
	server := "127.0.0.1:" + port
	if err := http.ListenAndServe(server, nil); err != nil {
		panic(err)
	}
	// browser.OpenURL("https://" + server)
}
