package goreflect

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

const Unsupported = "UNSUPPORTED"

type TYPE int

const (
	JSON TYPE = iota
	YAML
	XML
	HCL
	PROTO
	TF
)

func (d TYPE) String() string {
	// declare an array of strings
	// ... operator counts how many
	// items in the array (7)
	names := [...]string{
		"JSON",
		"XML",
		"YAML"}

	if d < JSON || d > PROTO {
		panic(errors.New(Unsupported))
	}
	if d == PROTO {
		panic(errors.New(Unsupported))
	}
	return names[d]
}

func EncodeXML(v interface{}, w io.Writer) error {
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	defer w.Write([]byte("\n"))
	e := xml.NewEncoder(w)
	e.Indent("", "\t")
	return e.Encode(v)
}

func EncodeJSON(v interface{}, w io.Writer, pretty bool) error {
	if pretty {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		var out bytes.Buffer
		err = json.Indent(&out, b, "", "\t")
		if err != nil {
			return err
		}
		_, err = io.Copy(w, &out)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("\n"))
		return err
	}
	return json.NewEncoder(w).Encode(v)
}

func EncodeYAML(v interface{}, w io.Writer) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func ToYAML(v interface{}, r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, v)
}

func ToJSON(v interface{}, r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, v)
}

func InterFaceToXML(v interface{}, r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return xml.Unmarshal(b, v)
}

// toJson encodes an item into a JSON string
func ToJSONString(v interface{}) string {
	output, _ := json.Marshal(v)
	return string(output)
}

// toPrettyJson encodes an item into a pretty (indented) JSON string
func ToPrettyJSONString(v interface{}) string {
	output, _ := json.MarshalIndent(v, "", "  ")
	return string(output)
}

func MarshalReader(in io.Reader, data TYPE, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(in); err != nil {
		return errors.WithStack(err)
	}
	switch data {
	case YAML:
		if err := yaml.Unmarshal(buf.Bytes(), &c); err != nil {
			return errors.WithStack(err)
		}

	case JSON:
		if err := json.Unmarshal(buf.Bytes(), &c); err != nil {
			return errors.WithStack(err)
		}
	case XML:
		if err := xml.Unmarshal(buf.Bytes(), &c); err != nil {
			return errors.WithStack(err)
		}
	case HCL:
		obj, err := hcl.Parse(string(buf.Bytes()))
		if err != nil {
			return errors.WithStack(err)
		}
		if err = hcl.DecodeObject(&c, obj.Node); err != nil {
			return errors.WithStack(err)
		}
	case TF:
		dir := prompt("required: absolute path to your .hcldec file")
		if dir == "" {
			return errors.New(`found empty path, see: "https://github.com/hashicorp/hcl2/blob/master/cmd/hcldec/spec-format.md"`)
		}
		if err := json.Unmarshal(buf.Bytes(), &c); err != nil {
			return errors.WithStack(err)
		}
	}

	InsensitivizeMap(c)
	return nil
}

func MarshalWriter(w io.Writer, c map[string]interface{}, data TYPE) error {

	switch data {
	case JSON:
		b, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = fmt.Fprintf(w, "%s", b)
		if err != nil {
			return errors.WithStack(err)
		}

	case HCL:
		b, err := json.Marshal(c)
		ast, err := hcl.Parse(string(b))
		if err != nil {
			return errors.WithStack(err)
		}
		err = printer.Fprint(w, ast.Node)
		if err != nil {
			return errors.WithStack(err)
		}

	case YAML:
		b, err := yaml.Marshal(c)
		if err != nil {
			return errors.WithStack(err)
		}

		if _, err = fmt.Fprintf(w, "%s", b); err != nil {
			return errors.WithStack(err)
		}
	case TF:
		dir := prompt("please provide an absolute path to directory containing your .tf state files")
		b, err := runb("terraform", "output", "-state", dir, "-json")
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = fmt.Fprintf(w, "%s", b)
		if err != nil {
			return errors.WithStack(err)
		}

	}
	return nil
}
