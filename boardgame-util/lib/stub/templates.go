package stub

import (
	"errors"
	"text/template"
)

//TemplateSet is a collection of templates that can create a derived and
//expanded FileContents when given an Options struct.
type TemplateSet map[string]*template.Template

//DefaultTemplateSet returns the default template set for this stub.
func DefaultTemplateSet() TemplateSet {
	return nil
}

//Generate generates FileContents based on this TemplateSet, using those
//options to expand. Names of files will also be run through templates and
//expanded.
func (t TemplateSet) Generate(opt *Options) (FileContents, error) {
	return nil, errors.New("Not yet implemented")
}
