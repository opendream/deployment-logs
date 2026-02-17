package services

import (
	"encoding/json"
	"path"
)

type BranchConfig struct {
	Pattern      string `json:"pattern"`
	Environment  string `json:"environment"`
	DateStrategy string `json:"date_strategy"`
}

func ParseBranchConfigs(raw string) []BranchConfig {
	if raw == "" {
		return nil
	}
	var configs []BranchConfig
	if err := json.Unmarshal([]byte(raw), &configs); err != nil {
		return nil
	}
	return configs
}

// MatchBranch finds the BranchConfig whose pattern matches the given branch name.
// Supports glob patterns via path.Match (e.g. "release/*" matches "release/1.1.11").
func MatchBranch(configs []BranchConfig, branchName string) *BranchConfig {
	for i, cfg := range configs {
		if cfg.Pattern == branchName {
			return &configs[i]
		}
		if matched, _ := path.Match(cfg.Pattern, branchName); matched {
			return &configs[i]
		}
	}
	return nil
}
