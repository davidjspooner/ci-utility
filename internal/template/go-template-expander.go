package template

import (
	"fmt"
	htmlTemplate "html/template"
	"io"

	textTemplate "text/template"
)

// expandGoTextTemplate parses and executes a Go text/template with the provided content
// and writes the result to the given io.Writer. It supports templateFunctions.
func expandGoTextTemplate(content string, w io.Writer, values *Values) error {
	tmpl, err := textTemplate.New("template").Funcs(values.Functions()).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse text template: %w", err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		return fmt.Errorf("failed to expand text template: %w", err)
	}
	return nil
}

// expandGoHTMLTemplate parses and executes a Go html/template with the provided content
// and writes the result to the given io.Writer. It supports templateFunctions.
func expandGoHTMLTemplate(content string, w io.Writer, values *Values) error {
	tmpl, err := htmlTemplate.New("template").Funcs(values.Functions()).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		return fmt.Errorf("failed to expand HTML template: %w", err)
	}
	return nil
}
