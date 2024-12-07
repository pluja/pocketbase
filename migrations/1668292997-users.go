package migrations

import (
	"database/sql"
	"errors"
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		_, err = app.FindFirstRecordByData(collection.Name, "email", UserEmailPassword)
		exists := true
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				exists = false
			} else {
				return err
			}
		}

		if exists {
			return nil
		}

		log.Println("inserting normal user: ", UserEmailPassword)

		r := core.NewRecord(collection)
		r.SetEmail(UserEmailPassword)
		r.SetVerified(true)
		r.SetPassword(UserEmailPassword)

		return app.Save(r)
	}, func(_ core.App) error {
		return nil
	})
}
