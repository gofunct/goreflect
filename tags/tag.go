package tags

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"strconv"
)

func scanMultiTag(field string) (map[string][]string, error) {
	mtag := make(map[string][]string)
	// This is mostly copied from reflect.StructTag.Get
	for field != "" {
		i := 0

		// Skip whitespace
		for i < len(field) && field[i] == ' ' {
			i++
		}
		// value is equal to an array with the length of the number of kv pairs found in the struct tag
		field = field[i:]

		// if value is empty, break
		if field == "" {
			break
		}

		//i is the first kv pair in the struct tag
		i = 0
		// Scan to colon to find key
		// while i is less than the number of the kv pairs in the struct
		// && the kv is not whitespace
		// && kv is not a colon && kv is not a quotation mark
		// add a
		for i < len(field) && field[i] != ' ' && field[i] != ':' && field[i] != '"' {
			i++
		}

		if i >= len(field) {
			return nil, errors.New(fmt.Sprintf("expected `:' after key name, but got end of tag (in `%field`)", x.value))
		}

		if field[i] != ':' {
			return nil, errors.New(fmt.Sprintf("expected `:' after key name, but got `%field' (in `%field`)", field[i], x.value))
		}

		if i+1 >= len(field) {
			return nil, errors.New(fmt.Sprintf("expected `\"' to start tag value at end of tag (in `%field`)", x.value))
		}

		if field[i+1] != '"' {
			return nil, errors.New(fmt.Sprintf("expected `\"' to start tag value, but got `%field' (in `%field`)", field[i+1], x.value))
		}

		name := field[:i]
		field = field[i+1:]

		// Scan quoted string to find value
		i = 1

		for i < len(field) && field[i] != '"' {
			if field[i] == '\n' {
				return nil, errors.New(fmt.Sprintf("expected end of tag value `\"' at end of tag (in `%field`)", x.value))
			}

			if field[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(field) {
			return nil, errors.New(fmt.Sprintf("expected end of tag value `\"' at end of tag (in `%field`)", x.value))
		}

		val, err := strconv.Unquote(field[:i+1])

		if err != nil {
			return nil, errors.New(fmt.Sprintf("Malformed value of tag `%field:%field` => %field (in `%field`)", name, field[:i+1], err, x.value))
		}

		field = field[i+1:]

		mtag[name] = append(mtag[name], val)
	}

	return mtag, nil
}

func ParseMultiTag(field string) error {
	vals, err := scanMultiTag(field)
	viper.Set(field, vals)

	return err
}

func Cached(field string) map[string][]string {
	if viper.GetStringMapStringSlice(field) == nil {
		cache, _ := scanMultiTag(field)
		if cache == nil {
			cache = make(map[string][]string)
		}
		viper.Set(field, cache)
	}
	return viper.GetStringMapStringSlice(field)
}

func Get(field string) string {
	c := viper.GetStringMapStringSlice(field)

	if field, ok := c[field]; ok {
		return field[len(field)-1]
	}

	return ""
}

func GetMany(field string) []string {
	c := viper.GetStringMapStringSlice(field)
	return c[field]
}

func Set(field string, value string) {
	c := viper.GetStringMapStringSlice(field)
	c[field] = []string{value}
}

func SetMany(field string, value []string) {
	c := viper.GetStringMapStringSlice(field)
	c[field] = value
}
