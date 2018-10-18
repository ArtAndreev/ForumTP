package queries

import (
	"database/sql"
)

func getLastInsertedID(res *sql.Rows) (int, error) {
	id := 0
	for res.Next() {
		err := res.Scan(&id)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}
