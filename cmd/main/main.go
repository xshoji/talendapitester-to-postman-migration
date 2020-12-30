package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"github.com/xshoji/go-sample-box/jsongetvalue/jsonutil"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Arguments struct {
	inputFilePath        *string
	outputFilePathPrefix *string
	debug                *bool
	help                 *bool
}

var (
	arguments = &Arguments{
		flag.String("i", "" /*      */, "[Required] Input file path of Talend API Tester json ( Support input from stdin )"),
		flag.String("o", "" /*      */, "[Required] Output file path prefix for postman ( example: /tmp/postman )"),
		flag.Bool("d", false /*   */, "\n[Optional] Debug"),
		flag.Bool("h", false /*   */, "\nHelp"),
	}
	PrefixSlashMatcherRegexp                *regexp.Regexp
	EnvironmentStringReplacingMatcherRegexp *regexp.Regexp
)

func init() {
	// 時刻と時刻のマイクロ秒、ディレクトリパスを含めたファイル名を出力
	log.SetFlags(log.Llongfile | log.LstdFlags)
	PrefixSlashMatcherRegexp = regexp.MustCompile("^/")
	EnvironmentStringReplacingMatcherRegexp = regexp.MustCompile("\\${([A-Z0-9]*)\\}")
}

// [ Usage ]
// go run cmd/main/main.go -i=/tmp/exported.json -o=/tmp/postman
// cat /tmp/exported.json |go run cmd/main/main.go -o=/tmp/postman
func main() {

	// Parse arguments
	flag.Parse()
	if *arguments.help || *arguments.outputFilePathPrefix == "" {
		flag.Usage()
		os.Exit(0)
	}
	jsonObject := inputToJsonObject(arguments)

	// Output collections json
	for _, talendEntity := range jsonutil.AsSlice(jsonObject, "entities") {
		postmanJson := make(map[string]interface{})
		var postmanItems []interface{}
		for _, talendEntityChild := range jsonutil.AsSlice(talendEntity, "children") {
			postmanItems = append(postmanItems, toPostmanItem(talendEntityChild))
		}
		postmanJson["info"] = toPostmanInfo(talendEntity)
		postmanJson["item"] = postmanItems
		environmentVariableFixedJsonString := EnvironmentStringReplacingMatcherRegexp.ReplaceAllString(jsonutil.ToJsonStringPretty(postmanJson), "{{$1}}")

		// output json
		if err := ioutil.WriteFile(*arguments.outputFilePathPrefix+"_collection_"+jsonutil.AsString(talendEntity, "entity.name")+".json", []byte(environmentVariableFixedJsonString), 0666); err != nil {
			log.Fatal(err)
		}
	}

	// Output environment json
	for _, talendEnvironment := range jsonutil.AsSlice(jsonObject, "environments") {
		postmanJson := make(map[string]interface{})
		var postmanEnvironmentValues []interface{}
		talendEnvironmentVariables, ok := jsonutil.Get(talendEnvironment, "variables").(map[string]interface{})
		if !ok {
			log.Fatal("Cannot cast talend environment variables to map[string]interface{}")
		}
		for _, talendEnvironmentVariable := range talendEnvironmentVariables {
			key := jsonutil.AsString(talendEnvironmentVariable, "name")
			if key == "" {
				continue
			}
			postmanEnvironmentValues = append(postmanEnvironmentValues, map[string]interface{}{
				"key":     jsonutil.AsString(talendEnvironmentVariable, "name"),
				"value":   jsonutil.AsString(talendEnvironmentVariable, "value"),
				"enabled": true,
			})
		}
		postmanJson["id"] = createPostmanId()
		postmanJson["name"] = jsonutil.AsString(talendEnvironment, "name")
		postmanJson["values"] = postmanEnvironmentValues
		postmanJson["_postman_variable_scope"] = "environment"
		postmanJson["_postman_exported_at"] = time.Now().Format("2006-01-02T15:04:05Z")
		postmanJson["_postman_exported_using"] = "Postman/7.36.1"

		environmentVariableFixedJsonString := EnvironmentStringReplacingMatcherRegexp.ReplaceAllString(jsonutil.ToJsonStringPretty(postmanJson), "{{$1}}")

		// output json
		if err := ioutil.WriteFile(*arguments.outputFilePathPrefix+"_environment_"+jsonutil.AsString(talendEnvironment, "name")+".json", []byte(environmentVariableFixedJsonString), 0666); err != nil {
			log.Fatal(err)
		}
	}
}

func inputToJsonObject(arguments *Arguments) interface{} {
	// Load json contents
	stat, _ := os.Stdin.Stat()
	var jsonContentBytes []byte
	isStdinMode := (stat.Mode() & os.ModeCharDevice) == 0
	isSpecifiedJsonFilePath := *arguments.inputFilePath != ""
	if isStdinMode {
		// read from stdin
		jsonContentBytes, _ = ioutil.ReadAll(os.Stdin)
	} else if isSpecifiedJsonFilePath {
		// read from file
		file, err := os.Open(*arguments.inputFilePath)
		if err != nil {
			log.Fatal(err)
		}

		jsonContentBytes, err = ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("Empty input.\n")
	}
	jsonContents := string(jsonContentBytes)

	// Convert to json object
	var jsonObject interface{}
	err := json.Unmarshal([]byte(jsonContents), &jsonObject)
	if err != nil {
		log.Fatal(err)
	}
	return jsonObject
}

func toPostmanInfo(jsonContents interface{}) interface{} {
	return map[string]string{
		"_postman_id": createPostmanId(),
		"name":        jsonutil.AsString(jsonContents, "entity.name"),
		"schema":      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
	}
}

func toPostmanItem(jsonContents interface{}) interface{} {
	originMethod := jsonutil.AsString(jsonContents, "entity.method.name")
	return map[string]interface{}{
		"name": jsonutil.AsString(jsonContents, "entity.name"),
		"request": map[string]interface{}{
			"method":  originMethod,
			"headers": toPostmanItemRequestHeaders(jsonContents),
			"body":    toPostmanItemRequestBody(jsonContents, originMethod),
			"url":     toPostmanItemRequestUri(jsonContents),
		},
	}
}

func toPostmanItemRequestHeaders(jsonContents interface{}) []map[string]string {
	originHeaders := jsonutil.AsSlice(jsonContents, "entity.headers")
	var headers []map[string]string
	for _, header := range originHeaders {
		headers = append(headers, map[string]string{
			"key":   jsonutil.AsString(header, "name"),
			"value": jsonutil.AsString(header, "value"),
			"type":  "text",
		})
	}
	return headers
}

func toPostmanItemRequestBody(jsonContents interface{}, originMethod string) map[string]string {
	body := make(map[string]string)
	switch originMethod {
	case "POST":
		body["mode"] = "raw"
		body["raw"] = jsonutil.AsString(jsonContents, "entity.body.textBody")
		return body
	case "GET":
		// do nothing
	default:
		// do nothing
	}
	return nil
}

func toPostmanItemRequestUri(jsonContents interface{}) map[string]interface{} {
	originPathString := PrefixSlashMatcherRegexp.ReplaceAllString(jsonutil.AsString(jsonContents, "entity.uri.path"), "")
	paths := strings.Split(originPathString, "/")
	if originPathString == "" {
		paths = nil
	}

	originHostString := jsonutil.AsString(jsonContents, "entity.uri.host")
	hosts := strings.Split(originHostString, ".")
	if originHostString == "" {
		hosts = nil
	}
	originQueries := jsonutil.AsSlice(jsonContents, "entity.uri.query.items")
	var queries []map[string]string
	for _, query := range originQueries {
		queries = append(queries, map[string]string{
			"key":   jsonutil.AsString(query, "name"),
			"value": jsonutil.AsString(query, "value"),
		})
	}
	uri := map[string]interface{}{
		"protocol": jsonutil.AsString(jsonContents, "entity.uri.scheme.name"),
		"host":     hosts,
		"path":     paths,
		"query":    queries,
	}
	return uri
}

func createPostmanId() string {
	// Create random integer
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(99999999-1) + 10000000

	// Create random string
	seed := strconv.FormatInt(time.Now().UnixNano(), 10)
	shaBytes := sha256.Sum256([]byte(seed))
	randString := hex.EncodeToString(shaBytes[:])

	return strings.Join([]string{
		strconv.Itoa(randNum),
		randString[0:4],
		randString[4:8],
		randString[8:12],
		randString[12:24],
	}, `-`)
}
