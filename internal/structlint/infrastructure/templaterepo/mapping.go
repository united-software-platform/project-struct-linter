package templaterepo

import "struct-linter/internal/structlint/domain"

// templateFile — представление файла эталона в YAML; поля snake_case соответствуют схеме.
type templateFile struct {
	SchemaVersion string           `yaml:"schema_version"`
	ID            string           `yaml:"id"`
	Version       string           `yaml:"version"`
	Language      string           `yaml:"language"`
	Detect        []string         `yaml:"detect"`
	Required      []pathRuleFile   `yaml:"required"`
	Forbidden     []pathRuleFile   `yaml:"forbidden"`
	Naming        []namingRuleFile `yaml:"naming"`
	Layers        []layerRuleFile  `yaml:"layers"`
}

type pathRuleFile struct {
	Name    string `yaml:"name"`
	Kind    string `yaml:"kind"`
	Match   string `yaml:"match"`
	Pattern string `yaml:"pattern"`
	Min     int    `yaml:"min"`
	Reason  string `yaml:"reason"`
}

type namingRuleFile struct {
	Name   string `yaml:"name"`
	Scope  string `yaml:"scope"`
	Kind   string `yaml:"kind"`
	Regex  string `yaml:"regex"`
	Reason string `yaml:"reason"`
}

type layerRuleFile struct {
	Name         string   `yaml:"name"`
	Selector     string   `yaml:"selector"`
	AllowedPaths []string `yaml:"allowed_paths"`
	Reason       string   `yaml:"reason"`
}

// toDomain переводит разобранный файл эталона в доменный тип.
func (f templateFile) toDomain() domain.Template {
	tmpl := domain.Template{
		SchemaVersion: f.SchemaVersion,
		ID:            f.ID,
		Version:       f.Version,
		Language:      f.Language,
		Detect:        f.Detect,
	}
	for _, r := range f.Required {
		tmpl.Required = append(tmpl.Required, r.toDomain())
	}
	for _, r := range f.Forbidden {
		tmpl.Forbidden = append(tmpl.Forbidden, r.toDomain())
	}
	for _, r := range f.Naming {
		tmpl.Naming = append(tmpl.Naming, domain.NamingRule{
			Name:   r.Name,
			Scope:  r.Scope,
			Kind:   domain.Kind(r.Kind),
			Regex:  r.Regex,
			Reason: r.Reason,
		})
	}
	for _, r := range f.Layers {
		tmpl.Layers = append(tmpl.Layers, domain.LayerRule{
			Name:         r.Name,
			Selector:     r.Selector,
			AllowedPaths: r.AllowedPaths,
			Reason:       r.Reason,
		})
	}
	return tmpl
}

// toDomain переводит правило пути из файла в доменный тип.
func (r pathRuleFile) toDomain() domain.PathRule {
	return domain.PathRule{
		Name:    r.Name,
		Kind:    domain.Kind(r.Kind),
		Match:   domain.Match(r.Match),
		Pattern: r.Pattern,
		Min:     r.Min,
		Reason:  r.Reason,
	}
}
