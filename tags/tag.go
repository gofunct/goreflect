package tags

import (
	"errors"
	"fmt"
	"strconv"
)

type MultiTag struct {
	value string
	cache map[string][]string
}

func NewMultiTag(v string) MultiTag {
	return MultiTag{
		value: v,
	}
}

func (x *MultiTag) Scan() (map[string][]string, error) {

	v := x.value

	ret := make(map[string][]string)

	// This is mostly copied from reflect.StructTag.Get
	for v != "" {
		i := 0

		// Skip whitespace
		for i < len(v) && v[i] == ' ' {
			i++
		}
		// value is equal to an array with the length of the number of kv pairs found in the struct tag
		v = v[i:]

		// if value is empty, break
		if v == "" {
			break
		}

		//i is the first kv pair in the struct tag
		i = 0
		// Scan to colon to find key
		// while i is less than the number of the kv pairs in the struct
		// && the kv is not whitespace
		// && kv is not a colon && kv is not a quotation mark
		// add a
		for i < len(v) && v[i] != ' ' && v[i] != ':' && v[i] != '"' {
			i++
		}

		if i >= len(v) {
			return nil, errors.New(fmt.Sprintf("expected `:' after key name, but got end of tag (in `%v`)", x.value))
		}

		if v[i] != ':' {
			return nil, errors.New(fmt.Sprintf("expected `:' after key name, but got `%v' (in `%v`)", v[i], x.value))
		}

		if i+1 >= len(v) {
			return nil, errors.New(fmt.Sprintf("expected `\"' to start tag value at end of tag (in `%v`)", x.value))
		}

		if v[i+1] != '"' {
			return nil, errors.New(fmt.Sprintf("expected `\"' to start tag value, but got `%v' (in `%v`)", v[i+1], x.value))
		}

		name := v[:i]
		v = v[i+1:]

		// Scan quoted string to find value
		i = 1

		for i < len(v) && v[i] != '"' {
			if v[i] == '\n' {
				return nil, errors.New(fmt.Sprintf("expected end of tag value `\"' at end of tag (in `%v`)", x.value))
			}

			if v[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(v) {
			return nil, errors.New(fmt.Sprintf("expected end of tag value `\"' at end of tag (in `%v`)", x.value))
		}

		val, err := strconv.Unquote(v[:i+1])

		if err != nil {
			return nil, errors.New(fmt.Sprintf("Malformed value of tag `%v:%v` => %v (in `%v`)", name, v[:i+1], err, x.value))
		}

		v = v[i+1:]

		ret[name] = append(ret[name], val)
	}

	return ret, nil
}

func (x *MultiTag) Parse() error {
	vals, err := x.Scan()
	x.cache = vals

	return err
}

func (x *MultiTag) Cached() map[string][]string {
	if x.cache == nil {
		cache, _ := x.Scan()

		if cache == nil {
			cache = make(map[string][]string)
		}

		x.cache = cache
	}

	return x.cache
}

func (x *MultiTag) Get(key string) string {
	c := x.Cached()

	if v, ok := c[key]; ok {
		return v[len(v)-1]
	}

	return ""
}

func (x *MultiTag) GetMany(key string) []string {
	c := x.Cached()
	return c[key]
}

func (x *MultiTag) Set(key string, value string) {
	c := x.Cached()
	c[key] = []string{value}
}

func (x *MultiTag) SetMany(key string, value []string) {
	c := x.Cached()
	c[key] = value
}
