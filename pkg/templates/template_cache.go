package templates

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sync"
)

const templateBasePath = "templates"
const templateFileExtension = "tmpl"

var templateCache = make(map[KnownTemplate]*CachedTemplate)
var templateCacheMutex sync.Mutex

// CachedTemplate represents a cached template
type CachedTemplate struct {
	templateName    KnownTemplate
	templateContent *template.Template
}

// GetTemplate returns a cached template instance according to the template name
func GetTemplate(templateName KnownTemplate) (*template.Template, error) {
	templateCacheMutex.Lock()
	defer templateCacheMutex.Unlock()
	fullPath := filepath.Join(templateBasePath, fmt.Sprintf("%s.%s", templateName, templateFileExtension))

	cachedTemplate, exists := templateCache[templateName]

	if exists {
		return cachedTemplate.templateContent, nil
	}

	tmpl, err := template.ParseFiles(fullPath)

	if err != nil {
		return nil, err
	}

	templateCache[templateName] = &CachedTemplate{
		templateName:    templateName,
		templateContent: tmpl,
	}

	return tmpl, err
}
