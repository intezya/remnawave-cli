package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	jmespath "github.com/danielgtaylor/go-jmespath-plus"
	"github.com/danielgtaylor/openapi-cli-generator/cli"
	"github.com/spf13/viper"
)

type agentFormatter struct {
	fallback cli.ResponseFormatter
}

type tableData struct {
	label     string
	rows      []interface{}
	total     int
	shown     int
	truncated bool
}

func configureAgentOutput() {
	cli.AddGlobalFlag("output", "", "Output format [json, yaml, agent, ndjson]", "")
	cli.AddGlobalFlag("fields", "", "Comma-separated fields for agent/ndjson output", "")
	cli.AddGlobalFlag("limit", "", "Maximum rows for agent output; 0 disables truncation", 50)
	cli.AddGlobalFlag("save-full", "", "Write the full JSON response to this path before compacting output", "")
	cli.Formatter = &agentFormatter{fallback: cli.Formatter}
}

func (f *agentFormatter) Format(data interface{}) error {
	if data == nil {
		return nil
	}

	projected, err := applyQuery(data)
	if err != nil {
		return err
	}

	if path := viper.GetString("save-full"); path != "" {
		if err := writeFullJSON(path, projected); err != nil {
			return err
		}
	}

	switch outputFormat() {
	case "agent":
		return formatAgent(projected)
	case "ndjson":
		return formatNDJSON(projected)
	default:
		return f.fallback.Format(data)
	}
}

func outputFormat() string {
	if value := strings.TrimSpace(viper.GetString("output")); value != "" {
		return strings.ToLower(value)
	}
	return strings.ToLower(strings.TrimSpace(viper.GetString("output-format")))
}

func applyQuery(data interface{}) (interface{}, error) {
	query := strings.TrimSpace(viper.GetString("query"))
	if query == "" {
		return data, nil
	}
	return jmespath.Search(query, data)
}

func writeFullJSON(path string, data interface{}) error {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(path, encoded, 0600)
}

func formatNDJSON(data interface{}) error {
	rows := rowsForOutput(data)
	fields := selectedFields(rows)

	for _, row := range rows {
		value := row
		if len(fields) > 0 {
			value = projectRow(row, fields)
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return err
		}
		fmt.Fprintln(cli.Stdout, string(encoded))
	}

	return nil
}

func formatAgent(data interface{}) error {
	table := extractTable(data)
	fields := selectedFields(table.rows)
	if len(fields) == 0 {
		fields = inferFields(table.rows, 8)
	}

	fmt.Fprintf(cli.Stdout, "ok %s total=%d shown=%d truncated=%t", table.label, table.total, table.shown, table.truncated)
	if path := viper.GetString("save-full"); path != "" {
		fmt.Fprintf(cli.Stdout, " full=%s", path)
	}
	fmt.Fprintln(cli.Stdout)

	if len(table.rows) == 0 {
		return nil
	}

	if len(fields) == 0 {
		return printCompactValues(table.rows)
	}

	fmt.Fprintln(cli.Stdout, strings.Join(fields, "\t"))
	for _, row := range table.rows {
		values := make([]string, 0, len(fields))
		for _, field := range fields {
			values = append(values, compactValue(getPath(row, field)))
		}
		fmt.Fprintln(cli.Stdout, strings.Join(values, "\t"))
	}

	return nil
}

func extractTable(data interface{}) tableData {
	label := "response"
	rows := rowsForOutput(data)

	if namedLabel, namedRows, ok := findNamedRows(data); ok {
		label = namedLabel
		rows = namedRows
	}

	total := len(rows)
	limit := viper.GetInt("limit")
	truncated := false
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
		truncated = true
	}

	return tableData{
		label:     label,
		rows:      rows,
		total:     total,
		shown:     len(rows),
		truncated: truncated,
	}
}

func rowsForOutput(data interface{}) []interface{} {
	switch value := data.(type) {
	case []interface{}:
		return value
	default:
		return []interface{}{value}
	}
}

func findNamedRows(data interface{}) (string, []interface{}, bool) {
	obj, ok := data.(map[string]interface{})
	if !ok {
		return "", nil, false
	}

	preferred := []string{"users", "nodes", "hosts", "inbounds", "squads", "items", "data", "response"}
	for _, key := range preferred {
		if rows, ok := asRows(obj[key]); ok {
			return key, rows, true
		}
	}

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if rows, ok := asRows(obj[key]); ok {
			return key, rows, true
		}
	}

	return "", nil, false
}

func asRows(value interface{}) ([]interface{}, bool) {
	switch typed := value.(type) {
	case []interface{}:
		return typed, true
	case map[string]interface{}:
		if label, rows, ok := findNamedRows(typed); ok && label != "" {
			return rows, true
		}
	}
	return nil, false
}

func selectedFields(rows []interface{}) []string {
	raw := strings.TrimSpace(viper.GetString("fields"))
	if raw == "" {
		return nil
	}

	fields := strings.Split(raw, ",")
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			result = append(result, field)
		}
	}
	return result
}

func inferFields(rows []interface{}, max int) []string {
	if len(rows) == 0 {
		return nil
	}

	obj, ok := rows[0].(map[string]interface{})
	if !ok {
		return nil
	}

	keys := make([]string, 0, len(obj))
	for key, value := range obj {
		if isScalar(value) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	if len(keys) > max {
		keys = keys[:max]
	}
	return keys
}

func projectRow(row interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		result[field] = getPath(row, field)
	}
	return result
}

func getPath(value interface{}, path string) interface{} {
	current := value
	for _, part := range strings.Split(path, ".") {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = obj[part]
	}
	return current
}

func isScalar(value interface{}) bool {
	switch value.(type) {
	case nil, bool, string, json.Number, int, int64, float64:
		return true
	default:
		return false
	}
}

func compactValue(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return "-"
	case string:
		if typed == "" {
			return "-"
		}
		return compactString(typed, 80)
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case json.Number:
		return typed.String()
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return compactString(string(encoded), 80)
	}
}

func compactString(value string, max int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func printCompactValues(rows []interface{}) error {
	for _, row := range rows {
		fmt.Fprintln(cli.Stdout, compactValue(row))
	}
	return nil
}
