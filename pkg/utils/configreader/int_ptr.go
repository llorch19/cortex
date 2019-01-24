/*
Copyright 2019 Cortex Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configreader

import (
	"io/ioutil"

	s "github.com/cortexlabs/cortex/pkg/api/strings"
	"github.com/cortexlabs/cortex/pkg/utils/cast"
	"github.com/cortexlabs/cortex/pkg/utils/errors"
)

type IntPtrValidation struct {
	Required             bool
	Default              *int
	DisallowNull         bool
	AllowedValues        []int
	GreaterThan          *int
	GreaterThanOrEqualTo *int
	LessThan             *int
	LessThanOrEqualTo    *int
	Validator            func(*int) (*int, error)
}

func makeIntValValidation(v *IntPtrValidation) *IntValidation {
	return &IntValidation{
		AllowedValues:        v.AllowedValues,
		GreaterThan:          v.GreaterThan,
		GreaterThanOrEqualTo: v.GreaterThanOrEqualTo,
		LessThan:             v.LessThan,
		LessThanOrEqualTo:    v.LessThanOrEqualTo,
	}
}

func IntPtr(inter interface{}, v *IntPtrValidation) (*int, error) {
	if inter == nil {
		return ValidateIntPtr(nil, v)
	}
	casted, castOk := cast.InterfaceToInt(inter)
	if !castOk {
		return nil, errors.New(s.ErrInvalidPrimitiveType(inter, s.PrimTypeInt))
	}
	return ValidateIntPtr(&casted, v)
}

func IntPtrFromInterfaceMap(key string, iMap map[string]interface{}, v *IntPtrValidation) (*int, error) {
	inter, ok := ReadInterfaceMapValue(key, iMap)
	if !ok {
		val, err := ValidateIntPtrMissing(v)
		if err != nil {
			return nil, errors.Wrap(err, key)
		}
		return val, nil
	}
	val, err := IntPtr(inter, v)
	if err != nil {
		return nil, errors.Wrap(err, key)
	}
	return val, nil
}

func IntPtrFromStrMap(key string, sMap map[string]string, v *IntPtrValidation) (*int, error) {
	valStr, ok := sMap[key]
	if !ok || valStr == "" {
		val, err := ValidateIntPtrMissing(v)
		if err != nil {
			return nil, errors.Wrap(err, key)
		}
		return val, nil
	}
	val, err := IntPtrFromStr(valStr, v)
	if err != nil {
		return nil, errors.Wrap(err, key)
	}
	return val, nil
}

func IntPtrFromStr(valStr string, v *IntPtrValidation) (*int, error) {
	if valStr == "" {
		return ValidateIntPtrMissing(v)
	}
	casted, castOk := s.ParseInt(valStr)
	if !castOk {
		return nil, errors.New(s.ErrInvalidPrimitiveType(valStr, s.PrimTypeInt))
	}
	return ValidateIntPtr(&casted, v)
}

func IntPtrFromEnv(envVarName string, v *IntPtrValidation) (*int, error) {
	valStr := ReadEnvVar(envVarName)
	if valStr == nil || *valStr == "" {
		val, err := ValidateIntPtrMissing(v)
		if err != nil {
			return nil, errors.Wrap(err, s.EnvVar(envVarName))
		}
		return val, nil
	}
	val, err := IntPtrFromStr(*valStr, v)
	if err != nil {
		return nil, errors.Wrap(err, s.EnvVar(envVarName))
	}
	return val, nil
}

func IntPtrFromFile(filePath string, v *IntPtrValidation) (*int, error) {
	valBytes, err := ioutil.ReadFile(filePath)
	if err != nil || len(valBytes) == 0 {
		val, err := ValidateIntPtrMissing(v)
		if err != nil {
			return nil, errors.Wrap(err, filePath)
		}
		return val, nil
	}
	valStr := string(valBytes)
	val, err := IntPtrFromStr(valStr, v)
	if err != nil {
		return nil, errors.Wrap(err, filePath)
	}
	return val, nil
}

func IntPtrFromEnvOrFile(envVarName string, filePath string, v *IntPtrValidation) (*int, error) {
	valStr := ReadEnvVar(envVarName)
	if valStr != nil && *valStr != "" {
		return IntPtrFromEnv(envVarName, v)
	}
	return IntPtrFromFile(filePath, v)
}

func IntPtrFromPrompt(promptOpts *PromptOptions, v *IntPtrValidation) (*int, error) {
	valStr := prompt(promptOpts)
	if valStr == "" {
		return ValidateIntPtrMissing(v)
	}
	return IntPtrFromStr(valStr, v)
}

func ValidateIntPtrMissing(v *IntPtrValidation) (*int, error) {
	if v.Required {
		return nil, errors.New(s.ErrMustBeDefined)
	}
	return ValidateIntPtr(v.Default, v)
}

func ValidateIntPtr(val *int, v *IntPtrValidation) (*int, error) {
	if v.DisallowNull {
		if val == nil {
			return nil, errors.New(s.ErrCannotBeNull)
		}
	}

	if val != nil {
		err := ValidateIntVal(*val, makeIntValValidation(v))
		if err != nil {
			return nil, err
		}
	}

	if v.Validator != nil {
		return v.Validator(val)
	}
	return val, nil
}
