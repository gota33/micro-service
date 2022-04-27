package entity

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

type FieldMask struct {
	Paths []string `json:"paths,omitempty"`
}

func (mask FieldMask) ToMap(e any) (m map[string]any, err error) {
	var (
		data    []byte
		mFields map[string]any
	)
	if data, err = json.Marshal(e); err != nil {
		return
	}
	if err = json.Unmarshal(data, &mFields); err != nil {
		return
	}

	if size := len(mask.Paths); size > 0 {
		m = make(map[string]any, size)
		for _, path := range mask.Paths {
			if value, ok := mFields[path]; ok {
				m[path] = value
			}
		}
	} else {
		delete(mFields, "id")
		m = mFields
	}
	return
}

type ListRequestFragment struct {
	PageSize          int    `query:"pageSize"`
	PageToken         string `query:"pageToken"`
	FallbackPageSize  int
	FallbackPageToken string
}

func (r ListRequestFragment) GetPageSize() int {
	if r.PageSize > 0 {
		return r.PageSize
	}
	if r.FallbackPageSize > 0 {
		return r.FallbackPageSize
	}
	return 20
}

func (r ListRequestFragment) GetPageToken() string {
	if r.PageToken != "" {
		return r.PageToken
	}
	if r.FallbackPageToken != "" {
		return r.FallbackPageToken
	}
	return "0"
}

type ListResponseFragment struct {
	NextPageToken string `json:"nextPageToken"`
}

type UpdateRequestFragment struct {
	UpdateMask FieldMask `json:"updateMask"`
}

func SQLUpdate(table string, id any, fields map[string]any) (script string, args []any) {
	var sb strings.Builder
	sb.WriteString("update ")
	sb.WriteString(table)
	sb.WriteString(" set ")
	args = append(args, mapJoin(&sb, fields, ", ")...)
	sb.WriteString(" where id = ?")
	args = append(args, id)
	script = sb.String()
	return
}

func mapJoin(sb io.StringWriter, m map[string]any, sep string) (args []any) {
	var flag bool
	for field, value := range m {
		if flag {
			sb.WriteString(sep)
		} else {
			flag = true
		}
		sb.WriteString(field)
		sb.WriteString(" = ?")
		args = append(args, value)
	}
	return
}
