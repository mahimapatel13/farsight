// NOT TO BE USED TILL MANAGING SERVER AND STORAGE INFRASTRUCTURE AS WELL.

package emailtypes

// import (
// 	"bytes"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	html "html/template"
// 	"os"
// 	"path/filepath"
// 	text "text/template"
// )

// // TemplateRenderer renders email templates in HTML, Text, and JSON formats
// type TemplateRenderer struct {
// 	templateDir string
// 	htmlCache   map[string]*html.Template
// 	textCache   map[string]*text.Template
// 	jsonCache   map[string]string // Stores JSON templates as raw string
// }

// // ErrTemplateNotFound is returned when a requested template is not found
// var ErrTemplateNotFound = errors.New("template not found")

// // NewTemplateRenderer creates a new TemplateRenderer instance with preloaded templates
// func NewTemplateRenderer(templateDir string) (*TemplateRenderer, error) {
// 	tr := &TemplateRenderer{
// 		templateDir: templateDir,
// 		htmlCache:   make(map[string]*html.Template),
// 		textCache:   make(map[string]*text.Template),
// 		jsonCache:   make(map[string]string),
// 	}

// 	// Preload HTML, Text, and JSON templates
// 	if err := tr.loadHTMLTemplates("html"); err != nil {
// 		return nil, fmt.Errorf("failed to load HTML templates: %w", err)
// 	}
// 	if err := tr.loadTextTemplates("text"); err != nil {
// 		return nil, fmt.Errorf("failed to load text templates: %w", err)
// 	}
// 	if err := tr.loadJSONTemplates("json"); err != nil {
// 		return nil, fmt.Errorf("failed to load JSON templates: %w", err)
// 	}

// 	return tr, nil
// }

// // loadHTMLTemplates loads and caches HTML templates
// func (tr *TemplateRenderer) loadHTMLTemplates(format string) error {
// 	templatePath := filepath.Join(tr.templateDir, format)
// 	files, err := os.ReadDir(templatePath)
// 	if err != nil {
// 		return fmt.Errorf("failed to read %s template directory: %w", format, err)
// 	}

// 	for _, file := range files {
// 		if !file.IsDir() && filepath.Ext(file.Name()) == ".html" {
// 			templateName := file.Name()
// 			tmplPath := filepath.Join(templatePath, templateName)
// 			tmpl, err := html.ParseFiles(tmplPath) // Corrected to use html alias
// 			if err != nil {
// 				return fmt.Errorf("failed to parse HTML template %s: %w", templateName, err)
// 			}
// 			tr.htmlCache[templateName] = tmpl
// 		}
// 	}
// 	return nil
// }

// // loadTextTemplates loads and caches Text templates
// func (tr *TemplateRenderer) loadTextTemplates(format string) error {
// 	templatePath := filepath.Join(tr.templateDir, format)
// 	files, err := os.ReadDir(templatePath)
// 	if err != nil {
// 		return fmt.Errorf("failed to read %s template directory: %w", format, err)
// 	}

// 	for _, file := range files {
// 		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
// 			templateName := file.Name()
// 			tmplPath := filepath.Join(templatePath, templateName)
// 			tmpl, err := text.ParseFiles(tmplPath) // Corrected to use text alias
// 			if err != nil {
// 				return fmt.Errorf("failed to parse text template %s: %w", templateName, err)
// 			}
// 			tr.textCache[templateName] = tmpl
// 		}
// 	}
// 	return nil
// }

// // loadJSONTemplates loads and caches JSON templates as raw strings
// func (tr *TemplateRenderer) loadJSONTemplates(format string) error {
// 	templatePath := filepath.Join(tr.templateDir, format)
// 	files, err := os.ReadDir(templatePath)
// 	if err != nil {
// 		return fmt.Errorf("failed to read %s template directory: %w", format, err)
// 	}

// 	for _, file := range files {
// 		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
// 			templateName := file.Name()
// 			tmplPath := filepath.Join(templatePath, templateName)
// 			content, err := os.ReadFile(tmplPath)
// 			if err != nil {
// 				return fmt.Errorf("failed to read %s JSON template: %w", templateName, err)
// 			}
// 			tr.jsonCache[templateName] = string(content)
// 		}
// 	}
// 	return nil
// }

// // Render renders a template based on format: "html", "text", or "json"
// func (tr *TemplateRenderer) Render(templateName string, data interface{}, format string) (string, error) {
// 	switch format {
// 	case "html":
// 		return tr.renderHTMLTemplate(templateName+".html", data)
// 	case "text":
// 		return tr.renderTextTemplate(templateName+".txt", data)
// 	case "json":
// 		return tr.renderJSONTemplate(templateName+".json", data)
// 	default:
// 		return "", fmt.Errorf("unsupported template format: %s", format)
// 	}
// }

// // renderHTMLTemplate renders HTML templates
// func (tr *TemplateRenderer) renderHTMLTemplate(templateName string, data interface{}) (string, error) {
// 	tmpl, ok := tr.htmlCache[templateName]
// 	if !ok {
// 		return "", ErrTemplateNotFound
// 	}

// 	var buf bytes.Buffer
// 	if err := tmpl.Execute(&buf, data); err != nil {
// 		return "", fmt.Errorf("failed to render HTML template %s: %w", templateName, err)
// 	}
// 	return buf.String(), nil
// }

// // renderTextTemplate renders Text templates
// func (tr *TemplateRenderer) renderTextTemplate(templateName string, data interface{}) (string, error) {
// 	tmpl, ok := tr.textCache[templateName]
// 	if !ok {
// 		return "", ErrTemplateNotFound
// 	}

// 	var buf bytes.Buffer
// 	if err := tmpl.Execute(&buf, data); err != nil {
// 		return "", fmt.Errorf("failed to render text template %s: %w", templateName, err)
// 	}
// 	return buf.String(), nil
// }

// // renderJSONTemplate renders JSON templates with dynamic data
// func (tr *TemplateRenderer) renderJSONTemplate(templateName string, data interface{}) (string, error) {
// 	rawTemplate, ok := tr.jsonCache[templateName]
// 	if !ok {
// 		return "", ErrTemplateNotFound
// 	}

// 	var templateData map[string]interface{}
// 	if err := json.Unmarshal([]byte(rawTemplate), &templateData); err != nil {
// 		return "", fmt.Errorf("failed to unmarshal JSON template %s: %w", templateName, err)
// 	}

// 	// Merge provided data into templateData
// 	for key, value := range data.(map[string]interface{}) {
// 		templateData[key] = value
// 	}

// 	renderedJSON, err := json.Marshal(templateData)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal rendered JSON: %w", err)
// 	}

// 	return string(renderedJSON), nil
// }

