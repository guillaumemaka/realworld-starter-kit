package models

import (
	"reflect"
	"testing"
)

func TestNewListOptions(t *testing.T) {
	t.SkipNow()
	// somesetup of test cases
	empty := map[string]interface{}{}
	singleTag := map[string][]string{"tag": []string{"AngularJS"}}
	multiTag := map[string][]string{"tag": []string{"AngularJS", "Testing", "development"}}
	oneOfEach := map[string][]string{"author": []string{"ChilledOJ"}, "tag": []string{"AngularJS"}, "favorite": []string{"johnjacob"}}
	vSpecific := map[string][]string{"author": []string{"ChilledOJ", "johnjacob"}, "tag": []string{"dragons", "training"}, "favorite": []string{"tester1", "tester2"}}
	tests := []struct {
		name string
		args map[string]interface{}
		want ListArticleOptions
	}{
		{"Defaults",
			empty,
			ListArticleOptions{Limit: 20, Offset: 0, Filters: map[string][]string{}},
		},
		{"AllParamsPassedIn",
			map[string]interface{}{"limit": 30, "offset": 10, "filters": singleTag},
			ListArticleOptions{Limit: 30, Offset: 10, Filters: singleTag},
		},
		{"MultipleTagFilter",
			map[string]interface{}{"filters": multiTag},
			ListArticleOptions{Limit: 20, Offset: 0, Filters: multiTag},
		},
		{"OneOfEachFilter",
			map[string]interface{}{"filters": oneOfEach},
			ListArticleOptions{Limit: 20, Offset: 0, Filters: oneOfEach},
		},
		{"VerySpecificFilter",
			map[string]interface{}{"filters": vSpecific},
			ListArticleOptions{Limit: 20, Offset: 0, Filters: vSpecific},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewListOptions(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewListOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}
