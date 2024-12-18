package migrations

import (
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		for _, c := range []string{PostsAdmin, PostsUser, PostsPublic} {
			collection, err := app.FindCollectionByNameOrId(c)
			if err != nil {
				return err
			}

			log.Println("inserting post to: ", c)

			r := core.NewRecord(collection)
			r.Set("field", "test")

			if err := app.Save(r); err != nil {
				return err
			}
		}

		return nil
	}, func(_ core.App) error {
		return nil
	})
}
