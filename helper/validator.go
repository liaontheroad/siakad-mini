package helper

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterValidation("maxyear", func(fl validator.FieldLevel) bool {
		year := fl.Field().Int()
		return year >= 1900 && year <= 2026
	})


	validate.RegisterValidation("tahunakademik", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		re := regexp.MustCompile(`^(\d{4})/(\d{4})-(Ganjil|Genap)$`)
		matches := re.FindStringSubmatch(val)
		if len(matches) != 4 {
			return false
		}

		var y1, y2 int
		fmt.Sscanf(matches[1], "%d", &y1)
		fmt.Sscanf(matches[2], "%d", &y2)

		return y2 == y1+1
	})
}

func Validate(payload any) map[string][]string {
	errs := make(map[string][]string)

	err := validate.Struct(payload)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string][]string{"general": {err.Error()}}
	}

	for _, e := range validationErrors {
		fieldName := getJSONFieldName(payload, e.Field())
		msg := messageFor(e)

		errs[fieldName] = append(errs[fieldName], msg)
	}

	return errs
}

func getJSONFieldName(structObj any, fieldName string) string {
	t := reflect.TypeOf(structObj)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return strings.ToLower(fieldName)
	}

	field, found := t.FieldByName(fieldName)
	if !found {
		return strings.ToLower(fieldName)
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return strings.ToLower(fieldName)
	}

	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

func messageFor(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "Field ini wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "min":
		return fmt.Sprintf("Minimal harus bernilai atau sepanjang %s", e.Param())
	case "max":
		return fmt.Sprintf("Maksimal bernilai atau sepanjang %s", e.Param())
	case "len":
		return "Harus tepat sepanjang 12 karakter"
	case "numeric":
		return "Harus berupa angka"
	case "gte":
		return fmt.Sprintf("Harus bernilai lebih dari atau sama dengan %s", e.Param())
	case "lte":
		return fmt.Sprintf("Harus bernilai kurang dari atau sama dengan %s", e.Param())
	case "maxyear":
		return "Angkatan tidak boleh melebihi tahun berjalan"
	case "tahunakademik":
		return "format harus 2026/2027-Ganjil atau 2026/2027-Genap"
	default:
		return fmt.Sprintf("Gagal pada validasi '%s'", e.Tag())
	}
}