package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// expandDepth is how many levels of nesting get one line per member. Deeper
// values stay on one line, which puts each viewer's game object, each
// state group and each deck on a single diffable line.
const expandDepth = 5

// Marshal renders the Document as the bytes that are checked in. The output
// is deterministic: struct fields in declaration order, embedded JSON with
// sorted keys, and a fixed layout.
func (d *Document) Marshal() ([]byte, error) {
	var compact bytes.Buffer
	encoder := json.NewEncoder(&compact)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(d); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	decoder := json.NewDecoder(&compact)
	decoder.UseNumber()
	if err := layout(decoder, &out, 0); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

// layout copies one JSON value from decoder to out, preserving key order.
func layout(decoder *json.Decoder, out *bytes.Buffer, depth int) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, isDelim := token.(json.Delim)
	if !isDelim {
		return writeScalar(out, token)
	}
	closing := byte('}')
	if delim == '[' {
		closing = ']'
	}
	expand := depth < expandDepth
	out.WriteByte(byte(delim))
	for count := 0; decoder.More(); count++ {
		if count > 0 {
			out.WriteByte(',')
		}
		if expand {
			out.WriteByte('\n')
			for i := 0; i <= depth; i++ {
				out.WriteString(" ")
			}
		}
		if delim == '{' {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			if err := writeScalar(out, key); err != nil {
				return err
			}
			out.WriteByte(':')
			if expand {
				out.WriteByte(' ')
			}
		}
		if err := layout(decoder, out, depth+1); err != nil {
			return err
		}
		if expand && !decoder.More() {
			out.WriteByte('\n')
			for i := 0; i < depth; i++ {
				out.WriteString(" ")
			}
		}
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	out.WriteByte(closing)
	return nil
}

func writeScalar(out *bytes.Buffer, token json.Token) error {
	switch value := token.(type) {
	case json.Number:
		out.WriteString(value.String())
		return nil
	case nil:
		out.WriteString("null")
		return nil
	case string, bool:
		var scalar bytes.Buffer
		encoder := json.NewEncoder(&scalar)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(value); err != nil {
			return err
		}
		out.Write(bytes.TrimRight(scalar.Bytes(), "\n"))
		return nil
	default:
		return fmt.Errorf("unexpected JSON token %v", token)
	}
}
