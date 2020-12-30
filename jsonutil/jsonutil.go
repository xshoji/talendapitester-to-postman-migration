package jsonutil

import (
	"bytes"
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

// interface{}型のjsonオブジェクトからキー指定で値を取り出す（object["aaa"][0]["bbb"] -> keyChain: "aaa.0.bbb"）
func Get(object interface{}, keyChain string) interface{} {
	keys := strings.Split(keyChain, ".")
	var result interface{}
	var exists bool
	for _, key := range keys {
		exists = false
		value, ok := object.(map[string]interface{})
		if ok {
			exists = true
			object = value[key]
			result = object
			continue
		}
		values, ok := object.([]interface{})
		if ok {
			for i, v := range values {
				if strconv.FormatInt(int64(i), 10) == key {
					exists = true
					object = v
					result = object
					continue
				}
			}
		}
	}
	if exists {
		return result
	}
	return nil
}

func AsString(object interface{}, keyChain string) string {
	maybeString := Get(object, keyChain)
	if stringValue, ok := maybeString.(string); ok {
		return stringValue
	} else {
		log.Printf("jsonutil.AsString is failed. [ keyCain: %v, maybeString: %v ]\n", keyChain, maybeString)
		return ""
	}
}

func AsSlice(object interface{}, keyChain string) []interface{} {
	maybeSlice := Get(object, keyChain)

	if slice, ok := maybeSlice.([]interface{}); ok {
		return slice
	} else {
		log.Printf("jsonutil.AsSlice is failed. [ keyCain: %v, maybeSlice: %v ]\n", keyChain, maybeSlice)
		return []interface{}{}
	}
}

// 値をjson形式の文字列に変換する
func ToJsonString(v interface{}) string {
	result, _ := json.Marshal(v)
	return string(result)
}

func ToJsonStringPretty(v interface{}) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(ToJsonString(v)), "", "  "); err != nil {
		panic(err)
	}
	return buf.String()
}
