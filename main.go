package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./store.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(
		`
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			price REAL
		);
	`)

	if err != nil {
		log.Fatal(err)
	}
}
func homePage(w http.ResponseWriter, r *http.Request) {
	// Fetch items from the database
	rows, err := db.Query("SELECT id, name, price FROM items")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Price); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	// Log the fetched items for debugging purposes
	log.Println("Fetched items:", items)

	// Parse the template (using template.Must which handles errors internally)
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))

	// Execute the template with the items data
	err = tmpl.Execute(w, items) // <-- This will use the base template name (index.html)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the success of template rendering
	log.Printf("Rendering template with items: %+v", items)
}

func addItem(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		priceStr := r.FormValue("price")

		// Convert price to float64
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			http.Error(w, "Invalid price value", http.StatusBadRequest)
			return
		}

		// Log the received data
		log.Printf("Received item: Name=%s, Price=%f", name, price)

		// Insert item into the database
		_, err = db.Exec("INSERT INTO items (name, price) VALUES (?, ?)", name, price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to the homepage to see the new list of items
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Route handler for deleting an item
func deleteItem(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	_, err := db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	// Initialize routes
	r := mux.NewRouter()
	r.HandleFunc("/", homePage).Methods("GET")
	r.HandleFunc("/add", addItem).Methods("POST")
	r.HandleFunc("/delete/{id}", deleteItem).Methods("GET")

	// Start the server
	log.Println("Server started at :8080")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
