package templates

import (
	"bytes"
	"fmt"
	"html/template"
	textTemplate "text/template"

	"github.com/video-converter/notification/internal/models"
)

// Engine handles email template rendering
type Engine struct {
	htmlTemplates map[models.NotificationType]*template.Template
	textTemplates map[models.NotificationType]*textTemplate.Template
}

// NewEngine creates a new template engine
func NewEngine() *Engine {
	engine := &Engine{
		htmlTemplates: make(map[models.NotificationType]*template.Template),
		textTemplates: make(map[models.NotificationType]*textTemplate.Template),
	}
	
	// Load built-in templates
	engine.loadBuiltinTemplates()
	
	return engine
}

// RenderTemplate renders an email template with the given data
func (e *Engine) RenderTemplate(notificationType models.NotificationType, data map[string]interface{}) (*models.EmailTemplate, error) {
	htmlTemplate, htmlExists := e.htmlTemplates[notificationType]
	textTemplate, textExists := e.textTemplates[notificationType]
	
	if !htmlExists && !textExists {
		return nil, fmt.Errorf("no template found for notification type: %s", notificationType)
	}
	
	result := &models.EmailTemplate{}
	
	// Render HTML template
	if htmlExists {
		var htmlBuf bytes.Buffer
		if err := htmlTemplate.Execute(&htmlBuf, data); err != nil {
			return nil, fmt.Errorf("failed to render HTML template: %w", err)
		}
		result.HTMLBody = htmlBuf.String()
		
		// Extract subject from HTML template if available
		if subjectTemplate := htmlTemplate.Lookup("subject"); subjectTemplate != nil {
			var subjectBuf bytes.Buffer
			if err := subjectTemplate.Execute(&subjectBuf, data); err == nil {
				result.Subject = subjectBuf.String()
			}
		}
	}
	
	// Render text template
	if textExists {
		var textBuf bytes.Buffer
		if err := textTemplate.Execute(&textBuf, data); err != nil {
			return nil, fmt.Errorf("failed to render text template: %w", err)
		}
		result.TextBody = textBuf.String()
		
		// Extract subject from text template if HTML didn't provide one
		if result.Subject == "" {
			if subjectTemplate := textTemplate.Lookup("subject"); subjectTemplate != nil {
				var subjectBuf bytes.Buffer
				if err := subjectTemplate.Execute(&subjectBuf, data); err == nil {
					result.Subject = subjectBuf.String()
				}
			}
		}
	}
	
	// Set default subject if none was extracted
	if result.Subject == "" {
		result.Subject = e.getDefaultSubject(notificationType)
	}
	
	return result, nil
}

// loadBuiltinTemplates loads the built-in email templates
func (e *Engine) loadBuiltinTemplates() {
	// Welcome email templates
	e.htmlTemplates[models.NotificationTypeWelcome] = template.Must(template.New("welcome").Parse(welcomeHTMLTemplate))
	e.textTemplates[models.NotificationTypeWelcome] = textTemplate.Must(textTemplate.New("welcome").Parse(welcomeTextTemplate))
	
	// Conversion complete templates
	e.htmlTemplates[models.NotificationTypeConversionComplete] = template.Must(template.New("conversion_complete").Parse(conversionCompleteHTMLTemplate))
	e.textTemplates[models.NotificationTypeConversionComplete] = textTemplate.Must(textTemplate.New("conversion_complete").Parse(conversionCompleteTextTemplate))
	
	// Conversion error templates
	e.htmlTemplates[models.NotificationTypeConversionError] = template.Must(template.New("conversion_error").Parse(conversionErrorHTMLTemplate))
	e.textTemplates[models.NotificationTypeConversionError] = textTemplate.Must(textTemplate.New("conversion_error").Parse(conversionErrorTextTemplate))
}

// getDefaultSubject returns a default subject for the notification type
func (e *Engine) getDefaultSubject(notificationType models.NotificationType) string {
	switch notificationType {
	case models.NotificationTypeWelcome:
		return "Welcome to Video Converter!"
	case models.NotificationTypeConversionComplete:
		return "Your video conversion is complete"
	case models.NotificationTypeConversionError:
		return "Video conversion failed"
	default:
		return "Notification from Video Converter"
	}
}

// AddCustomTemplate adds a custom template for a notification type
func (e *Engine) AddCustomTemplate(notificationType models.NotificationType, htmlTemplate, textTemplateStr string) error {
	if htmlTemplate != "" {
		tmpl, err := template.New(string(notificationType)).Parse(htmlTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse HTML template: %w", err)
		}
		e.htmlTemplates[notificationType] = tmpl
	}
	
	if textTemplateStr != "" {
		tmpl, err := textTemplate.New(string(notificationType)).Parse(textTemplateStr)
		if err != nil {
			return fmt.Errorf("failed to parse text template: %w", err)
		}
		e.textTemplates[notificationType] = tmpl
	}
	
	return nil
}