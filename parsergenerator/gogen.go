package parsergenerator

import (
	"fmt"
	"strings"

	"github.com/moznion/gowrtr/generator"
)

type serviceOutput struct {
	serviceInterface string
	paramsStructs    []string
	fiberInitializer string
}

type defOutput struct {
	services []serviceOutput
	models   []string
}

func newDefOutput() *defOutput {
	return &defOutput{services: make([]serviceOutput, 0), models: make([]string, 0)}
}

type GoGenerator struct{}

func (g *GoGenerator) GenerateAll(opts *Options, gs []GenDef, ctx *Context) (*string, error) {
	ret := newDefOutput()
	for _, def := range gs {
		err := g.generate(&def, ret, ctx)
		if err != nil {
			return nil, err
		}
	}
	var sb strings.Builder
	for _, m := range ret.models {
		sb.WriteString(m)
		sb.WriteString("\n\n")
	}
	for _, s := range ret.services {
		for _, paramsStruct := range s.paramsStructs {
			sb.WriteString(paramsStruct)
			sb.WriteString("\n\n")
		}
		sb.WriteString(s.serviceInterface)
		sb.WriteString("\n\n")
		sb.WriteString(s.fiberInitializer)
		sb.WriteString("\n\n")
	}
	generator := generator.NewRoot(
		generator.NewComment(" THIS CODE WAS AUTO GENERATED"),
		generator.NewPackage(opts.Go.Package),
		generator.NewNewline(),
		generator.NewImport(opts.Go.FiberImport),
		generator.NewRawStatement(sb.String()),
	)
	r, err := generator.Gofmt("-s").Generate(0)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (g *GoGenerator) generate(def *GenDef, ret *defOutput, ctx *Context) error {
	for modelName, m := range def.Models {
		mGen, err := generateModel(modelName, m, ctx)
		if err != nil {
			return err
		}
		ret.models = append(ret.models, *mGen)
	}

	for serviceName, s := range def.Services {
		sGen, err := generateService(serviceName, s, ctx)
		if err != nil {
			return err
		}
		ret.services = append(ret.services, *sGen)
	}
	return nil
}

func addFieldWithJson(fieldName string, t string, s *generator.Struct) *generator.Struct {
	return s.AddField(
		capitalFirst(fieldName),
		t,
		`json:"`+fieldName+`"`,
	)
}

func generateModel(name string, m model, ctx *Context) (*string, error) {
	modelStruct := generator.NewStruct(name)
	for memberName, memberType := range m {
		t, err := generateType(memberType, ctx)
		if err != nil {
			return nil, err
		}
		modelStruct = addFieldWithJson(memberName, *t, modelStruct)
	}
	ret, err := modelStruct.Generate(0)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func generateService(name string, s service, ctx *Context) (*serviceOutput, error) {
	ret := serviceOutput{paramsStructs: make([]string, 0)}
	serviceInterface := generator.NewInterface(name)
	fiberInitializer := generator.NewFunc(
		nil,
		generator.NewFuncSignature("initializeFiber"+name).
			AddParameters(generator.NewFuncParameter("app", "*fiber.App")).
			AddParameters(generator.NewFuncParameter("service", name)),
	)
	for funcName, funcArgsOrReturn := range s {
		paramsStructName := funcName + "Params"
		paramsStruct := generator.NewStruct(paramsStructName)
		sig := generator.NewFuncSignature(funcName)
		fields := make([]string, 0)
		for paramNameOrReturn, raw := range funcArgsOrReturn {
			t, err := generateType(raw, ctx)
			if err != nil {
				return nil, err
			}
			if paramNameOrReturn == "return" {
				sig = sig.AddReturnTypes(*t)
			} else {
				sig = sig.AddParameters(generator.NewFuncParameter(paramNameOrReturn, *t))
				paramsStruct = addFieldWithJson(paramNameOrReturn, *t, paramsStruct)
				fields = append(fields, paramNameOrReturn)
			}
		}
		sig = sig.AddReturnTypes("error")
		serviceInterface = serviceInterface.AddSignatures(sig)
		paramsStructStr, err := paramsStruct.Generate(0)
		if err != nil {
			return nil, err
		}
		ret.paramsStructs = append(ret.paramsStructs, paramsStructStr)
		fiberInitializer = fiberInitializer.AddStatements(generateFiberInit(name, funcName, paramsStructName, fields), generator.NewNewline())
	}
	serviceInterfaceStr, err := serviceInterface.Generate(0)
	if err != nil {
		return nil, err
	}
	ret.serviceInterface = serviceInterfaceStr

	fiberInitializerStr, err := fiberInitializer.Generate(0)
	if err != nil {
		return nil, err
	}
	ret.fiberInitializer = fiberInitializerStr

	return &ret, nil
}

func generateFiberInit(serviceName string, funcName string, paramsStructName string, fieldNames []string) generator.Statement {
	varName := "var" + paramsStructName
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s.%s(", serviceName, funcName))
	for i, f := range fieldNames {
		sb.WriteString(fmt.Sprintf("%s.%s", varName, capitalFirst(f)))
		if i != len(fieldNames)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")
	serviceCall := sb.String()
	errStr := `
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}`
	stmt := fmt.Sprintf(`app.Post("/%s.%s", func(c *fiber.Ctx) error {`, serviceName, funcName) +
		fmt.Sprintf("\n%s := new(%s)", varName, paramsStructName) +
		fmt.Sprintf("\nerr := c.BodyParser(%s)", varName) +
		errStr +
		fmt.Sprintf("\nret, err := %s", serviceCall) +
		errStr +
		"\nreturn c.JSON(ret)" +
		"\n})"
	println(stmt)
	return generator.NewRawStatement(stmt)
}

func generateType(raw rawType, ctx *Context) (*string, error) {
	t := string(raw)
	c, err := parseCompoundType(t, ctx)
	if err != nil {
		return nil, err
	}
	var ret strings.Builder
	generateGoType(c, &ret)
	r := ret.String()
	return &r, nil
}

func generateGoType(t anyType, ret *strings.Builder) {
	switch t := t.(type) {
	case string:
		ret.WriteString(t)
	case array:
		{
			ret.WriteString("[]")
			generateGoType(t.of, ret)
		}
	case mapType:
		{
			ret.WriteString("map[" + string(t.from) + "]")
			generateGoType(t.to, ret)
		}
	}
}
