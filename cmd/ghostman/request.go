package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/liamhendricks/ghostman/pkg/collection"
	"github.com/liamhendricks/ghostman/pkg/config"
	"github.com/liamhendricks/ghostman/pkg/env"
	"github.com/liamhendricks/ghostman/pkg/executor"
	"github.com/liamhendricks/ghostman/pkg/parser"
	"github.com/spf13/cobra"
)

var pathParamPattern = regexp.MustCompile(`/:([a-zA-Z][a-zA-Z0-9_]*)`)

func newRequestCmd() *cobra.Command {
	var envName string
	var dryRun bool
	var insecure bool

	cmd := &cobra.Command{
		Use:   "request <collection/name>",
		Short: "Execute a named API request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRequest(cmd, args[0], envName, dryRun, insecure)
		},
	}
	cmd.Flags().StringVarP(&envName, "env", "e", "", "environment name (e.g. staging)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print substituted request without sending")
	cmd.Flags().BoolVar(&insecure, "insecure", false, "skip TLS certificate verification")
	return cmd
}

func runRequest(cmd *cobra.Command, collectionName string, envName string, dryRun bool, insecure bool) error {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 2. Build collection roots
	roots, err := collection.Roots(cfg)
	if err != nil {
		return err
	}

	// 3. Find the request file
	result, err := collection.Find(collectionName, roots)
	if err != nil {
		return err
	}

	// 4. Parse the request markdown
	spec, err := parser.Parse(result.FilePath)
	if err != nil {
		return err
	}

	// 5. Determine active env name
	activeEnv := envName
	if activeEnv == "" {
		activeEnv = cfg.DefaultEnv
	}

	// 6. Check env requirement
	if containsEnvRef(spec) && activeEnv == "" {
		return fmt.Errorf("--env is required: request contains {{env:*}} references but no environment is active\n  Set --env <name> or configure default_env in ~/.ghostmanrc")
	}

	// 7 & 8. Load environment and col vars
	var environment env.Env
	colVars, err := collection.LoadColVars(result.CollectionDir)
	if err != nil {
		return err
	}

	if activeEnv != "" {
		envFilePath := filepath.Join(result.CollectionRoot, "env", activeEnv+".yaml")
		environment, err = env.Load(envFilePath)
		if err != nil {
			return err
		}
		environment = environment.WithColVars(colVars)
	} else {
		// 9. No env refs and no active env — create empty env with col vars only
		environment = env.Env{EnvVars: map[string]string{}, ColVars: colVars}
	}

	// 10. Substitute all fields
	vars := env.Vars{Env: environment.EnvVars, Col: environment.ColVars}

	spec.BaseURL, err = env.Substitute(spec.BaseURL, vars)
	if err != nil {
		return fmt.Errorf("base_url: %w", err)
	}

	spec.Path, err = env.Substitute(spec.Path, vars)
	if err != nil {
		return fmt.Errorf("path: %w", err)
	}

	for k, v := range spec.Headers {
		substituted, err := env.Substitute(v, vars)
		if err != nil {
			return fmt.Errorf("header %s: %w", k, err)
		}
		spec.Headers[k] = substituted
	}

	for k, v := range spec.QueryParams {
		substituted, err := env.Substitute(v, vars)
		if err != nil {
			return fmt.Errorf("query param %s: %w", k, err)
		}
		spec.QueryParams[k] = substituted
	}

	bodyStr, err := env.Substitute(string(spec.Body), vars)
	if err != nil {
		return fmt.Errorf("body: %w", err)
	}
	spec.Body = []byte(bodyStr)

	spec.Cert, err = env.Substitute(spec.Cert, vars)
	if err != nil {
		return fmt.Errorf("cert: %w", err)
	}
	spec.Key, err = env.Substitute(spec.Key, vars)
	if err != nil {
		return fmt.Errorf("key: %w", err)
	}

	// 11. Resolve :param in Path
	var pathParamErr error
	spec.Path = pathParamPattern.ReplaceAllStringFunc(spec.Path, func(match string) string {
		if pathParamErr != nil {
			return match
		}
		key := match[2:] // strip "/:"
		if v, ok := colVars[key]; ok {
			return "/" + v
		}
		if v, ok := environment.EnvVars[key]; ok {
			return "/" + v
		}
		pathParamErr = fmt.Errorf("undefined path parameter :%s", key)
		return match
	})
	if pathParamErr != nil {
		return pathParamErr
	}

	// 12. Dry run
	if dryRun {
		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "DRY RUN — request not sent")
		fmt.Fprintf(out, "METHOD  %s\n", spec.Method)
		fmt.Fprintf(out, "URL     %s\n", spec.BaseURL+spec.Path)
		if len(spec.Headers) > 0 {
			first := true
			for k, v := range spec.Headers {
				if first {
					fmt.Fprintf(out, "HEADERS %s: %s\n", k, v)
					first = false
				} else {
					fmt.Fprintf(out, "        %s: %s\n", k, v)
				}
			}
		}
		if len(spec.Body) > 0 {
			fmt.Fprintln(out, "BODY")
			fmt.Fprintf(out, "%s\n", string(spec.Body))
		}
		return nil
	}

	// 13. Build HTTP client
	var clientCert *tls.Certificate
	if spec.Cert != "" || spec.Key != "" {
		if spec.Cert == "" || spec.Key == "" {
			return fmt.Errorf("cert and key must both be set")
		}
		certPath, err := expandHome(spec.Cert)
		if err != nil {
			return fmt.Errorf("cert: %w", err)
		}
		keyPath, err := expandHome(spec.Key)
		if err != nil {
			return fmt.Errorf("key: %w", err)
		}
		pair, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return fmt.Errorf("loading client certificate: %w", err)
		}
		clientCert = &pair
	}
	client := buildClient(insecure, clientCert)

	// 14. Execute
	resp, execErr := executor.New(client).Execute(cmd.Context(), spec)

	// 15. Write body to stdout (always, even on 4xx/5xx)
	if len(resp.Body) > 0 {
		cmd.OutOrStdout().Write(resp.Body) //nolint:errcheck
	}

	// 16. Write metadata to stderr
	if resp.StatusCode != 0 {
		// Strip the numeric prefix from Status since we print StatusCode separately
		statusText := strings.TrimPrefix(resp.Status, fmt.Sprintf("%d ", resp.StatusCode))
		fmt.Fprintf(cmd.ErrOrStderr(), "%d %s | %s | %d bytes\n",
			resp.StatusCode,
			statusText,
			resp.Duration.Round(time.Millisecond),
			resp.BodySize,
		)
	}

	// 17 & 18. Return error if any
	return execErr
}

// containsEnvRef returns true if any field in the spec contains a {{env:*}} placeholder.
func containsEnvRef(spec parser.RequestSpec) bool {
	if strings.Contains(spec.BaseURL, "{{env:") {
		return true
	}
	if strings.Contains(spec.Path, "{{env:") {
		return true
	}
	for _, v := range spec.Headers {
		if strings.Contains(v, "{{env:") {
			return true
		}
	}
	for _, v := range spec.QueryParams {
		if strings.Contains(v, "{{env:") {
			return true
		}
	}
	if strings.Contains(string(spec.Body), "{{env:") {
		return true
	}
	return false
}

// buildClient creates an HTTP client with TLS configuration.
func buildClient(insecure bool, clientCert *tls.Certificate) *http.Client {
	tlsCfg := &tls.Config{InsecureSkipVerify: insecure} //nolint:gosec
	if clientCert != nil {
		tlsCfg.Certificates = []tls.Certificate{*clientCert}
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}
}

// expandHome expands a leading ~/ to the user's home directory.
func expandHome(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}
