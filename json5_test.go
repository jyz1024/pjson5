package pjson5

import (
	"log"
	"testing"
)

var rawArrayJson = `{
  "nums": [1, 2, 3, 4],//test
  "strs": ["a", "b", "c",],
  // 行注释
  "nested": [[1, 2], [3, 4],/*这是一段注释*/[5,6]],
  "mixed": [1, "two", true, null, {"key": "val"}], /* 块注释 */
  "multiline": [
    10,
	// test
    20,
    30,
  ],
 // 尾部行注释
	/* 尾部块注释 
*/
}`

func TestArray_Parse(t *testing.T) {
	node := New(rawArrayJson)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse error:", err)
	}
	// basic array type check
	arr := node.Get("nums")
	if !arr.IsArray() {
		t.Fatal("expected nums to be Array type")
	}
	if arr.Len() != 4 {
		t.Fatalf("expected len=4, got %d", arr.Len())
	}
	// get by index
	v0 := node.Get("nums.0")
	if v0.Value() != "1" {
		t.Fatalf("expected nums[0]=1, got %q", v0.Value())
	}
	v3 := node.Get("nums.3")
	if v3.Value() != "4" {
		t.Fatalf("expected nums[3]=4, got %q", v3.Value())
	}
	// string array
	sa := node.Get("strs.1")
	if sa.Value() != `"b"` {
		t.Fatalf("expected strs[1]=\"b\", got %q", sa.Value())
	}
	// mixed types including null
	nullElem := node.Get("mixed.3")
	if nullElem.Type() != Null {
		t.Fatalf("expected Null type, got %v", nullElem.Type())
	}
	objElem := node.Get("mixed.4")
	if !objElem.IsObject() {
		t.Fatal("expected mixed[4] to be object")
	}
	nestedVal := node.Get("mixed.4.key")
	if nestedVal.Value() != `"val"` {
		t.Fatalf("expected mixed[4].key=\"val\", got %q", nestedVal.Value())
	}
	// nested arrays
	inner := node.Get("nested.0")
	if !inner.IsArray() {
		t.Fatal("expected nested[0] to be array")
	}
	if inner.Get("1").Value() != "2" {
		t.Fatalf("expected nested[0][1]=2, got %q", inner.Get("1").Value())
	}
	// multiline array
	ml := node.Get("multiline")
	if ml.Len() != 3 {
		t.Fatalf("expected multiline len=3, got %d", ml.Len())
	}
	log.Println("array parse OK")
}

func TestArray_ForEach(t *testing.T) {
	node := New(rawArrayJson)
	var keys []string
	var vals []string
	node.Get("nums").ForEach(func(key string, value *Node) bool {
		keys = append(keys, key)
		vals = append(vals, value.Value())
		return true
	})
	if len(keys) != 4 {
		t.Fatalf("expected 4 iterations, got %d", len(keys))
	}
	log.Printf("ForEach keys=%v vals=%v", keys, vals)
}

func TestArray_Set(t *testing.T) {
	t.Run("update_nums_element", func(t *testing.T) {
		node := New(rawArrayJson)
		node.Set("nums.0", 99)
		if node.Error() != nil {
			t.Fatal("set error:", node.Error())
		}
		if v := node.Get("nums.0").Value(); v != "99" {
			t.Fatalf("expected nums[0]=99, got %q", v)
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("append_nums_element", func(t *testing.T) {
		node := New(rawArrayJson)
		node.Set("nums.4", 5)
		if node.Error() != nil {
			t.Fatal("append error:", node.Error())
		}
		if node.Get("nums").Len() != 5 {
			t.Fatalf("expected len=5, got %d", node.Get("nums").Len())
		}
		if v := node.Get("nums.4").Value(); v != "5" {
			t.Fatalf("expected nums[4]=5, got %q", v)
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("update_strs_trailing_comma", func(t *testing.T) {
		// strs has trailing comma: ["a", "b", "c",]
		node := New(rawArrayJson)
		node.Set("strs.1", "B")
		if node.Error() != nil {
			t.Fatal("set error:", node.Error())
		}
		if v := node.Get("strs.1").Value(); v != `"B"` {
			t.Fatalf("expected strs[1]=\"B\", got %q", v)
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("append_strs", func(t *testing.T) {
		node := New(rawArrayJson)
		node.Set("strs.3", "d")
		if node.Error() != nil {
			t.Fatal("append error:", node.Error())
		}
		if node.Get("strs").Len() != 4 {
			t.Fatalf("expected len=4, got %d", node.Get("strs").Len())
		}
		if v := node.Get("strs.3").Value(); v != `"d"` {
			t.Fatalf("expected strs[3]=\"d\", got %q", v)
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("update_nested_subarray", func(t *testing.T) {
		// nested: [[1, 2], [3, 4], /*注释*/ [5,6]]
		node := New(rawArrayJson)
		node.Set("nested.2", []int{7, 8, 9})
		if node.Error() != nil {
			t.Fatal("set error:", node.Error())
		}
		if v := node.Get("nested.2").Value(); v != "[7,8,9]" {
			t.Fatalf("expected nested[2]=[7,8,9], got %q", v)
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("append_nested_subarray", func(t *testing.T) {
		node := New(rawArrayJson)
		node.Set("nested.3", []int{10, 11})
		if node.Error() != nil {
			t.Fatal("append error:", node.Error())
		}
		if node.Get("nested").Len() != 4 {
			t.Fatalf("expected len=4, got %d", node.Get("nested").Len())
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("update_mixed_types", func(t *testing.T) {
		// mixed: [1, "two", true, null, {"key": "val"}]
		node := New(rawArrayJson)
		// number -> string
		node.Set("mixed.0", "one")
		if node.Error() != nil {
			t.Fatal("set mixed[0] error:", node.Error())
		}
		if v := node.Get("mixed.0").Value(); v != `"one"` {
			t.Fatalf("expected mixed[0]=\"one\", got %q", v)
		}
		// bool -> number
		node.Set("mixed.2", 100)
		if node.Error() != nil {
			t.Fatal("set mixed[2] error:", node.Error())
		}
		if v := node.Get("mixed.2").Value(); v != "100" {
			t.Fatalf("expected mixed[2]=100, got %q", v)
		}
		// null -> object
		node.Set("mixed.3", map[string]int{"x": 1})
		if node.Error() != nil {
			t.Fatal("set mixed[3] error:", node.Error())
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("update_nested_object_in_array", func(t *testing.T) {
		// mixed[4] is {"key": "val"}, update the inner key
		node := New(rawArrayJson)
		node.Set("mixed.4.key", "new_val")
		if node.Error() != nil {
			t.Fatal("set error:", node.Error())
		}
		if v := node.Get("mixed.4.key").Value(); v != `"new_val"` {
			t.Fatalf("expected mixed[4].key=\"new_val\", got %q", v)
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("update_multiline_with_comment", func(t *testing.T) {
		// multiline: [10, // test\n 20, 30,]
		node := New(rawArrayJson)
		node.Set("multiline.1", 200)
		if node.Error() != nil {
			t.Fatal("set error:", node.Error())
		}
		if v := node.Get("multiline.1").Value(); v != "200" {
			t.Fatalf("expected multiline[1]=200, got %q", v)
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("append_multiline", func(t *testing.T) {
		node := New(rawArrayJson)
		node.Set("multiline.3", 40)
		if node.Error() != nil {
			t.Fatal("append error:", node.Error())
		}
		if node.Get("multiline").Len() != 4 {
			t.Fatalf("expected len=4, got %d", node.Get("multiline").Len())
		}
		if v := node.Get("multiline.3").Value(); v != "40" {
			t.Fatalf("expected multiline[3]=40, got %q", v)
		}
		log.Println("result:", node.Pretty())
	})

	t.Run("out_of_range_index_error", func(t *testing.T) {
		node := New(rawArrayJson)
		// nums has 4 elements, index 6 is out of range (not append)
		node.Set("nums.6", 100)
		if node.Error() == nil {
			t.Fatal("expected error for out-of-range index, got nil")
		}
		log.Println("expected error:", node.Error())
	})

	t.Run("non_numeric_index_error", func(t *testing.T) {
		node := New(rawArrayJson)
		node.Set("nums.abc", 100)
		if node.Error() == nil {
			t.Fatal("expected error for non-numeric array index, got nil")
		}
		log.Println("expected error:", node.Error())
	})
}

func TestArray_Delete(t *testing.T) {
	node := New(rawArrayJson)
	node.Delete("nums.1")
	if node.Error() != nil {
		t.Fatal("delete error:", node.Error())
	}
	arr := node.Get("nums")
	if arr.Len() != 3 {
		t.Fatalf("expected len=3 after delete, got %d", arr.Len())
	}
	// former index 2 (value 3) should now be at index 1
	if arr.Get("1").Value() != "3" {
		t.Fatalf("expected arr[1]=3 after delete, got %q", arr.Get("1").Value())
	}
	log.Println("delete result:", node.Pretty())
}

func TestArray_Pretty(t *testing.T) {
	node := New(rawArrayJson)
	// parse array explicitly to trigger rebuild
	node.Get("nums").ForEach(func(_ string, _ *Node) bool { return true })
	pretty := node.Pretty()
	log.Println("pretty:", pretty)
}

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

// ==================== JSON5 Feature Tests ====================

// JSON5: unquoted keys
func TestJSON5_UnquotedKeys(t *testing.T) {
	input := `{name: "hello", age: 25}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse unquoted keys error:", err)
	}
	if v := node.Get("name").Value(); v != `"hello"` {
		t.Fatalf("expected name=\"hello\", got %q", v)
	}
	if v := node.Get("age").Value(); v != "25" {
		t.Fatalf("expected age=25, got %q", v)
	}
}

// JSON5: single-quoted strings
func TestJSON5_SingleQuotedString(t *testing.T) {
	input := `{"key": 'hello'}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse single-quoted string error:", err)
	}
	v := node.Get("key").Value()
	if v != "'hello'" {
		t.Fatalf("expected key='hello', got %q", v)
	}
}

// JSON5: single-quoted key
func TestJSON5_SingleQuotedKey(t *testing.T) {
	input := `{'name': "hello"}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse single-quoted key error:", err)
	}
	v := node.Get("name").Value()
	if v != `"hello"` {
		t.Fatalf("expected name=\"hello\", got %q", v)
	}
}

// JSON5: trailing commas (object & array)
func TestJSON5_TrailingCommas(t *testing.T) {
	input := `{
		"a": 1,
		"b": [1, 2, 3,],
	}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse trailing commas error:", err)
	}
	if v := node.Get("a").Value(); v != "1" {
		t.Fatalf("expected a=1, got %q", v)
	}
	arr := node.Get("b")
	if arr.Len() != 3 {
		t.Fatalf("expected b len=3, got %d", arr.Len())
	}
}

// JSON5: hex numbers
func TestJSON5_HexNumber(t *testing.T) {
	input := `{"hex": 0xFF}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse hex number error:", err)
	}
	v := node.Get("hex").Value()
	if v != "0xFF" {
		t.Fatalf("expected hex=0xFF, got %q", v)
	}
	if node.Get("hex").Type() != Number {
		t.Fatalf("expected Number type, got %v", node.Get("hex").Type())
	}
}

// JSON5: Infinity and NaN
func TestJSON5_InfinityAndNaN(t *testing.T) {
	input := `{"inf": Infinity, "ninf": -Infinity, "nan": NaN}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse Infinity/NaN error:", err)
	}
	if v := node.Get("inf").Value(); v != "Infinity" {
		t.Fatalf("expected inf=Infinity, got %q", v)
	}
	if v := node.Get("ninf").Value(); v != "-Infinity" {
		t.Fatalf("expected ninf=-Infinity, got %q", v)
	}
	if v := node.Get("nan").Value(); v != "NaN" {
		t.Fatalf("expected nan=NaN, got %q", v)
	}
}

// JSON5: positive sign on numbers
func TestJSON5_PositiveSign(t *testing.T) {
	input := `{"pos": +1.5}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse positive sign error:", err)
	}
	if v := node.Get("pos").Value(); v != "+1.5" {
		t.Fatalf("expected pos=+1.5, got %q", v)
	}
}

// JSON5: leading decimal point (.5)
func TestJSON5_LeadingDecimalPoint(t *testing.T) {
	input := `{"ld": .5}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse leading decimal error:", err)
	}
	if v := node.Get("ld").Value(); v != ".5" {
		t.Fatalf("expected ld=.5, got %q", v)
	}
}

// JSON5: trailing decimal point (5.)
func TestJSON5_TrailingDecimalPoint(t *testing.T) {
	input := `{"td": 5.}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse trailing decimal error:", err)
	}
	if v := node.Get("td").Value(); v != "5." {
		t.Fatalf("expected td=5., got %q", v)
	}
}

// JSON5: single-line comments
func TestJSON5_SingleLineComment(t *testing.T) {
	input := `{
		// this is a comment
		"key": "val"
	}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse single-line comment error:", err)
	}
	if v := node.Get("key").Value(); v != `"val"` {
		t.Fatalf("expected key=\"val\", got %q", v)
	}
}

// JSON5: block comments
func TestJSON5_BlockComment(t *testing.T) {
	input := `{
		/* block comment */
		"key": /*inline*/ "val"
	}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse block comment error:", err)
	}
	if v := node.Get("key").Value(); v != `"val"` {
		t.Fatalf("expected key=\"val\", got %q", v)
	}
}

// JSON5: comprehensive - mix of JSON5 features
func TestJSON5_Comprehensive(t *testing.T) {
	input := `{
		// JSON5 comprehensive test
		name: "pjson5",
		version: 1,
		"enabled": true,
		items: [
			1,
			"two",
			true,
			null,
		],
	}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse comprehensive JSON5 error:", err)
	}
	if v := node.Get("name").Value(); v != `"pjson5"` {
		t.Fatalf("expected name=\"pjson5\", got %q", v)
	}
	if v := node.Get("version").Value(); v != "1" {
		t.Fatalf("expected version=1, got %q", v)
	}
	if v := node.Get("enabled").Value(); v != "true" {
		t.Fatalf("expected enabled=true, got %q", v)
	}
	arr := node.Get("items")
	if arr.Len() != 4 {
		t.Fatalf("expected items len=4, got %d", arr.Len())
	}
	if arr.Get("3").Type() != Null {
		t.Fatalf("expected items[3] to be null, got %v", arr.Get("3").Type())
	}
	log.Println("comprehensive pretty:", node.Pretty())
}

// JSON5: array with comments inside
func TestJSON5_ArrayWithComments(t *testing.T) {
	input := `{
		"arr": [
			// first element
			1,
			/* second */ 2,
			3, // trailing
		]
	}`
	node := New(input)
	if err := node.Parse().Error(); err != nil {
		t.Fatal("parse array with comments error:", err)
	}
	arr := node.Get("arr")
	if arr.Len() != 3 {
		t.Fatalf("expected len=3, got %d", arr.Len())
	}
	if v := arr.Get("0").Value(); v != "1" {
		t.Fatalf("expected arr[0]=1, got %q", v)
	}
}

// JSON5: Set/Delete on JSON5 with unquoted keys
func TestJSON5_SetDeleteUnquotedKeys(t *testing.T) {
	input := `{name: "hello", age: 25}`
	node := New(input)
	node.Set("name", "world")
	if node.Error() != nil {
		t.Fatal("set error:", node.Error())
	}
	if v := node.Get("name").Value(); v != `"world"` {
		t.Fatalf("expected name=\"world\", got %q", v)
	}
	node.Delete("age")
	if node.Error() != nil {
		t.Fatal("delete error:", node.Error())
	}
	if node.Get("age").IsExist() {
		t.Fatal("expected age to be deleted")
	}
	log.Println("set/delete unquoted:", node.Pretty())
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
