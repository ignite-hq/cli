package clidoc

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	// Docs represents a slice of Doc.
	Docs []Doc
	// Doc represents the struct documentation with tag comments.
	Doc struct {
		Key     string
		Type    string
		Value   Docs
		Comment string
	}
)

// String converts Docs to a string.
func (d Docs) String() string {
	var sb strings.Builder
	// Initial call with a negative level to avoid unwanted dash at the top level
	d.writeString(&sb, -1)
	return strings.TrimSpace(sb.String())
}

// writeString appends the contents of Docs to sb's buffer at level.
func (d Docs) writeString(sb *strings.Builder, level int) {
	indent := strings.Repeat("  ", level+1) // Two spaces per YAML indentation standard
	for _, doc := range d {
		sb.WriteString(indent)
		if doc.Type != "" {
			sb.WriteString(fmt.Sprintf("%s: [%s] # %s\n", doc.Key, doc.Type, doc.Comment))
		} else {
			sb.WriteString(fmt.Sprintf("%s: # %s\n", doc.Key, doc.Comment))
		}
		if len(doc.Value) > 0 {
			doc.Value.writeString(sb, level+1)
		}
	}
}

// GenDoc to generate documentation from a struct.
func GenDoc(v interface{}) (fields Docs, err error) {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Ptr {
		return fields, nil
	}
	for i := 0; i < t.NumField(); i++ {
		var (
			field = t.Field(i)
			doc   = field.Tag.Get("doc")
			yaml  = field.Tag.Get("yaml")
		)

		tags := strings.Split(yaml, ",")
		name := tags[0]
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		if len(tags) > 1 && strings.Contains(tags[1], "inline") {
			elemFields, err := GenDoc(reflect.New(field.Type).Elem().Interface())
			if err != nil {
				return nil, err
			}
			fields = append(fields, elemFields...)
			continue
		}

		var (
			elemFields Docs
			elemType   string
		)
		switch field.Type.Kind() { //nolint:exhaustive
		case reflect.Struct:
			elemType = field.Type.Kind().String()
			elemFields, err = GenDoc(reflect.New(field.Type).Elem().Interface())
			if err != nil {
				return nil, err
			}
		case reflect.Ptr:
			elemType = field.Type.Elem().Kind().String()
			elemFields, err = GenDoc(reflect.New(field.Type.Elem()).Elem().Interface())
			if err != nil {
				return nil, err
			}
		case reflect.Slice:
			elemType = fmt.Sprintf("[]%s", field.Type.Elem().Kind().String())
			elemFields, err = GenDoc(reflect.New(field.Type.Elem()).Elem().Interface())
			if err != nil {
				return nil, err
			}
		default:
			elemType = field.Type.Kind().String()
		}
		fields = append(fields, Doc{
			Key:     name,
			Comment: doc,
			Value:   elemFields,
			Type:    mapTypes(elemType),
		})
	}

	return fields, nil
}

func mapTypes(doc string) string {
	docTypes := map[string]string{
		"[]struct": "list",
		"struct":   "",
		"[]map":    "list",
		"map":      "",
		"[]slice":  "list",
		"slice":    "list",
	}
	if docType, ok := docTypes[doc]; ok {
		return docType
	}
	return doc
}
