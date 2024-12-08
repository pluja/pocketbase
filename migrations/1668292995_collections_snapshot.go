package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Auto generated migration with the most recent collections configuration.
func init() {
	m.Register(func(app core.App) error {
		jsonData := `[
    {
        "id": "4ea6djy249fzqfg",
        "listRule": null,
        "viewRule": null,
        "createRule": null,
        "updateRule": null,
        "deleteRule": null,
        "name": "posts_admin",
        "type": "base",
        "fields": [
            {
                "autogeneratePattern": "[a-z0-9]{15}",
                "hidden": false,
                "id": "text3208210256",
                "max": 15,
                "min": 15,
                "name": "id",
                "pattern": "^[a-z0-9]+$",
                "presentable": false,
                "primaryKey": true,
                "required": true,
                "system": true,
                "type": "text"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text1542800728",
                "max": 0,
                "min": 0,
                "name": "field",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": false,
                "system": false,
                "type": "text"
            }
        ],
        "indexes": [],
        "system": false
    },
    {
        "id": "5lt8hsy53jvfhwt",
        "listRule": "",
        "viewRule": "",
        "createRule": "",
        "updateRule": "",
        "deleteRule": "",
        "name": "posts_public",
        "type": "base",
        "fields": [
            {
                "autogeneratePattern": "[a-z0-9]{15}",
                "hidden": false,
                "id": "text3208210256",
                "max": 15,
                "min": 15,
                "name": "id",
                "pattern": "^[a-z0-9]+$",
                "presentable": false,
                "primaryKey": true,
                "required": true,
                "system": true,
                "type": "text"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text1542800728",
                "max": 0,
                "min": 0,
                "name": "field",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": false,
                "system": false,
                "type": "text"
            }
        ],
        "indexes": [],
        "system": false
    },
    {
        "id": "78h06xtuhv8dqzx",
        "listRule": "@request.auth.id != \"\"",
        "viewRule": "@request.auth.id != \"\"",
        "createRule": "@request.auth.id != \"\"",
        "updateRule": "@request.auth.id != \"\"",
        "deleteRule": "@request.auth.id != \"\"",
        "name": "posts_user",
        "type": "base",
        "fields": [
            {
                "autogeneratePattern": "[a-z0-9]{15}",
                "hidden": false,
                "id": "text3208210256",
                "max": 15,
                "min": 15,
                "name": "id",
                "pattern": "^[a-z0-9]+$",
                "presentable": false,
                "primaryKey": true,
                "required": true,
                "system": true,
                "type": "text"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text1542800728",
                "max": 0,
                "min": 0,
                "name": "field",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": false,
                "system": false,
                "type": "text"
            }
        ],
        "indexes": [],
        "system": false
    }
]`

		collections := []map[string]any{}
		if err := json.Unmarshal([]byte(jsonData), &collections); err != nil {
			return err
		}

		return app.ImportCollections(collections, false)
	}, func(_ core.App) error {
		// no revert since the configuration on the environment, on which
		// the migration was executed, could have changed via the UI/API
		return nil
	})
}
