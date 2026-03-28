package env

// Vars holds the resolved variable maps for each namespace.
type Vars struct {
	Env map[string]string // {{env:*}} variables
	Col map[string]string // {{col:*}} variables
}

// Env represents a loaded environment.
type Env struct {
	EnvVars map[string]string // {{env:*}} — from env YAML file
	ColVars map[string]string // {{col:*}} — from vars.yaml
}

// Load reads a YAML env file and returns an Env.
func Load(envFilePath string) (Env, error) {
	vars, err := LoadEnvFile(envFilePath)
	if err != nil {
		return Env{}, err
	}
	return Env{EnvVars: vars}, nil
}

// WithColVars returns a new Env with the same EnvVars but with the provided ColVars set.
func (e Env) WithColVars(colVars map[string]string) Env {
	return Env{EnvVars: e.EnvVars, ColVars: colVars}
}
