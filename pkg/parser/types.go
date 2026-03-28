package parser

// RequestSpec is the parsed, pre-substitution representation of a request file.
// It is the currency between pkg/parser and pkg/env.
type RequestSpec struct {
	GhostmanVersion int
	Method          string
	BaseURL         string            // may contain {{env:*}} or {{col:*}}
	Path            string            // may contain :param and {{vars}}
	Headers         map[string]string
	QueryParams     map[string]string
	Body            []byte            // extracted from code block, may contain {{vars}}
	BodyLang        string            // "json", "xml", "form", "graphql", "body", or ""
	Cert            string            // path to client TLS certificate, may contain {{vars}}
	Key             string            // path to client TLS private key, may contain {{vars}}
}

// frontmatterRaw is the YAML-decoded frontmatter before validation.
type frontmatterRaw struct {
	GhostmanVersion int               `yaml:"ghostman_version"`
	Method          string            `yaml:"method"`
	BaseURL         string            `yaml:"base_url"`
	Path            string            `yaml:"path"`
	Headers         map[string]string `yaml:"headers"`
	QueryParams     map[string]string `yaml:"query_params"`
	Cert            string            `yaml:"cert"`
	Key             string            `yaml:"key"`
}
