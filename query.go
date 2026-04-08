package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=db port=5432 user=bap password=bap_admin_1 dbname=bap_app sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO roller (role_id, name, description) VALUES (4, 'hakem', 'Hakem') ON CONFLICT (role_id) DO UPDATE SET name=EXCLUDED.name;")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Hakem role inserted successfully with role_id=4.")
}
