package static

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

const ClientCheckSchemaVersion = 1

type ClientDiagnostic struct {
	Source      string `json:"source"`
	Code        string `json:"code"`
	Severity    string `json:"severity"`
	Package     string `json:"package,omitempty"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	Message     string `json:"message"`
	Remediation string `json:"remediation,omitempty"`
}

type ClientCheckResult struct {
	Version     int                `json:"version"`
	OK          bool               `json:"ok"`
	Diagnostics []ClientDiagnostic `json:"diagnostics"`
}

func NewClientCheckResult(diagnostics []ClientDiagnostic) ClientCheckResult {
	diagnostics = append([]ClientDiagnostic{}, diagnostics...)
	sort.SliceStable(diagnostics, func(i, j int) bool {
		a, b := diagnostics[i], diagnostics[j]
		if a.Package != b.Package {
			return a.Package < b.Package
		}
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Column != b.Column {
			return a.Column < b.Column
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
	return ClientCheckResult{
		Version:     ClientCheckSchemaVersion,
		OK:          len(diagnostics) == 0,
		Diagnostics: diagnostics,
	}
}

func (r ClientCheckResult) WriteJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(r)
}

func (r ClientCheckResult) WriteHuman(w io.Writer) error {
	if r.OK {
		_, err := fmt.Fprintln(w, "Client checks passed")
		return err
	}
	for _, diagnostic := range r.Diagnostics {
		location := diagnostic.File
		if diagnostic.Line > 0 {
			location = fmt.Sprintf("%s:%d", location, diagnostic.Line)
			if diagnostic.Column > 0 {
				location = fmt.Sprintf("%s:%d", location, diagnostic.Column)
			}
		}
		if location != "" {
			location += ": "
		}
		if _, err := fmt.Fprintf(w, "%s%s %s: %s\n", location, diagnostic.Severity, diagnostic.Code, diagnostic.Message); err != nil {
			return err
		}
		if diagnostic.Remediation != "" {
			if _, err := fmt.Fprintf(w, "  Fix: %s\n", diagnostic.Remediation); err != nil {
				return err
			}
		}
	}
	return nil
}
