package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	err := convertGoEnums(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

type TypeDecl struct {
	Name   string
	Type   string
	Values []*NameValue
}
type NameValue struct {
	Name  string
	Value string
	Text  string // parsed to text
}

func convertGoEnums(args []string) error {
	var fileName string = "-"
	if len(args) >= 1 {
		fileName = args[0]
		// return fmt.Errorf("requires file")
	}

	file, fset, err := parseGoFile(fileName)
	if err != nil {
		return err
	}
	decls, err := parseEnumDecls(fset, file)
	if err != nil {
		return err
	}
	tsDecl := FormatTypescriptEnums(decls)
	fmt.Println(tsDecl)

	goDecl := FormatGoMapping(decls)
	fmt.Println(goDecl)

	return nil
}

func FormatTypescriptEnums(decls []*TypeDecl) string {
	var tsDecls []string
	for _, decl := range decls {
		tsDecls = append(tsDecls, FormatTypescriptEnum(decl))
	}

	return "// auto generated \n" + strings.Join(tsDecls, "\n")
}

func FormatGoMapping(decls []*TypeDecl) string {
	var goDecls []string
	for _, decl := range decls {
		goDecls = append(goDecls, FormatGoEnum(decl))
	}

	return "// auto generated \n" + strings.Join(goDecls, "\n")
}

func FormatGoEnum(decl *TypeDecl) string {
	typeName := decl.Name

	var kvs []string
	for _, val := range decl.Values {
		kvs = append(kvs, fmt.Sprintf("    %s: %q, ", val.Name, val.Text))
	}
	return fmt.Sprintf("var %sTextMapping = map[%s]string{\n%s\n}", typeName, typeName, strings.Join(kvs, "\n"))
}
func FormatTypescriptEnum(decl *TypeDecl) string {
	var tsValues []string
	var enumValues []string
	var mappingValus []string

	lowerCaseName := strings.ToLower(decl.Name[0:1]) + decl.Name[1:]
	for _, declValue := range decl.Values {
		name := declValue.Name
		if strings.HasPrefix(name, decl.Name+"_") {
			name = strings.TrimPrefix(name, decl.Name+"_")
		}
		tsValues = append(tsValues, fmt.Sprintf("    %s = %s,", name, declValue.Value))
		enumValues = append(enumValues, fmt.Sprintf("%s.%s", decl.Name, name))
		mappingValus = append(mappingValus, fmt.Sprintf(`    [%s.%s]: %q,`, decl.Name, name, declValue.Text))
	}
	return fmt.Sprintf("enum %s {\n%s\n}\nconst %sValues:%s[] = [%s]\nconst %sMapping:Record<%s,string> = {\n%s\n}\n",
		decl.Name, strings.Join(tsValues, "\n"),
		lowerCaseName, decl.Name, strings.Join(enumValues, ","),
		lowerCaseName, decl.Name, strings.Join(mappingValus, "\n"),
	)
}
func parseGoFile(file string) (*ast.File, *token.FileSet, error) {
	fileName := file
	var contentReader io.Reader
	if file == "-" {
		fileName = "<stdin>"
		contentReader = os.Stdin
	} else {
		readFile, err := os.Open(file)
		if err != nil {
			return nil, nil, err
		}
		contentReader = readFile
	}

	content, err := ioutil.ReadAll(contentReader)
	if err != nil {
		return nil, nil, err
	}
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	var hasPackage bool
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			hasPackage = true
			break
		}
	}
	if !hasPackage {
		contentStr = "package main;" + contentStr
	}

	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, fileName, contentStr, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return ast, fset, nil
}

func parseEnumDecls(fset *token.FileSet, file *ast.File) ([]*TypeDecl, error) {
	typeMapping := make(map[string]*TypeDecl)
	var typeList []*TypeDecl

	lineComments := make(map[int]string)
	for _, cmts := range file.Comments {
		for _, cmt := range cmts.List {
			line := fset.Position(cmt.Pos()).Line
			lineComments[line] = cmt.Text
		}
	}
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						var typeStr string
						if idt, ok := typeSpec.Type.(*ast.Ident); ok {
							typeStr = idt.Name
						}
						name := typeSpec.Name.Name
						_, ok := typeMapping[name]
						if ok {
							return nil, fmt.Errorf("duplicate type: %s", name)
						}
						decl := &TypeDecl{
							Name: name,
							Type: typeStr,
						}
						typeMapping[name] = decl
						typeList = append(typeList, decl)
					}
				}
			}
			if genDecl.Tok == token.CONST {
				for _, spec := range genDecl.Specs {
					if constSpec, ok := spec.(*ast.ValueSpec); ok {
						var typeName string
						if idt, ok := constSpec.Type.(*ast.Ident); ok {
							typeName = idt.Name
						}
						if typeName == "" {
							return nil, fmt.Errorf("unknown type for const decl: %T %v", constSpec.Type, constSpec.Type)
						}
						typeDecl := typeMapping[typeName]
						if typeDecl == nil {
							return nil, fmt.Errorf("unknown type: %s", typeName)
						}
						var names []string
						var lines []int
						for _, name := range constSpec.Names {
							names = append(names, name.Name)
							line := fset.Position(name.Pos()).Line
							lines = append(lines, line)
						}

						var values []string
						for _, value := range constSpec.Values {
							var valueLit string
							if lit, ok := value.(*ast.BasicLit); ok {
								valueLit = lit.Value
							} else {
								return nil, fmt.Errorf("unknown const spec value: %T %v", value, value)
							}
							values = append(values, valueLit)
						}
						if len(names) != len(values) {
							return nil, fmt.Errorf("mismatch decl: names=%v,values=%v", names, values)
						}
						for i := 0; i < len(names); i++ {
							line := lines[i]
							lineComment := lineComments[line-1]
							var text string
							const prefix = "// text:"
							if strings.HasPrefix(lineComment, prefix) {
								text = lineComment[len(prefix):]
								text = strings.TrimSpace(text)
							}
							typeDecl.Values = append(typeDecl.Values, &NameValue{
								Name:  names[i],
								Value: values[i],
								Text:  text,
							})
						}
					}
				}
			}
		}
	}
	return typeList, nil
}
