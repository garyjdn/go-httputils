package httputils

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	apperror "github.com/garyjdn/go-apperror"
)

// ParseJSON parses JSON from request body into the provided struct
func ParseJSON(r *http.Request, v interface{}) *apperror.AppError {
	if r.Body == nil {
		return apperror.NewAppError(http.StatusBadRequest, "Request body is empty", nil)
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return apperror.NewAppError(http.StatusBadRequest, "Invalid JSON format: "+err.Error(), nil)
	}

	return nil
}

// ValidateStruct validates a struct using field tags
func ValidateStruct(v interface{}) *apperror.AppError {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check for required tag
		requiredTag := fieldType.Tag.Get("required")
		if requiredTag == "true" {
			if isZeroValue(field) {
				fieldName := getFieldName(fieldType)
				return apperror.NewAppError(http.StatusBadRequest, fieldName+" is required", nil)
			}
		}

		// Check for min tag (for strings and slices)
		minTag := fieldType.Tag.Get("min")
		if minTag != "" {
			minLength, err := strconv.Atoi(minTag)
			if err == nil {
				if field.Kind() == reflect.String {
					if len(field.String()) < minLength {
						fieldName := getFieldName(fieldType)
						return apperror.NewAppError(http.StatusBadRequest, fieldName+" must be at least "+minTag+" characters long", nil)
					}
				} else if field.Kind() == reflect.Slice {
					if field.Len() < minLength {
						fieldName := getFieldName(fieldType)
						return apperror.NewAppError(http.StatusBadRequest, fieldName+" must have at least "+minTag+" items", nil)
					}
				}
			}
		}
	}

	return nil
}

// GetPageParam extracts page parameter from query string with default value of 1
func GetPageParam(r *http.Request) int {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		return 1
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 1
	}

	return page
}

// GetLimitParam extracts limit parameter from query string with default value of 10
func GetLimitParam(r *http.Request) int {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return 10
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		return 10
	}

	return limit
}

// Helper function to check if a value is zero
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

// Helper function to get field name from struct tag or field name
func getFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		// Remove omitempty if present
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}
		if jsonTag != "" && jsonTag != "-" {
			return jsonTag
		}
	}
	return field.Name
}
