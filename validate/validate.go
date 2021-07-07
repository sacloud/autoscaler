// Copyright 2021 The sacloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
)

var validatorInstance = validator.New()

func validate(v interface{}) error {
	validatorInstance.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("name"), ",", 2)[0]
		if name == "" {
			// nameタグがない場合はyamlタグを参照
			name = strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
		}
		if name == "-" {
			return ""
		}
		return name
	})
	return validatorInstance.Struct(v)
}

func InitValidatorAlias(zones []string) {
	validatorInstance.RegisterAlias("zone", fmt.Sprintf("oneof=%s", strings.Join(zones, " ")))
	validatorInstance.RegisterAlias("zones", "dive,zone")
}

func Struct(v interface{}) error {
	err := validate(v)
	if err != nil {
		if err != nil {
			// see https://github.com/go-playground/validator/blob/f6584a41c8acc5dfc0b62f7962811f5231c11530/_examples/simple/main.go#L59-L65
			if _, ok := err.(*validator.InvalidValidationError); ok {
				return err
			}

			errors := &multierror.Error{}
			for _, err := range err.(validator.ValidationErrors) {
				errors = multierror.Append(errors, errorFromValidationErr(v, err))
			}
			return errors.ErrorOrNil()
		}
	}

	return nil
}

func StructWithMultiError(v interface{}) []error {
	err := validate(v)
	if err != nil {
		if err != nil {
			// see https://github.com/go-playground/validator/blob/f6584a41c8acc5dfc0b62f7962811f5231c11530/_examples/simple/main.go#L59-L65
			if _, ok := err.(*validator.InvalidValidationError); ok {
				return []error{err}
			}

			errors := &multierror.Error{}
			for _, err := range err.(validator.ValidationErrors) {
				errors = multierror.Append(errors, errorFromValidationErr(v, err))
			}
			return errors.Errors
		}
	}

	return nil
}

func errorFromValidationErr(target interface{}, err validator.FieldError) error {
	namespaces := strings.Split(err.Namespace(), ".")
	actualName := namespaces[len(namespaces)-1] // .で区切った末尾の要素

	param := err.Param()
	detail := err.ActualTag()
	if param != "" {
		detail += "=" + param
	}

	// detailがvalidatorのタグ名だけの場合の対応をここで行う。
	switch detail {
	case "file":
		detail = fmt.Sprintf("invalid file path: %v", err.Value())
	}

	return newError(actualName, detail)
}

func newError(name, message string) error {
	return fmt.Errorf("%s: %s", name, message)
}
