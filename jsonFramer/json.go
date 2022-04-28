package jsonFramer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/tidwall/gjson"
	"github.com/yesoreyeram/grafana-framer/gframer"
)

type JSONFramerOptions struct {
	FrameName    string
	RootSelector string
	Columns      []ColumnSelector
}

type ColumnSelector struct {
	Selector string
	Alias    string
	Type     string
}

func JsonStringToFrame(jsonString string, options JSONFramerOptions) (frame *data.Frame, err error) {
	if strings.Trim(jsonString, " ") == "" {
		return frame, errors.New("empty json received")
	}
	if !gjson.Valid(jsonString) {
		return frame, errors.New("invalid json response received")
	}
	outString := jsonString
	if options.RootSelector != "" {
		r := gjson.Get(string(jsonString), options.RootSelector)
		if !r.Exists() {
			return frame, errors.New("root object doesn't exist in the response. Root selector:" + options.RootSelector)
		}
		outString = r.String()
	}
	outString, err = getColumnValuesFromResponseString(outString, options.Columns)
	if err != nil {
		return frame, err
	}
	return getFrameFromResponseString(outString, options.FrameName)
}

func getColumnValuesFromResponseString(responseString string, columns []ColumnSelector) (string, error) {
	if len(columns) > 0 {
		outString := responseString
		result := gjson.Parse(outString)
		out := []map[string]interface{}{}
		if result.IsArray() {
			result.ForEach(func(key, value gjson.Result) bool {
				oi := map[string]interface{}{}
				for _, col := range columns {
					name := col.Alias
					if name == "" {
						name = col.Selector
					}
					oi[name] = gjson.Get(value.Raw, col.Selector).Value()
				}
				out = append(out, oi)
				return true
			})
		}
		if !result.IsArray() && result.IsObject() {
			oi := map[string]interface{}{}
			for _, col := range columns {
				name := col.Alias
				if name == "" {
					name = col.Selector
				}
				oi[name] = gjson.Get(result.Raw, col.Selector).Value()
			}
			out = append(out, oi)
		}
		a, err := json.Marshal(out)
		if err != nil {
			return "", err
		}
		return string(a), nil
	}
	return responseString, nil
}

func getFrameFromResponseString(responseString string, frameName string) (frame *data.Frame, err error) {
	var out interface{}
	err = json.Unmarshal([]byte(responseString), &out)
	if err != nil {
		return frame, fmt.Errorf("error while un-marshaling response. %s", err.Error())
	}
	return gframer.ToDataFrame(out, gframer.FramerOptions{
		FrameName: frameName,
	})
}
