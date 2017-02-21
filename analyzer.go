package goi18np

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

// Const values
const (
	DefaultAnalyzerFuncName = "T"
	MaxDepth                = 5
)

// Analyzer interface
type Analyzer interface {
	Name() string
}

// DefaultAnalyzer struct
type DefaultAnalyzer struct {
	FuncName string // Function name
	Debug    bool
	Records  []I18NRecord
}

// Name returns the function name that is used to analyze go source code
func (da DefaultAnalyzer) Name() string {
	if da.FuncName != "" {
		return da.FuncName
	}
	return DefaultAnalyzerFuncName
}

// I18NRecord has `id` and `translation` field
type I18NRecord struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

// AnalyzeFromFile analyzes a go file
func (da *DefaultAnalyzer) AnalyzeFromFile(filename string) []I18NRecord {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		panic(err)
	}

	if da.Debug {
		ast.Print(fset, f)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			ident := traversalToIdent(x.Fun, 0)
			if ident == nil {
				return true
			}

			fn := ident.Name
			if fn == "T" {
				key := strings.Trim(x.Args[0].(*ast.BasicLit).Value, "\"")
				r := I18NRecord{
					ID: key,
				}
				if !containsID(key, da.Records) {
					da.Records = append(da.Records, r)
				}
			}
		}
		return true
	})

	return da.Records
}

func traversalToIdent(n interface{}, depth int) *ast.Ident {
	switch x := n.(type) {
	case *ast.Ident:
		return x
	case *ast.SelectorExpr:
		if depth < MaxDepth {
			return traversalToIdent(x.Sel, depth+1)
		}
	default:
		return nil
	}
	return nil
}

// AnalyzeFromFiles analyzes multiple go source files
func (da *DefaultAnalyzer) AnalyzeFromFiles(files []string) []I18NRecord {
	for _, filename := range files {
		da.AnalyzeFromFile(filename)
	}
	return da.Records
}

// DumpJSON returns []byte and error marshaled data analyzed from source
func (da DefaultAnalyzer) DumpJSON() ([]byte, error) {
	return json.Marshal(da.Records)
}

// SaveJSON saves JSON based on go-i18np format
func (da DefaultAnalyzer) SaveJSON(path string) error {
	out, err := da.DumpJSON()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func containsID(id string, rs []I18NRecord) bool {
	for _, r := range rs {
		if r.ID == id {
			return true
		}
	}
	return false
}
