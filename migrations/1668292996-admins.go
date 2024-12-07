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
		superUsersCol := func() *core.Collection {
			col, err := app.FindCachedCollectionByNameOrId(core.CollectionNameSuperusers)
			if err != nil {
				panic(err)
			}
			return col
		}()

		exists := true
		_, err := app.FindAuthRecordByEmail(superUsersCol,
			AdminEmailPassword)
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

		log.Println("inserting admin user: ", AdminEmailPassword)

		record := core.NewRecord(superUsersCol)
		record.SetEmail(AdminEmailPassword)
		record.SetPassword(AdminEmailPassword)

		return app.Save(record)
	}, func(_ core.App) error {
		return nil
	})
}
