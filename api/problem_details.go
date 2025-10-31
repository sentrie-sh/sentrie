package api

import "encoding/json"

func NewProblemDetails(type_, title, detail, instance string, status int, ext map[string]any) *ProblemDetails {
	return &ProblemDetails{
		Type:     type_,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Ext:      ext,
	}
}

// ProblemDetails represents an RFC 9457 Problem Details for HTTP APIs
type ProblemDetails struct {
	Type     string         `json:"type,omitempty"`
	Title    string         `json:"title"`
	Status   int            `json:"status,omitempty"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Ext      map[string]any `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for ProblemDetails
func (p *ProblemDetails) MarshalJSON() ([]byte, error) {
	// Create a map to hold all fields including extensions
	result := make(map[string]any)

	// Add standard fields
	if p.Type != "" {
		result["type"] = p.Type
	}
	if p.Title != "" {
		result["title"] = p.Title
	}
	if p.Status != 0 {
		result["status"] = p.Status
	}
	if p.Detail != "" {
		result["detail"] = p.Detail
	}
	if p.Instance != "" {
		result["instance"] = p.Instance
	}

	// Add extension fields
	for k, v := range p.Ext {
		result[k] = v
	}

	return json.Marshal(result)
}
