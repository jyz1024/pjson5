package pjson5

import (
	"log"
	"testing"
)

var rawJson = `{ // 首行注释
  "number_key": 2,// 人数
  "string_key": /*key中注释*/"www.com",// 字符串类型后注释
  "array_key": [1, 2, 3, 4], // 数组类型
  // 字典类型行注释
  "map_key": {
    // 字典类型首行注释
    "name": "This is name", // 字典字符串
	"val": 60000, // val
	// array
    "data_list": [5000],
  },
}// 尾行注释
// 末尾注释
`

func TestParse(t *testing.T) {
	node := New(rawJson)
	err := node.parse().Error()
	if err != nil {
		t.Fatal("parse error:", err.Error())
	}
	t.Log(node.block)
}

func TestPretty(t *testing.T) {
	pretty := New(rawJson).parse().Pretty()
	t.Log(pretty)
}

func TestNode_Exists(t *testing.T) {
	node := New(rawJson)
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_exist_ex_key",
			args: args{
				path: "map_key",
			},
		},
		{
			name: "test_exist_ne_key",
			args: args{
				path: "level_0_key",
			},
		},
		{
			name: "test_exist_ex_object_key",
			args: args{
				path: "map_key.val",
			},
		},
		{
			name: "test_exist_ne_object_key",
			args: args{
				path: "map_key.ne_key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := node.Exists(tt.args.path)
			log.Println(got)
		})
	}
}

func TestNode_Get(t *testing.T) {
	node := New(rawJson)
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_get_ex_key",
			args: args{
				path: "map_key",
			},
		},
		{
			name: "test_get_ne_key",
			args: args{
				path: "level_0_key",
			},
		},
		{
			name: "test_get_ex_object_key",
			args: args{
				path: "map_key.val",
			},
		},
		{
			name: "test_get_ne_object_key",
			args: args{
				path: "map_key.ne_key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := node.Get(tt.args.path)
			log.Println(got.Pretty())
		})
	}
}

func TestNode_Set(t *testing.T) {
	node := New(rawJson)
	type args struct {
		path string
		val  any
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_set_ne_map",
			args: args{
				path: "level_0_key",
				val: map[string]interface{}{
					"key1": 1,
					"key2": "abc",
					"key3": true,
					"key4": map[string]interface{}{
						"sub_key_1": 3.16,
					},
					"key5": []int{1, 2, 3},
				},
			},
		},
		{
			name: "test_change_ex_key_same_type",
			args: args{
				path: "string_key",
				val:  "new_string_key_val",
			},
		},
		{
			name: "test_set_ne_key",
			args: args{
				path: "string_key_new",
				val:  []string{"new_string_key_val"},
			},
		},
		{
			name: "test_set_object_key_other_type",
			args: args{
				path: "map_key.val",
				val:  []string{"12345"},
			},
		},
		{
			name: "test_set_object_ne_key",
			args: args{
				path: "map_key.new",
				val:  1000,
			},
		},
		{
			name: "test_change_object_ex_key",
			args: args{
				path: "map_key.val",
				val:  -1,
			},
		},
		{
			name: "test_set_all",
			args: args{
				path: "$",
				val:  struct{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := node.Set(tt.args.path, tt.args.val)
			if node.Error() != nil {
				t.Error("set level_0_key error:", node.Error())
				return
			}
			log.Println(got.Pretty())
		})
	}
}

func TestNode_SetMulti(t *testing.T) {
	node := New(rawJson)
	type args struct {
		path string
		val  any
	}
	tests := []struct {
		name string
		args []args
	}{
		{
			name: "test_set_ne_key",
			args: []args{
				{
					path: "new_filed_1",
					val:  []string{"filed_1_val"},
				},
				{
					path: "new_filed_2",
					val:  []string{"filed_2_val"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, arg := range tt.args {
				node.Set(arg.path, arg.val)
				if node.Error() != nil {
					t.Error("set level_0_key error:", node.Error())
					return
				}
			}
			log.Println(node.Pretty())
		})
	}
}

func TestNode_Delete(t *testing.T) {
	node := New(rawJson)
	type args struct {
		path string
		val  any
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_delete_object_key",
			args: args{
				path: "map_key.val",
			},
		},
		{
			name: "test_delete_object_ne_key",
			args: args{
				path: "map_key.ne_key",
			},
		},
		{
			name: "test_delete_normal",
			args: args{
				path: "number_key",
			},
		},
		{
			name: "test_delete_object",
			args: args{
				path: "map_key",
			},
		},
		{
			name: "test_delete_not_exit",
			args: args{
				path: "ne_key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := node.Delete(tt.args.path)
			if node.Error() != nil {
				t.Error("set level_0_key error:", node.Error())
				return
			}
			log.Println(got.Pretty())
		})
	}
}
