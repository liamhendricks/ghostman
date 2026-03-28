package parser

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

// codeBlock holds an extracted fenced code block's language and content.
type codeBlock struct {
	lang    string
	content []byte
}

// Parse reads a request markdown file and returns a RequestSpec.
func Parse(path string) (RequestSpec, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return RequestSpec{}, fmt.Errorf("read %s: %w", path, err)
	}

	md := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
	ctx := gmparser.NewContext()
	var buf bytes.Buffer
	if err := md.Convert(src, &buf, gmparser.WithContext(ctx)); err != nil {
		return RequestSpec{}, fmt.Errorf("parse %s: %w", path, err)
	}

	d := frontmatter.Get(ctx)
	if d == nil {
		return RequestSpec{}, fmt.Errorf("%s: no frontmatter found (missing --- delimiters)", path)
	}

	var fm frontmatterRaw
	if err := d.Decode(&fm); err != nil {
		return RequestSpec{}, fmt.Errorf("%s: invalid frontmatter: %w", path, err)
	}

	if fm.GhostmanVersion == 0 {
		return RequestSpec{}, fmt.Errorf("%s: missing required field 'ghostman_version'", path)
	}
	if fm.GhostmanVersion != 1 {
		return RequestSpec{}, fmt.Errorf("%s: unsupported ghostman_version %d (expected 1)", path, fm.GhostmanVersion)
	}

	body, bodyLang, err := extractBody(src)
	if err != nil {
		return RequestSpec{}, fmt.Errorf("%s: %w", path, err)
	}

	return RequestSpec{
		GhostmanVersion: fm.GhostmanVersion,
		Method:          strings.ToUpper(fm.Method),
		BaseURL:         fm.BaseURL,
		Path:            fm.Path,
		Headers:         fm.Headers,
		QueryParams:     fm.QueryParams,
		Body:            body,
		BodyLang:        bodyLang,
		Cert:            fm.Cert,
		Key:             fm.Key,
	}, nil
}

// extractBody parses the markdown source and returns the request body and its language tag.
func extractBody(src []byte) ([]byte, string, error) {
	blocks, err := extractCodeBlocks(src)
	if err != nil {
		return nil, "", err
	}
	return selectBody(blocks)
}

// extractCodeBlocks walks the goldmark AST and collects all fenced code blocks.
func extractCodeBlocks(src []byte) ([]codeBlock, error) {
	md := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
	reader := text.NewReader(src)
	doc := md.Parser().Parse(reader)

	var blocks []codeBlock
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		fcb, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}
		lang := string(fcb.Language(src))
		var content []byte
		lines := fcb.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			content = append(content, line.Value(src)...)
		}
		blocks = append(blocks, codeBlock{lang: lang, content: content})
		return ast.WalkContinue, nil
	})
	return blocks, err
}

// selectBody picks the request body from a list of code blocks.
// Rules:
//  1. If a block with lang "body" exists, it wins immediately (escape hatch).
//  2. Content-type langs: json, xml, form, graphql.
//  3. If two blocks share the same content-type lang, that is ambiguous — error.
//  4. If multiple different content-type langs exist, first wins.
//  5. No candidates = no body (valid for GET etc.).
func selectBody(blocks []codeBlock) ([]byte, string, error) {
	contentTypeLangs := map[string]bool{
		"json":    true,
		"xml":     true,
		"form":    true,
		"graphql": true,
	}

	var candidates []codeBlock
	for _, b := range blocks {
		if b.lang == "body" {
			// Explicit escape hatch — short-circuit immediately.
			return b.content, "body", nil
		}
		if contentTypeLangs[b.lang] {
			candidates = append(candidates, b)
		}
	}

	if len(candidates) == 0 {
		return nil, "", nil
	}

	// Check for ambiguity: same language appearing more than once.
	langCounts := map[string]int{}
	for _, c := range candidates {
		langCounts[c.lang]++
	}
	for lang, count := range langCounts {
		if count > 1 {
			return nil, "", fmt.Errorf(
				"ambiguous body: %d '%s' blocks found — use ```body to disambiguate",
				count, lang,
			)
		}
	}

	// Multiple different content-type langs: first wins.
	return candidates[0].content, candidates[0].lang, nil
}
