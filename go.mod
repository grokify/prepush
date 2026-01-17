module github.com/agentplexus/release-agent-team

go 1.23

require (
	github.com/agentplexus/aiassistkit v0.0.0
	github.com/spf13/cobra v1.10.2
	github.com/toon-format/toon-go v0.0.0-20251202084852-7ca0e27c4e8c
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/agentplexus/multi-agent-spec/sdk/go v0.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

// Local development - remove before publishing
replace github.com/agentplexus/aiassistkit => ../aiassistkit

replace github.com/agentplexus/multi-agent-spec/sdk/go => ../multi-agent-spec/sdk/go
