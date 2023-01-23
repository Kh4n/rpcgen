package parsergenerator

import (
	"errors"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var keywords = map[string]interface{}{
	"break":       nil,
	"default":     nil,
	"func":        nil,
	"interface":   nil,
	"select":      nil,
	"case":        nil,
	"defer":       nil,
	"go":          nil,
	"map":         nil,
	"struct":      nil,
	"chan":        nil,
	"else":        nil,
	"goto":        nil,
	"package":     nil,
	"switch":      nil,
	"const":       nil,
	"fallthrough": nil,
	"if":          nil,
	"range":       nil,
	"type":        nil,
	"continue":    nil,
	"for":         nil,
	"import":      nil,
	"return":      nil,
	"var":         nil,
}

var baseTypes = map[string]interface{}{
	"string":  nil,
	"int32":   nil,
	"int64":   nil,
	"uint32":  nil,
	"uint64":  nil,
	"float32": nil,
	"float64": nil,
}

func isKeyword(s string) bool {
	_, present := keywords[s]
	return present || strings.HasSuffix(s, "Params")
}
func isKeywordError(kw string) error {
	return errors.New(kw + " is a reserved keyword or has reserved suffix 'Params', use a different name")
}

func isBaseType(t string) bool {
	_, present := baseTypes[t]
	return present
}

func capitalFirst(s string) string {
	return cases.Title(language.English, cases.NoLower).String(s)
}
