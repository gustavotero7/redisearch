package redisearch

import (
	"reflect"
	"testing"
)

func Test_parseSearchResults(t *testing.T) {
	type TestModel struct {
		Title       string  `json:"title"`
		Year        int     `json:"year"`
		Active      bool    `json:"active"`
		Score       float32 `json:"score"`
		Unsupported []int   `json:"unsupported"`
	}
	type args struct {
		raw interface{}
		out interface{}
	}
	type want struct {
		total int64
		out   interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "ok: 1 result, using struct for {out} arg",
			args: args{
				raw: []interface{}{
					int64(1),
					"doc:id",
					[]interface{}{
						"title",
						"Test Title",
						"year",
						"2021",
						"active",
						"true",
						"score",
						"12.8",
						"unsupported",
						"0",
						"notinmodel",
						"0",
					},
				},
				out: &[]TestModel{},
			},
			want: want{
				total: 1,
				out: &[]TestModel{
					{
						Title:  "Test Title",
						Year:   2021,
						Active: true,
						Score:  12.8,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ok: 1 result, using strings map for {out} arg",
			args: args{
				raw: []interface{}{
					int64(1),
					"doc:id",
					[]interface{}{
						"title",
						"Test Title",
						"year",
						"2021",
						"active",
						"true",
					},
				},
				out: &[]map[string]string{},
			},
			want: want{
				total: 1,
				out: &[]map[string]string{
					{
						"title":  "Test Title",
						"year":   "2021",
						"active": "true",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ok: 1 result, using map[string]interface{} for {out} arg",
			args: args{
				raw: []interface{}{
					int64(1),
					"doc:id",
					[]interface{}{
						"title",
						"Test Title",
						"year",
						"2021",
						"active",
						"true",
					},
				},
				out: &[]map[string]interface{}{},
			},
			want: want{
				total: 1,
				out: &[]map[string]interface{}{
					{
						"title":  "Test Title",
						"year":   "2021",
						"active": "true",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid output type (not an slice)",
			args: args{
				raw: []interface{}{
					int64(1),
					"doc:id",
					[]interface{}{
						"title",
						"Test Title",
						"year",
						"2021",
						"active",
						"true",
					},
				},
				out: &TestModel{},
			},
			want: want{
				total: 0,
				out:   &TestModel{},
			},
			wantErr: true,
		},
		{
			name: "invalid output type (map with invalid value type)",
			args: args{
				raw: []interface{}{
					int64(1),
					"doc:id",
					[]interface{}{
						"title",
						"Test Title",
						"year",
						"2021",
						"active",
						"true",
					},
				},
				out: &[]map[string]int64{},
			},
			want: want{
				total: 0,
				out:   &[]map[string]int64{},
			},
			wantErr: true,
		},
		{
			name: "invalid output type (nil)",
			args: args{
				raw: []interface{}{
					int64(1),
					"doc:id",
					[]interface{}{
						"title",
						"Test Title",
						"year",
						"2021",
						"active",
						"true",
					},
				},
				out: nil,
			},
			want: want{
				total: 0,
				out:   nil,
			},
			wantErr: true,
		},
		{
			name: "{out} is not a pointer",
			args: args{
				raw: []interface{}{
					int64(1),
					"doc:id",
					[]interface{}{
						"title",
						"Test Title",
						"year",
						"2021",
						"active",
						"true",
					},
				},
				out: []TestModel{},
			},
			want: want{
				total: 0,
				out:   []TestModel{},
			},
			wantErr: true,
		},
		{
			name: "{out} have unaddressable value",
			args: args{
				raw: []interface{}{
					int64(1),
					"doc:id",
					[]interface{}{
						"title",
						"Test Title",
						"year",
						"2021",
						"active",
						"true",
					},
				},
				out: func() interface{} {
					var t *[]TestModel
					return t
				}(),
			},
			want: want{
				total: 0,
				out: func() interface{} {
					var t *[]TestModel
					return t
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSearchResults(tt.args.raw, tt.args.out)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSearchResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want.total {
				t.Errorf("parseSearchResults() got total = %v, want %v", got, tt.want.total)
			}

			if !reflect.DeepEqual(tt.args.out, tt.want.out) {
				t.Errorf("parseSearchResults() out = %+v, want %+v", tt.args.out, tt.want.out)
			}
		})
	}
}

func Benchmark_parseSearchResults_map(b *testing.B) {
	b.ReportAllocs()
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		raw := []interface{}{
			int64(1),
			"doc:id",
			[]interface{}{
				"title",
				"Test Title",
				"year",
				"2021",
				"active",
				"true",
				"score",
				"12.8",
				"unsupported",
				"0",
				"notinmodel",
				"0",
			},
		}
		var out []map[string]interface{}
		_, err := parseSearchResults(raw, &out)
		if err != nil {
			b.Error(err)
		}

	}
}

func Benchmark_parseSearchResults_struct(b *testing.B) {
	b.ReportAllocs()
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		raw := []interface{}{
			int64(1),
			"doc:id",
			[]interface{}{
				"title",
				"Test Title",
				"year",
				"2021",
				"active",
				"true",
				"score",
				"12.8",
				"unsupported",
				"0",
				"notinmodel",
				"0",
			},
		}
		var out []struct{
			Title string`json:"title"`
			Year int `json:"year"`
			Active bool `json:"active"`
			Score float32 `json:"score"`
			Another float32 `json:"another"`
		}
		_, err := parseSearchResults(raw, &out)
		if err != nil {
			b.Error(err)
		}
	}
}