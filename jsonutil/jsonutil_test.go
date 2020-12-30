package jsonutil

import (
	"encoding/json"
	"testing"
)

func TestGet(t *testing.T) {
	t.Run("test Get (single value) ", func(t *testing.T) {

		// setup
		var jsonObject interface{}
		json.Unmarshal([]byte(`
{
  "string": "stringValue1",
  "number": 1,
  "null": null,
  "boolean": true,
  "stringArray": [
    "stringValue2",
    "stringValue3"
  ],
  "numberArray": [
    2,
    3
  ],
  "object": {
    "string": "stringValue4",
    "number": 4
  },
  "objectArray": [
    {
      "string": "stringValue5",
      "number": 5
    },
    {
      "string": "stringValue6",
      "number": 6
    }
  ]
}
`), &jsonObject)

		// test cases
		testCases := []struct {
			TestCase string
			Input    string
			Expected interface{}
		}{
			{
				TestCase: "string",
				Input:    "string",
				Expected: "stringValue1",
			},
			{
				TestCase: "number",
				Input:    "number",
				Expected: 1.0,
			},
			{
				TestCase: "null",
				Input:    "null",
				Expected: nil,
			},
			{
				TestCase: "boolean",
				Input:    "boolean",
				Expected: true,
			},
			{
				TestCase: "stringArray",
				Input:    "stringArray.0",
				Expected: "stringValue2",
			},
			{
				TestCase: "numberArray",
				Input:    "numberArray.0",
				Expected: 2.0,
			},
			{
				TestCase: "object.string",
				Input:    "object.string",
				Expected: "stringValue4",
			},
			{
				TestCase: "object.number",
				Input:    "object.number",
				Expected: 4.0,
			},
			{
				TestCase: "objectArray.string",
				Input:    "objectArray.0.string",
				Expected: "stringValue5",
			},
			{
				TestCase: "objectArray.number",
				Input:    "objectArray.0.number",
				Expected: 5.0,
			},
		}

		// run
		for i := range testCases {
			param := testCases[i]
			actual := Get(jsonObject, param.Input)
			t.Logf("Case:%v\n", param.TestCase)
			if actual != param.Expected {
				t.Errorf("  Failed: actual -> %v(%T), expected -> %v(%T)\n", actual, actual, param.Expected, param.Expected)
			}
		}
	})

	t.Run("test Get (array) ", func(t *testing.T) {

		// setup
		var jsonObject interface{}
		json.Unmarshal([]byte(`
{
  "stringArray": [
    "stringValue2",
    "stringValue3"
  ],
  "numberArray": [
    2,
    3
  ],
  "nullArray": [
    null,
    null
  ],
  "booleanArray": [
    true,
    false
  ]
}
`), &jsonObject)

		// test cases
		testCases := []struct {
			TestCase string
			Input    string
			Expected []interface{}
		}{
			{
				TestCase: "stringArray",
				Input:    "stringArray",
				Expected: createInterfaceArray("stringValue2", "stringValue3"),
			},
			{
				TestCase: "numberArray",
				Input:    "numberArray",
				Expected: createInterfaceArray(2.0, 3.0),
			},
			{
				TestCase: "nullArray",
				Input:    "nullArray",
				Expected: createInterfaceArray(nil, nil),
			},
			{
				TestCase: "booleanArray",
				Input:    "booleanArray",
				Expected: createInterfaceArray(true, false),
			},
		}

		// run
		for i := range testCases {
			param := testCases[i]
			actual := Get(jsonObject, param.Input).([]interface{})
			t.Logf("Case:%v\n", param.TestCase)
			if len(actual) != len(param.Expected) {
				t.Errorf("  Failed: len(actual) -> %v, len(expected) -> %v\n", len(actual), len(param.Expected))
			}
			for i, _ := range actual {
				if actual[i] != param.Expected[i] {
					t.Errorf("  Failed: i -> %v, actual -> %v(%T), expected -> %v(%T)\n", i, actual[i], actual[i], param.Expected[i], param.Expected[i])
				}
			}
		}
	})

	t.Run("test Get (object array) ", func(t *testing.T) {

		// setup
		var jsonObject interface{}
		json.Unmarshal([]byte(`
{
  "objectArray": [
    {
      "string": "stringValue5",
      "number": 5
    },
    {
      "string": "stringValue6",
      "number": 6
    }
  ]
}
`), &jsonObject)

		// test cases
		testCases := []struct {
			TestCase string
			Input    string
			Expected []interface{}
		}{
			{
				TestCase: "objectArray",
				Input:    "objectArray",
				Expected: createInterfaceArray(map[string]interface{}{"string": "stringValue5", "number": 5.0}, map[string]interface{}{"string": "stringValue6", "number": 6.0}),
			},
		}

		// run
		for i := range testCases {
			param := testCases[i]
			actual := Get(jsonObject, param.Input).([]interface{})
			t.Logf("Case:%v\n", param.TestCase)
			if len(actual) != len(param.Expected) {
				t.Errorf("  Failed: len(actual) -> %v, len(expected) -> %v\n", len(actual), len(param.Expected))
			}
			for i, _ := range actual {
				actualMap := actual[i].(map[string]interface{})
				expectedMap := param.Expected[i].(map[string]interface{})
				if len(actualMap) != len(expectedMap) {
					t.Errorf("  Failed: len(actualMap) -> %v, len(expectedMap) -> %v\n", len(actualMap), len(expectedMap))
				}
				for j, _ := range actualMap {
					if actualMap[j] != expectedMap[j] {
						t.Errorf("  Failed: actual -> %v(%T), expected -> %v(%T)\n", actualMap[j], actualMap[j], expectedMap[j], expectedMap[j])
					}
				}
			}
		}
	})

}

func TestToJsonValue(t *testing.T) {
	t.Run("test ToJsonString", func(t *testing.T) {

		// test cases
		testCases := []struct {
			TestCase string
			Input    string
			Expected string
		}{
			{
				TestCase: "string",
				Input:    `{"key":"string"}`,
				Expected: `"string"`,
			},
			{
				TestCase: "number",
				Input:    `{"key":1}`,
				Expected: "1",
			},
			{
				TestCase: "null",
				Input:    `{"key":null}`,
				Expected: "null",
			},
			{
				TestCase: "boolean",
				Input:    `{"key":true}`,
				Expected: "true",
			},
			{
				TestCase: "array",
				Input:    `{"key":["a","b"]}`,
				Expected: `["a","b"]`,
			},
		}

		// run
		var jsonObject interface{}
		for i := range testCases {
			param := testCases[i]
			t.Logf("Case:%v\n", param.TestCase)
			json.Unmarshal([]byte(param.Input), &jsonObject)
			actual := ToJsonString(Get(jsonObject, "key"))
			if actual != param.Expected {
				t.Errorf("  Failed: actual -> %v(%T), expected -> %v(%T)\n", actual, actual, param.Expected, param.Expected)
			}
		}
	})
}

func createInterfaceArray(values ...interface{}) []interface{} {
	result := make([]interface{}, len(values))
	for i, s := range values {
		result[i] = s
	}
	return result
}
