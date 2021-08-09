package redisearch

import (
	stdContext "context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// Client hold basic methods to interact with redisearch module for redis
type Client interface {
	Search(ctx stdContext.Context, opts SearchOptions, out interface{}) (int64, error)
	CreateIndex(ctx stdContext.Context, opts IndexOptions, dropIfExists bool) error
	DropIndex(ctx stdContext.Context, name string, purgeIndexData bool) error
	IndexExists(ctx stdContext.Context, name string) (bool, error)
	Add(ctx stdContext.Context, key string, value interface{}, override bool) error
}

var supportedDataTypes = map[reflect.Kind]struct{}{
	reflect.String:  {},
	reflect.Bool:    {},
	reflect.Int:     {},
	reflect.Int8:    {},
	reflect.Int16:   {},
	reflect.Int32:   {},
	reflect.Int64:   {},
	reflect.Uint:    {},
	reflect.Uint8:   {},
	reflect.Uint16:  {},
	reflect.Uint32:  {},
	reflect.Uint64:  {},
	reflect.Uintptr: {},
	reflect.Float32: {},
	reflect.Float64: {},
}

// RediSearch implements Client
type RediSearch struct {
	client *redis.Client
}

// New return a new redisearch implementation instance
func New(opts *redis.Options) Client {
	client := redis.NewClient(opts)
	return &RediSearch{client: client}
}

// Add simple wrapper for redis.HSet. This is just a utility, you can still use the data added using HSET command
// key: set key
// value: map ([string]string or [string]interface{}) or struct to be stored in the set
// override: Delete precious set to create a fresh one only with the values provided
func (r *RediSearch) Add(ctx stdContext.Context, key string, value interface{}, override bool) error {
	if key == "" || value == nil {
		return errors.New("invalid key or nil value")
	}

	if override {
		if err := r.client.Del(ctx, key).Err(); err != nil {
			return err
		}
	}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Invalid {
		return errors.New("invalid {value} type")
	}
	if val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() != reflect.Map && val.Kind() != reflect.Struct {
		return errors.New("{values} arg must be of type map or struct")
	}

	var values []interface{}
	for i := 0; i < val.NumField(); i++ {
		if _, ok := supportedDataTypes[val.Field(i).Type().Kind()]; !ok {
			continue // ignore unsupported type
		}
		k := val.Type().Field(i).Name
		v := val.Field(i).Interface()
		values = append(values, k, v)

	}
	return r.client.HSet(ctx, key, values...).Err()
}

// Search the index with a textual query
func (r *RediSearch) Search(ctx stdContext.Context, opts SearchOptions, out interface{}) (int64, error) {
	if opts.IndexName == "" || opts.Query == "" {
		return 0, errors.New("missing required (IndexName or Query)")
	}
	args := []interface{}{
		"FT.SEARCH",
		opts.IndexName,
		opts.Query,
	}

	for _, flag := range opts.Flags {
		args = append(args, flag)
	}

	for _, filter := range opts.Filters {
		args = append(args,
			"FILTER",
			filter.NumericFieldName,
			filter.Min,
			filter.Max,
		)
	}
	if opts.GeoFilter != nil {
		if opts.GeoFilter.Unit == "" {
			opts.GeoFilter.Unit = "m"
		}
		args = append(args,
			"GEOFILTER",
			opts.GeoFilter.GeoFieldName,
			opts.GeoFilter.Longitude,
			opts.GeoFilter.Latitude,
			opts.GeoFilter.Radius,
			opts.GeoFilter.Unit,
		)
	}

	if fields := opts.InKeys; len(fields) != 0 {

		args = append(args,
			"INKEYS",
			len(fields),
		)
		for _, key := range fields {
			args = append(args, key)
		}
	}

	if fields := opts.InFields; len(fields) != 0 {
		args = append(args,
			"INFIELDS",
			len(fields),
		)
		for _, field := range fields {
			args = append(args, field)
		}
	}

	if fields := opts.Return; len(fields) != 0 {
		args = append(args,
			"RETURN",
			len(fields),
		)
		for _, field := range fields {
			args = append(args, field)
		}
	}

	if opts.Summarize != nil {
		args = append(args, "SUMMARIZE")
		if fields := opts.Summarize.Fields; len(fields) != 0 {
			args = append(args,
				"FIELDS",
				len(fields),
			)
			for _, field := range fields {
				args = append(args, field)
			}
		}
		if opts.Summarize.Fragments != 0 {
			args = append(args,
				"FRAGS",
				opts.Summarize.Fragments,
			)
		}
		if opts.Summarize.Length != 0 {
			args = append(args,
				"LEN",
				opts.Summarize.Length,
			)
		}
		if opts.Summarize.Separator != "" {
			args = append(args,
				"SEPARATOR",
				opts.Summarize.Separator,
			)
		}
	}

	if opts.Highlight != nil {
		args = append(args, "HIGHLIGHT")
		if fields := opts.Highlight.Fields; len(fields) != 0 {
			args = append(args,
				"FIELDS",
				len(fields),
			)
			for _, field := range fields {
				args = append(args, field)
			}
		}
		if opts.Highlight.OpenTag != "" && opts.Highlight.CloseTag != "" {
			args = append(args,
				"TAGS",
				opts.Highlight.OpenTag,
				opts.Highlight.CloseTag,
			)
		}
	}

	if opts.Slop != nil {
		args = append(args,
			"SLOP",
			opts.Slop,
		)
	}

	if opts.Language != "" {
		args = append(args,
			"LANGUAGE",
			opts.Language,
		)
	}

	if opts.Expander != "" {
		args = append(args,
			"EXPANDER",
			opts.Expander,
		)
	}

	if opts.Scorer != "" {
		args = append(args,
			"SCORER",
			opts.Scorer,
		)
	}

	if opts.Payload != "" {
		args = append(args,
			"PAYLOAD",
			opts.Payload,
		)
	}

	if opts.SortBy != nil {
		order := "ASC"
		if opts.SortBy.Descending {
			order = "DESC"
		}
		args = append(args,
			"SORTBY",
			opts.SortBy.FieldName,
			order,
		)
	}

	if opts.Limit != nil {
		args = append(args,
			"LIMIT",
			opts.Limit.Offset,
			opts.Limit.Max,
		)
	}

	do := r.client.Do(ctx, args...)
	res, err := do.Result()
	if err != nil {
		println("SEARCH ERR: ", err.Error())
		return 0, err
	}
	fmt.Printf("SEARCH RES: ***%+v*** %T\n", res, res)
	return parseSearchResults(res, out)
}

// CreateIndex with the given spec
func (r *RediSearch) CreateIndex(ctx stdContext.Context, opts IndexOptions, dropIfExists bool) error {
	exists, err := r.IndexExists(ctx, opts.IndexName)
	if err != nil {
		return err
	}
	if exists && dropIfExists {
		if err := r.DropIndex(ctx, opts.IndexName, false); err != nil {
			return err
		}
	} else if exists {
		return errors.New("index already exists")
	}

	args := []interface{}{
		"FT.CREATE",
		opts.IndexName,
		"ON",
		"HASH",
	}
	if pLen := len(opts.Prefix); pLen != 0 {
		p := make([]interface{}, pLen+2)
		p[0] = "PREFIX"
		p[1] = pLen
		for i := 0; i < pLen; i++ {
			p[i+2] = opts.Prefix[i]
		}
		args = append(args, p...)
	}
	if opts.Filter != "" {
		args = append(args, "FILTER", opts.Filter)
	}
	if opts.Language != "" {
		args = append(args, "LANGUAGE", opts.Language)
	}
	if opts.LanguageField != "" {
		args = append(args, "LANGUAGE_FIELD", opts.LanguageField)
	}
	if opts.Score > 0 {
		args = append(args, "SCORE", opts.Score)
	}
	if opts.ScoreField != "" {
		args = append(args, "SCORE_FIELD", opts.ScoreField)
	}
	if opts.PayloadField != "" {
		args = append(args, "PAYLOAD_FIELD", opts.PayloadField)
	}
	if opts.Temporary > 0 {
		args = append(args, "TEMPORARY", opts.Temporary)
	}
	if swLen := len(opts.StopWords); swLen != 0 {
		sw := make([]interface{}, swLen+2)
		sw[0] = "STOPWORDS"
		sw[1] = swLen
		for i := 0; i < swLen; i++ {
			sw[i+2] = opts.StopWords[i]
		}
		args = append(args, sw...)
	}
	if fLen := len(opts.Flags); fLen != 0 {
		flags := make([]interface{}, fLen)
		for i := 0; i < fLen; i++ {
			flags[i] = opts.Flags[i]
		}
		args = append(args, flags...)
	}
	if len(opts.Schema) != 0 {
		args = append(args, "SCHEMA")
		for field, schema := range opts.Schema {
			args = append(args, field, schema.Type)
			for _, option := range schema.Options {
				args = append(args, option...)
			}
		}
	}
	fmt.Println(args...)
	do := r.client.Do(ctx, args...)
	if _, err := do.Result(); err != nil {
		return err
	}
	return nil
}

// DropIndex with the given name. Optionally delete all indexed data
func (r *RediSearch) DropIndex(ctx stdContext.Context, name string, purgeIndexData bool) error {
	args := []interface{}{
		"FT.DROPINDEX",
		name,
	}
	if purgeIndexData {
		args = append(args, "DD")
	}
	do := r.client.Do(ctx, args...)
	_, err := do.Result()
	return err
}

// IndexExists return true if the index exists
func (r *RediSearch) IndexExists(ctx stdContext.Context, name string) (bool, error) {
	args := []interface{}{
		"FT.INFO",
		name,
	}
	do := r.client.Do(ctx, args...)
	_, err := do.Result()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unknown index") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// parseSearchResults into the given list of structs or maps
func parseSearchResults(raw interface{}, out interface{}) (int64, error) {
	v := reflect.ValueOf(out)
	if v.Kind() == reflect.Invalid {
		return 0, errors.New("invalid {out} type")
	}
	if v.Kind() != reflect.Ptr {
		return 0, errors.New("{out} arg must be a pointer")
	}
	if !v.Elem().CanSet() {
		return 0, errors.New("using unaddressable value")
	}

	if v.Elem().Kind() != reflect.Slice {
		return 0, errors.New("{out} arg must reference a slice")
	}

	// [totalHits hashKey1 [field1 value1 ...] hashKey2 [field1 value1 ...] ....]
	resSlice, ok := raw.([]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid redis response type: %T", raw)
	}
	if len(resSlice) < 3 { // no results
		return 0, nil
	}

	total, ok := resSlice[0].(int64)
	if !ok {
		return 0, fmt.Errorf("invalid redis response.total type: %T", resSlice[0])
	}

	parsedMaps := make([]map[string]string, len(resSlice[1:])/2)
	for i := 1; i < len(resSlice); i += 2 {
		docSlice, ok := resSlice[i+1].([]interface{})
		if !ok {
			return 0, fmt.Errorf("invalid redis response hash value type: %T", resSlice[i+1])
		}

		m := make(map[string]string, len(docSlice)/2)
		for j := 0; j < len(docSlice); j += 2 {
			field, value := docSlice[j].(string), docSlice[j+1].(string)
			m[field] = value
		}

		ii := i - 1
		if ii != 0 {
			ii /= 2
		}
		parsedMaps[ii] = m
	}

	// avoid type parsing if {out} is the same type of parsed maps
	if v.Elem().Type().AssignableTo(reflect.TypeOf(parsedMaps)) {
		v.Elem().Set(reflect.ValueOf(parsedMaps))
		return total, nil
	}

	switch t := v.Elem().Type().Elem(); t.Kind() {
	case reflect.Struct:
		fieldsIndexMap := make(map[string]int, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			key := t.Field(i).Name
			if tag := t.Field(i).Tag.Get("json"); tag != "" {
				tagSlice := strings.Split(tag, ",")
				if tagSlice[0] != "" {
					key = tagSlice[0]
				}

			}
			if key != "" && key != "-" {
				fieldsIndexMap[key] = i
			}
		}

		e := initValueAndGetElem(v, len(parsedMaps))
		for i, m := range parsedMaps {
			sliceItemToSet := e.Index(i)
			for k, v := range m {
				fieldIndex, ok := fieldsIndexMap[k]
				if !ok {
					continue
				}
				field := sliceItemToSet.Field(fieldIndex)
				if _, ok := supportedDataTypes[field.Kind()]; !ok {
					log.Println(k + ":" + field.Kind().String() + " type is not supported and will be ignored")
					continue // ignore unsupported type
				}
				switch kind := field.Kind(); {
				case kind == reflect.String:
					field.SetString(v)
				case kind >= reflect.Int && kind <= reflect.Uint64:
					vInt, _ := strconv.ParseInt(v, 10, 64)
					field.Set(reflect.ValueOf(vInt).Convert(field.Type()))
				case kind >= reflect.Float32 && kind <= reflect.Float64:
					vFloat, _ := strconv.ParseFloat(v, 64)
					field.Set(reflect.ValueOf(vFloat).Convert(field.Type()))
				case kind == reflect.Bool:
					vBool := v[0] == 't' || v[0] == '1'
					field.SetBool(vBool)
				}
			}
		}
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return 0, errors.New("map key must be of type string")
		}
		if t.Elem().Kind() != reflect.String && t.Elem().Kind() != reflect.Interface {
			return 0, errors.New("map value type must be of type string or interface{}")
		}

		e := initValueAndGetElem(v, len(parsedMaps))
		for i, m := range parsedMaps {
			sliceItemToSet := e.Index(i)
			if sliceItemToSet.IsNil() {
				sliceItemToSet.Set(reflect.MakeMap(sliceItemToSet.Type()))
			}
			for k, v := range m {
				sliceItemToSet.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
			}
		}
	default:
		return 0, errors.New("{out} must be and slice of structs, interface{} or string map")
	}
	return total, nil
}

// init v of kind slice
func initValueAndGetElem(v reflect.Value, size int) reflect.Value {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	e := v.Elem()
	e.Set(reflect.MakeSlice(e.Type(), size, size))
	return e
}
