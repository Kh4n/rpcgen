package parsergenerator

import (
	"errors"
	"fmt"
)

type GenDef struct {
	Models   map[string]model
	Services map[string]service
}
type model map[string]rawType
type service map[string]funcDef
type funcDef map[string]rawType

func validateGenDef(g *GenDef) error {
	for modelName, model := range g.Models {
		if isKeyword(modelName) {
			return isKeywordError(modelName)
		}
		if err := validateModel(model); err != nil {
			return err
		}
	}
	for serviceName, service := range g.Services {
		if isKeyword(serviceName) {
			return isKeywordError(serviceName)
		}
		if err := validateService(service); err != nil {
			return err
		}
	}
	return nil
}

func validateFields(fields map[string]rawType) error {
	seen := make(map[string]interface{})
	for f := range fields {
		if _, present := seen[f]; present {
			return errors.New("duplicate field name: " + f)
		}
		cap := capitalFirst(f)
		if _, present := seen[f]; present {
			return fmt.Errorf("duplicate field name (first letter is case insensitive): %s (%s)", f, cap)
		}
	}
	return nil
}

func validateModel(m model) error {
	err := validateFields(m)
	if err != nil {
		return err
	}
	for fieldName := range m {
		if isKeyword(fieldName) {
			return isKeywordError(fieldName)
		}
	}
	return nil
}

func validateService(s service) error {
	for funcName, funcDef := range s {
		if isKeyword(funcName) {
			return isKeywordError(funcName)
		}
		if len(funcDef) == 0 {
			return errors.New(funcName + " needs at least one argument or 'return'")
		}
		if err := validateFunction(funcDef); err != nil {
			return err
		}
	}
	return nil
}

func validateFunction(f funcDef) error {
	err := validateFields(f)
	if err != nil {
		return err
	}
	for argOrReturn := range f {
		if isKeyword(argOrReturn) && argOrReturn != "return" {
			return isKeywordError(argOrReturn)
		}
	}
	return nil
}
