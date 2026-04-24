package marshal

import (
	"reflect"
	"strings"
)

type fieldTag struct {
	Name      string
	Format    string
	Separator string
	OmitEmpty bool
	Required  bool
	Skip      bool
}

func parseTag(tag reflect.StructTag, fieldName string) fieldTag {
	raw := tag.Get("csv")
	if raw == "-" {
		return fieldTag{Skip: true, Name: fieldName}
	}
	ft := fieldTag{Name: fieldName, Separator: ","}
	if raw == "" {
		return ft
	}
	parts := strings.Split(raw, ",")
	if len(parts) > 0 && parts[0] != "" {
		ft.Name = parts[0]
	}
	for _, opt := range parts[1:] {
		opt = strings.TrimSpace(opt)
		switch {
		case opt == "omitempty":
			ft.OmitEmpty = true
		case opt == "required":
			ft.Required = true
		case strings.HasPrefix(opt, "format="):
			ft.Format = strings.TrimPrefix(opt, "format=")
		case strings.HasPrefix(opt, "sep="):
			ft.Separator = strings.TrimPrefix(opt, "sep=")
		}
	}
	return ft
}
