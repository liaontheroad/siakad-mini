package helper

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	if err := v.RegisterValidation("maxyear", func(fl validator.FieldLevel) bool {
		return fl.Field().Int() <= int64(time.Now().Year())
	}); err != nil {
		panic(err) 
	}

	return v
}

func Validate(s any) map[string][]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return map[string][]string{"body": {"tidak dapat divalidasi"}}
	}

	out := make(map[string][]string, len(verrs))
	for _, fe := range verrs {
		field := fe.Field()
		out[field] = append(out[field], messageFor(fe))
	}
	return out
}

func messageFor(fe validator.FieldError) string {
	isText := fe.Kind() == reflect.String

	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if isText {
			return fmt.Sprintf("minimal %s karakter", fe.Param())
		}
		return fmt.Sprintf("minimal %s", fe.Param())
	case "max":
		if isText {
			return fmt.Sprintf("maksimal %s karakter", fe.Param())
		}
		return fmt.Sprintf("maksimal %s", fe.Param())
	case "len":
		return fmt.Sprintf("harus tepat %s karakter", fe.Param())
	case "numeric":
		return "hanya boleh berisi angka"
	case "gte":
		return fmt.Sprintf("minimal %s", fe.Param())
	case "lte":
		return fmt.Sprintf("maksimal %s", fe.Param())
	case "maxyear":
		return "tidak boleh melebihi tahun berjalan"
	case "oneof":
		return "harus salah satu dari: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "tidak valid"
	}
}
