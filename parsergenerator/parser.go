package parsergenerator

import (
	"errors"
	"strings"
)

type Context struct {
	availableModels map[string]interface{}
}

type Options struct {
	Folder       string
	OutputFolder string
	Go           GoOptions
}
type GoOptions struct {
	Package     string
	FiberImport string
}

type Generator interface {
	GenerateAll(*Options, []GenDef, *Context) (*string, error)
}

func (c *Context) isModel(t string) bool {
	_, present := c.availableModels[t]
	return present
}

func ParseAll(opts *Options, gs []GenDef, gen Generator) (*string, error) {
	ctx := Context{availableModels: make(map[string]interface{})}
	for _, g := range gs {
		err := validateGenDef(&g)
		if err != nil {
			return nil, err
		}
		err = collectAllModels(&g, &ctx)
		if err != nil {
			return nil, err
		}
	}
	return gen.GenerateAll(opts, gs, &ctx)
}

func parseCompoundType(raw string, ctx *Context) (anyType, error) {
	if isKeyword(raw) {
		return nil, isKeywordError(raw)
	}
	if isBaseType(raw) || ctx.isModel(raw) {
		return raw, nil
	}
	if len(raw) >= 3 && raw[:2] == "[]" {
		of, err := parseCompoundType(raw[2:], ctx)
		if err != nil {
			return nil, err
		}
		return array{of: of}, nil
	}
	if len(raw) >= 7 && raw[:3] == "map" {
		openBracket := strings.Index(raw, "[")
		closeBracket := strings.LastIndex(raw, "]")
		if openBracket == -1 || closeBracket == -1 {
			return nil, errors.New("missing open or close bracket: " + raw)
		}
		from, err := parseBaseType(raw[openBracket+1:closeBracket], ctx)
		if err != nil {
			return nil, err
		}
		to, err := parseCompoundType(raw[closeBracket+1:], ctx)
		if err != nil {
			return nil, err
		}
		return mapType{from: *from, to: to}, nil
	}
	return nil, errors.New("could not parse type: " + raw)
}

func collectAllModels(g *GenDef, ctx *Context) error {
	for modelName := range g.Models {
		if ctx.isModel(modelName) {
			return errors.New("duplicate model name: " + modelName)
		}
		ctx.availableModels[modelName] = nil
	}
	return nil
}

func parseBaseType(t string, ctx *Context) (*baseType, error) {
	if isBaseType(t) {
		ret := baseType(t)
		return &ret, nil
	}
	return nil, errors.New("not a raw type: " + t)
}
