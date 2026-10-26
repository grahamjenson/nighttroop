//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type Options struct {
	PackageName string   `json:"package"`
	Imports     []string `json:"imports"`
}

var TableNameMap = map[string]string{
	"InsulinDos": "InsulinDose",
}

var AttributeNameMap = map[string]string{}

var TypeLookupMap = map[string]map[string]string{
	"ControlIdentity": {
		"Role": "consts.ControlRole",
	},
	"SiteIdentity": {
		"Role":    "consts.IdentityRole",
		"Enabled": "bool",
	},
	"Job": {
		"Errors": "errors.ErrorList",
	},
	"JobStatsDaily": {
		"Errors": "errors.ErrorList",
	},
	"InsertFailedCronJobParams": {
		"Errors": "errors.ErrorList",
	},
	"SiteSyncSource": {
		"Type":               "consts.SyncSourceType",
		"Enabled":            "bool",
		"PreviousSyncErrors": "errors.ErrorList",
	},
	"FailJobParams": {
		"Errors": "errors.ErrorList",
	},
	"CreateSiteIdentityParams": {
		"Role": "consts.IdentityRole",
	},
	"CreateControlIdentityParams": {
		"Role": "consts.ControlRole",
	},
	"CreateIfNotExistsIdentityParams": {
		"Role": "consts.IdentityRole",
	},
	"CreateSyncSourceParams": {
		"Type":    "consts.SyncSourceType",
		"Enabled": "bool",
	},
	"UpdateSiteIdentityEnabledParams": {
		"Enabled": "bool",
	},
	"UpdateSyncSourceParams": {
		"Enabled": "bool",
		"Type":    "consts.SyncSourceType",
	},
	"GetSyncSourceParams": {
		"Type": "consts.SyncSourceType",
	},
	"DeleteJobsByIdsParams": {
		"Ids": "[]int64",
	},
	"UpdateJobStatsDailyParams": {
		"Errors": "errors.ErrorList",
	},
	"UpdateSyncSourceReportsFailureParams": {
		"ErrorMessages": "errors.ErrorList",
		"Type":          "consts.SyncSourceType",
	},
	"DeleteSyncSourceParams": {
		"Type": "consts.SyncSourceType",
	},
	"UpdateSyncReportAttemptParams": {
		"Type": "consts.SyncSourceType",
	},
	"UpdateSyncSourceReportsSuccessParams": {
		"Type": "consts.SyncSourceType",
	},
	"UpdateSyncSourceLastModifiedParams": {
		"Type": "consts.SyncSourceType",
	},
	"ProfilePreset": {
		"IsDefault": "bool",
	},
	"CreateProfilePresetParams": {
		"IsDefault": "bool",
	},
}

func run() error {
	reqBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	// Unmarshal protobuf instead of JSON
	var req plugin.GenerateRequest
	if err := proto.Unmarshal(reqBytes, &req); err != nil {
		return fmt.Errorf("decode request via protobuf: %w", err)
	}

	var opts Options
	if err := json.Unmarshal(req.PluginOptions, &opts); err != nil {
		return fmt.Errorf("decode plugin options via JSON: %w", err)
	}
	resp := &plugin.GenerateResponse{
		Files: make([]*plugin.File, 0),
	}

	importsString := ""
	if len(opts.Imports) > 0 {
		importsString = "import (\n"
		for _, imp := range opts.Imports {
			importsString += fmt.Sprintf("\t\"%s\"\n", imp)
		}
		importsString += ")\n\n"
	}

	totalContent := fmt.Sprintf(`
package %s

%s
`, opts.PackageName, importsString)

	// Iterate over all available schemas (SQLite often maps to empty or "main")
	for _, schema := range req.Catalog.Schemas {
		for _, table := range schema.Tables {
			modelName := tableNameLookup(table.Rel.GetName())
			fmt.Fprintf(os.Stderr, "Generating getters for table: %s\nColumns: %d\nModel: %v%v%v\n",
				modelName, len(table.Columns), table, schema, schema.Name)

			content := generateGetters(modelName, table.Columns)
			totalContent += content

		}
	}

	for _, query := range req.Queries {
		if query == nil {
			continue
		}
		queryName := query.Name
		cols := []*plugin.Column{}
		for _, param := range query.Params {
			if param != nil && param.Column != nil {
				// This will print out the stuff
				// jjj, _ := json.Marshal(param)
				// totalContent += "// " + string(jjj) + "\n"

				cols = append(cols, param.Column)
			}
		}
		content := generateParamGetters(queryName+"Params", cols)
		totalContent += content

	}
	resp.Files = append(resp.Files, &plugin.File{
		Name:     "z_getters.go",
		Contents: []byte(totalContent),
	})

	respBytes, err := proto.Marshal(resp)
	if err != nil {
		return fmt.Errorf("encode response via protobuf: %w", err)
	}

	// Write raw bytes out to stdout
	if _, err := os.Stdout.Write(respBytes); err != nil {
		return fmt.Errorf("write stdout: %w", err)
	}

	return nil
}

func tableNameLookup(name string) string {
	name = snakeToCamel(name)

	switch {
	case strings.HasSuffix(name, "ies"):
		name = name[:len(name)-3] + "y"

	case strings.HasSuffix(name, "ses"):
		name = name[:len(name)-2]

	case strings.HasSuffix(name, "xes"),
		strings.HasSuffix(name, "zes"),
		strings.HasSuffix(name, "ches"),
		strings.HasSuffix(name, "shes"):
		name = name[:len(name)-2]

	case strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss"):
		name = name[:len(name)-1]
	}

	if newName, ok := TableNameMap[name]; ok {
		return newName
	}
	return name
}

func attributeNameLookup(name string) string {
	name = snakeToCamel(name)
	if newName, ok := AttributeNameMap[name]; ok {
		return newName
	}
	return name
}

var goInitialisms = map[string]bool{
	"acl": true, "api": true, "ascii": true, "cpu": true, "css": true,
	"dns": true, "eof": true, "guid": true, "html": true, "http": true,
	"https": true, "id": true, "ip": true, "lhs": true,
	"qps": true, "ram": true, "rhs": true, "rpc": true, "sla": true,
	"smtp": true, "sql": true, "ssh": true, "tls": true, "ttl": true,
	"ui": true, "uid": true, "uuid": true, "uri": true,
	"utf8": true, "vm": true, "xml": true, "xmpp": true, "xsrf": true,
	"xss": true,
}

func snakeToCamel(s string) string {
	if s == "" {
		return ""
	}

	// Split by underscores
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}

		lowerPart := strings.ToLower(part)
		// If it's a known Go initialism, make it fully uppercase
		if goInitialisms[lowerPart] {
			parts[i] = strings.ToUpper(lowerPart)
		} else {
			// Otherwise capitalize the first letter correctly
			parts[i] = strings.ToUpper(string(part[0])) + lowerPart[1:]
		}
	}

	return strings.Join(parts, "")
}

func generateGetters(modelName string, columns []*plugin.Column) string {
	var sb strings.Builder

	for _, col := range columns {
		fieldName := attributeNameLookup(col.Name)
		fieldType := getSQLiteGoType(col, modelName, fieldName)

		sb.WriteString(fmt.Sprintf("// Get%s returns the %s value.\n", fieldName, fieldName))
		sb.WriteString(fmt.Sprintf("func (m %s) Get%s() %s {\n", modelName, fieldName, fieldType))
		sb.WriteString(fmt.Sprintf("	return m.%s\n", fieldName))
		sb.WriteString("}\n\n")

		sb.WriteString(fmt.Sprintf("// Set%s sets the %s value and returns the updated %s.\n", fieldName, fieldName, modelName))
		sb.WriteString(fmt.Sprintf("func (m *%s) Set%s(v %s) {\n", modelName, fieldName, fieldType))
		sb.WriteString(fmt.Sprintf("	m.%s = v\n", fieldName))
		sb.WriteString("}\n\n")

	}

	return sb.String()
}

func generateParamGetters(paramName string, columns []*plugin.Column) string {
	var sb strings.Builder

	for _, col := range columns {
		if len(col.Name) == 0 {
			continue
		}
		// If the first letter is lowercase use snakecase

		fieldName := col.Name
		if strings.ToLower(string(col.Name[0])) == string(col.Name[0]) {
			fieldName = snakeToCamel(col.Name)
		}
		fieldType := getSQLiteGoType(col, paramName, fieldName)

		sb.WriteString(fmt.Sprintf("// Get%s returns the %s value.\n", fieldName, fieldName))
		sb.WriteString(fmt.Sprintf("func (m %s) Get%s() %s {\n", paramName, fieldName, fieldType))
		sb.WriteString(fmt.Sprintf("	return m.%s\n", fieldName))
		sb.WriteString("}\n\n")

		sb.WriteString(fmt.Sprintf("// Set%s sets the %s value and returns the updated %s.\n", fieldName, fieldName, paramName))
		sb.WriteString(fmt.Sprintf("func (m *%s) Set%s(v %s) {\n", paramName, fieldName, fieldType))
		sb.WriteString(fmt.Sprintf("	m.%s = v\n", fieldName))
		sb.WriteString("}\n\n")

	}

	return sb.String()
}

func typeLookup(schemaName, fieldName string) string {

	if schemaMap, ok := TypeLookupMap[schemaName]; ok {
		if goType, ok := schemaMap[fieldName]; ok {
			return goType
		}
	}
	if strings.HasSuffix(strings.ToLower(fieldName), "enabled") {
		return "*bool"
	}
	if strings.ToLower(fieldName) == "units" {
		return "*consts.Units"
	}

	if strings.HasPrefix(strings.ToLower(fieldName), "is") {
		return "*bool"
	}
	return ""
}

func getSQLiteGoType(col *plugin.Column, modelName, fieldName string) string {
	ttt := getBaseSQLiteType(col.Type.Name)

	if goType := typeLookup(modelName, fieldName); goType != "" {
		return goType
	}

	if ttt == "interface{}" {
		return ttt
	}
	// 1. Handle Null Types (SQLite data types can be optional)
	if !col.NotNull {
		return "*" + ttt
	}

	// 2. Handle standard Non-Null Data Types
	return ttt
}

func getBaseSQLiteType(typeName string) string {
	switch strings.ToLower(typeName) {
	case "integer", "int", "bigint":
		return "int64"
	case "real", "double", "float":
		return "float64"
	case "boolean", "bool":
		return "bool"
	case "timestamp", "datetime":
		return "time.Time"
	case "blob":
		return "[]byte"
	case "text", "varchar", "char", "string":
		return "string"
	case "json":
		return "json.RawMessage"
	case "any":
		return "interface{}"
	default:
		return typeName // TEXT or VARCHAR
	}
}
