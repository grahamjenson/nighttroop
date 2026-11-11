//go:build ignore

// getters writes z_getters.go in a sqlc folder: a Get and Set method for every field of the
// table models in z_models.go and of the ...Params structs in the query files. It reads the
// structs sqlc wrote, so each method has the exact type sqlc chose, overrides included.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	folder := flag.String("folder", "", "sqlc folder to write z_getters.go in")
	flag.Parse()

	if *folder == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := writeGetters(*folder); err != nil {
		log.Fatalf("Failed to write getters: %v", err)
	}
}

func writeGetters(folder string) error {
	queryFiles, err := filepath.Glob(filepath.Join(folder, "*.sql.gen.go"))
	if err != nil {
		return err
	}
	files := append([]string{filepath.Join(folder, "z_models.go")}, queryFiles...)

	pkg := ""
	imports := map[string]string{} // import path to its spec as written
	var body strings.Builder

	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		pkg = file.Name.Name
		isModels := filepath.Base(path) == "z_models.go"

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok || (!isModels && !strings.HasSuffix(typeSpec.Name.Name, "Params")) {
					continue
				}
				for _, field := range structType.Fields.List {
					if err := addImports(file, field.Type, imports); err != nil {
						return fmt.Errorf("%s.%s: %w", typeSpec.Name.Name, field.Names, err)
					}
					fieldType := types.ExprString(field.Type)
					for _, name := range field.Names {
						writeMethods(&body, typeSpec.Name.Name, name.Name, fieldType)
					}
				}
			}
		}
	}

	importsString := ""
	if len(imports) > 0 {
		paths := make([]string, 0, len(imports))
		for path := range imports {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		importsString = "import (\n"
		for _, path := range paths {
			importsString += "\t" + imports[path] + "\n"
		}
		importsString += ")\n\n"
	}

	content := fmt.Sprintf("\npackage %s\n\n%s\n%s", pkg, importsString, body.String())
	return os.WriteFile(filepath.Join(folder, "z_getters.go"), []byte(content), 0o644)
}

// addImports records the import of each package the type expression names.
func addImports(file *ast.File, expr ast.Expr, imports map[string]string) error {
	var err error
	ast.Inspect(expr, func(n ast.Node) bool {
		selector, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkgName := selector.X.(*ast.Ident).Name
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			if spec.Name != nil {
				if spec.Name.Name == pkgName {
					imports[path] = spec.Name.Name + " " + spec.Path.Value
					return false
				}
			} else if filepath.Base(path) == pkgName {
				imports[path] = spec.Path.Value
				return false
			}
		}
		err = fmt.Errorf("no import for package %s", pkgName)
		return false
	})
	return err
}

func writeMethods(sb *strings.Builder, structName, fieldName, fieldType string) {
	fmt.Fprintf(sb, "// Get%s returns the %s value.\n", fieldName, fieldName)
	fmt.Fprintf(sb, "func (m %s) Get%s() %s {\n", structName, fieldName, fieldType)
	fmt.Fprintf(sb, "	return m.%s\n", fieldName)
	sb.WriteString("}\n\n")

	fmt.Fprintf(sb, "// Set%s sets the %s value and returns the updated %s.\n", fieldName, fieldName, structName)
	fmt.Fprintf(sb, "func (m *%s) Set%s(v %s) {\n", structName, fieldName, fieldType)
	fmt.Fprintf(sb, "	m.%s = v\n", fieldName)
	sb.WriteString("}\n\n")
}
