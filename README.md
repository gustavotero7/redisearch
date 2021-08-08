# Redisearch go client

### Init client
```golang
search := redisearch.New(&redis.Options{
    Network:    "tcp",
    Addr:       "redisAddress",
    Password:   "redisPassword",
    DB:         0,
    MaxRetries: 5,
})
```
### Create Index
```golang
err := search.CreateIndex(context.Background(), redisearch.IndexOptions{
    IndexName: "cities",
    Prefix:    []string{"city:"},
    Schema: map[string]redisearch.FieldSchema{
        "name": {
            Type: redisearch.FieldTypeText,
            Options: []redisearch.SchemaOpt{
                redisearch.SchemaOptWeight(2.0),
            },
        },
        "tags": {
            Type: redisearch.FieldTypeTag,
            Options: []redisearch.SchemaOpt{
                redisearch.SchemaOptTagSeparator(','),
            },
        },
        "population": {
            Type: redisearch.FieldTypeNumeric,
            Options: []redisearch.SchemaOpt{
                redisearch.SchemaOptSortable(),
            },
        },
    },
}, true) // warning, if 2nd argument is TRUE, all data matching the index prefix will be deleted from redis
if err != nil {
    println("got error: ", err.Error())
    return
}
println("index created")
```
### Add item to index
```golang
// warning: if 3rd argument (override) is true, existing key/val will be deleted BEFORE writing new value to redis
err = search.Add(ctx, "city:popayan", map[string]interface{}{
    "name":       "Popayan",
    "tags":       "colombia,cauca",
    "population": 320000,
}, false)
if err != nil {
    println("got error: ", err.Error())
    return
}
println("value successfully added")

// Note: This is a wrapper of HSET command, and it's usage is optional
```
### Search
```golang
// Search results can be parsed in a list of structs or maps
// When using structs,the field names to parse from redis are taken from json tags (if set)
// by default field names will be taken from exported struct fields
var out []struct {
    Name       string `json:"name"`
    Tags       string `json:"tags"`
    Population int    `json:"population"`
}
res, err := search.Search(ctx, redisearch.SearchOptions{
    IndexName: "cities",
    Query:     "Popayan",
}, &out)
if err != nil {
    println("got error: ", err.Error())
    return
}
fmt.Printf("search results: %+v", res)


// Search storing results in map list
var outMap []map[string]string
res, err = search.Search(ctx, redisearch.SearchOptions{
    IndexName: "cities",
    Query:     "Popayan",
}, &outMap)
if err != nil {
    println("got error: ", err.Error())
    return
}
fmt.Printf("search results: %+v", res)
```
### Drop index
```golang
// Remove the given index from redisearch.
// if purgeIndexData (last argument) is true, all the data(sets) related to the index will be deleted from redis
err = search.DropIndex(ctx, "cities", false)
if err != nil {
    println("got error: ", err.Error())
    return
}
println("index deleted")
```
### Full example
```golang
package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gustavotero7/redisearch"
)


func main() {
	// Init client
	search := redisearch.New(&redis.Options{
		Network:    "tcp",
		Addr:       "redisAddress",
		Password:   "redisPassword",
		DB:         0,
		MaxRetries: 5,
	})

	ctx := context.Background()
	err := search.CreateIndex(ctx, redisearch.IndexOptions{
		IndexName: "cities",
		Prefix:    []string{"city:"},
		Schema: map[string]redisearch.FieldSchema{
			"name": {
				Type: redisearch.FieldTypeText,
				Options: []redisearch.SchemaOpt{
					redisearch.SchemaOptWeight(2.0),
				},
			},
			"tags": {
				Type: redisearch.FieldTypeTag,
				Options: []redisearch.SchemaOpt{
					redisearch.SchemaOptTagSeparator(','),
				},
			},
			"population": {
				Type: redisearch.FieldTypeNumeric,
				Options: []redisearch.SchemaOpt{
					redisearch.SchemaOptSortable(),
				},
			},
		},
	}, true) // warning, if 2nd argument is TRUE, all data matching the index prefix will be deleted from redis
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	println("index created")


	// Add new item to redis
	// warning: if 3rd argument (override) is true, existing key/val will be deleted BEFORE writing new value to redis
	err = search.Add(ctx, "city:popayan", map[string]interface{}{
		"name":       "Popayan",
		"tags":       "colombia,cauca",
		"population": 320000,
	}, false)
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	println("value successfully added")


	// Search storing results in struct list
	var out []struct {
		Name       string `json:"name"`
		Tags       string `json:"tags"`
		Population int    `json:"population"`
	}
	res, err := search.Search(ctx, redisearch.SearchOptions{
		IndexName: "cities",
		Query:     "Popayan",
	}, &out)
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	fmt.Printf("search results: %+v", res)


	// Search storing results in map list
	var outMap []map[string]string
	res, err = search.Search(ctx, redisearch.SearchOptions{
		IndexName: "cities",
		Query:     "Popayan",
	}, &outMap)
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	fmt.Printf("search results: %+v", res)

	// Drop existing index
	// warning, if 2nd argument is TRUE, all data matching the index prefix will be deleted from redis
	err = search.DropIndex(ctx, "cities", false)
	if err != nil {
		println("got error: ", err.Error())
		return
	}
	println("index deleted")
}
```