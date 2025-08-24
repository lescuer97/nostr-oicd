package templ

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var templates *template.Template

func LoadTemplates(dir string) error {
	var err error
	templates, err = template.ParseGlob(filepath.Join(dir, "*.html"))
	return err
}

func Render(w http.ResponseWriter, name string, data interface{}) error {
	if templates == nil {
		if err := LoadTemplates("./templates"); err != nil {
			return err
		}
	}
	return templates.ExecuteTemplate(w, name, data)
}
