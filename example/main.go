package main

import (
	"errors"
	"log"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pluja/pocketbase"
)

type Post struct {
	Field string `json:"field"`
	ID    string `json:"id"`
}

func main() {
	// REMEMBER to start the Pocketbase before running this example with `make serve` command

	var errs error
	client := pocketbase.NewClient("http://localhost:8090")
	// Other configuration options:
	// pocketbase.WithAdminEmailPassword("admin@admin.com", "admin@admin.com")
	// pocketbase.WithUserEmailPassword("user@user.com", "user@user.com")
	// pocketbase.WithUserToken(token)
	// pocketbase.WithAdminToken(token)
	// pocketbase.WithDebug()

	response, err := client.Collection("posts_public").List(pocketbase.ParamsList{
		Size:    1,
		Page:    1,
		Sort:    "-created",
		Filters: "field~'test'",
	})

	errs = errors.Join(errs, err)

	log.Printf("Total items: %d, total pages: %d\n", response.TotalItems, response.TotalPages)
	for _, item := range response.Items {
		var test Post
		err := mapstructure.Decode(item, &test)
		errs = errors.Join(errs, err)

		log.Printf("Item: %#v\n", test)
	}

	log.Println("Inserting new item")
	// you can use struct type - just make sure it has JSON tags
	collection := pocketbase.Collection[Post]{
		Client: client,
		Name:   "posts_public",
	}
	_, err = collection.Create(Post{
		Field: "test_" + time.Now().Format(time.Stamp),
	})
	errs = errors.Join(errs, err)

	// or you can use simple map[string]any
	r, err := client.Collection("posts_public").Create(map[string]any{
		"field": "test_" + time.Now().Format(time.Stamp),
	})
	errs = errors.Join(errs, err)

	err = client.Collection("posts_public").Delete(r.ID)
	errs = errors.Join(errs, err)

	if errs != nil {
		log.Fatal(errs)
	}
}
