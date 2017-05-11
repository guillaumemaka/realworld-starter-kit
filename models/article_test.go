package models

import (
	"net/url"
	"reflect"
	"testing"
)

type args struct {
	field string
	vars  []interface{}
}

type test struct {
	name  string
	args  args
	want  string
	want1 interface{}
	want2 bool
}

func Test_buildWhereClause(t *testing.T) {
	emptyQueryParams := url.Values{}
	queryParams := url.Values{}
	queryParams.Set("query", "value")

	tests := []test{
		test{
			name:  "string value",
			args:  args{field: "table.field", vars: []interface{}{"value"}},
			want:  "table.field = ?",
			want1: "value",
			want2: false,
		},
		test{
			name:  "[]string value",
			args:  args{field: "table.field", vars: []interface{}{"value", []string{"value1", "value2"}}},
			want:  "table.field IN (?)",
			want1: []string{"value1", "value2"},
			want2: false,
		},
		test{
			name:  "[]int value",
			args:  args{field: "table.field", vars: []interface{}{[]int{1, 2}}},
			want:  "table.field IN (?)",
			want1: []int{1, 2},
			want2: false,
		},
		test{
			name:  "url.Values value",
			args:  args{field: "table.field", vars: []interface{}{"query", queryParams}},
			want:  "table.field = ?",
			want1: "value",
			want2: false,
		},
		// Empty
		test{
			name:  "empty string value",
			args:  args{field: "table.field", vars: []interface{}{""}},
			want:  "",
			want1: nil,
			want2: true,
		},
		test{
			name:  "empty []string value",
			args:  args{field: "table.field", vars: []interface{}{[]string{}}},
			want:  "",
			want1: nil,
			want2: true,
		},
		test{
			name:  "empty []int value",
			args:  args{field: "table.field", vars: []interface{}{[]int{}}},
			want:  "",
			want1: nil,
			want2: true,
		},
		test{
			name:  "empty url.Values value",
			args:  args{field: "table.field", vars: []interface{}{"query", emptyQueryParams}},
			want:  "",
			want1: nil,
			want2: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := buildWhereClause(tt.args.field, tt.args.vars...)
			if got != tt.want {
				t.Errorf("buildWhereClause() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("buildWhereClause() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("buildWhereClause() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
