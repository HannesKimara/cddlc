package gogen

import (
	"fmt"
)

type Options struct {
	EmitIdent        bool   `json:"emit-ident-info" yaml:"emit-ident-info"`
	CaseConversion   string `json:"case-conversion" yaml:"case-conversion"`
	GeneratePlugStub bool   `json:"generate-plug-stub" yaml:"generate-plug-stub"`
}

func (o Options) Valid() error {
	if _, ok := converters[o.CaseConversion]; !ok {
		return fmt.Errorf("unknown case-converion `%s`", o.CaseConversion)
	}
	return nil
}
