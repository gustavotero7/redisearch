package redisearch

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

func ExampleRediSearch_CreateIndex() {
	search := New(&redis.Options{
		Network:    "tcp",
		Addr:       "redisAddress",
		Password:   "redisPassword",
		DB:         0,
		MaxRetries: 5,
	})

	ctx := context.Background()
	err := search.CreateIndex(ctx, IndexOptions{
		IndexName: "cities",
		Prefix:    []string{"city:"},
		Schema: map[string]FieldSchema{
			"name": {
				Type: FieldTypeText,
				Options: []SchemaOpt{
					SchemaOptWeight(2.0),
				},
			},
			"tags": {
				Type: FieldTypeTag,
				Options: []SchemaOpt{
					SchemaOptTagSeparator(','),
				},
			},
			"population": {
				Type: FieldTypeNumeric,
				Options: []SchemaOpt{
					SchemaOptSortable(),
				},
			},
		},
	}, true) // warning, if 2nd argument is TRUE, all data matching the index prefix will be deleted from redis
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	println("index created")
}

func ExampleRediSearch_Search() {
	search := New(&redis.Options{
		Network:    "tcp",
		Addr:       "redisAddress",
		Password:   "redisPassword",
		DB:         0,
		MaxRetries: 5,
	})

	ctx := context.Background()
	var out []struct {
		Name       string `json:"name"`
		Tags       string `json:"tags"`
		Population int    `json:"population"`
	}
	res, err := search.Search(ctx, SearchOptions{
		IndexName: "cities",
		Query:     "Popayan",
	}, &out)
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	fmt.Printf("search results: %+v", res)
}

func ExampleRediSearch_Search_map() {
	search := New(&redis.Options{
		Network:    "tcp",
		Addr:       "redisAddress",
		Password:   "redisPassword",
		DB:         0,
		MaxRetries: 5,
	})

	ctx := context.Background()
	var out []map[string]string
	res, err := search.Search(ctx, SearchOptions{
		IndexName: "cities",
		Query:     "Popayan",
	}, &out)
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	fmt.Printf("search results: %+v", res)
}

func ExampleRediSearch_Add() {
	search := New(&redis.Options{
		Network:    "tcp",
		Addr:       "redisAddress",
		Password:   "redisPassword",
		DB:         0,
		MaxRetries: 5,
	})

	ctx := context.Background()
	// warning: if 3rd argument (override) is true, existing key/val will be deleted BEFORE writing new value to redis
	err := search.Add(ctx, "", map[string]interface{}{
		"name":       "Popayan",
		"tags":       "colombia,cauca",
		"population": 320000,
	}, false)
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	println("value successfully added")
}

func ExampleRediSearch_IndexExists() {
	search := New(&redis.Options{
		Network:    "tcp",
		Addr:       "redisAddress",
		Password:   "redisPassword",
		DB:         0,
		MaxRetries: 5,
	})

	ctx := context.Background()
	exists, err := search.IndexExists(ctx, "cities")
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	println("exists: ", exists)
}

func ExampleRediSearch_DropIndex() {
	search := New(&redis.Options{
		Network:    "tcp",
		Addr:       "redisAddress",
		Password:   "redisPassword",
		DB:         0,
		MaxRetries: 5,
	})

	ctx := context.Background()
	// warning, if 2nd argument is TRUE, all data matching the index prefix will be deleted from redis
	err := search.DropIndex(ctx, "cities", false)
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	println("index deleted")
}
