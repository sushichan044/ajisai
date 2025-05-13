package repository

import (
	"github.com/sushichan044/ajisai/internal/bridge"
	"github.com/sushichan044/ajisai/internal/domain"
)

type windsurfAdapter struct {
	bridge domain.AgentBridge[bridge.WindsurfRule, bridge.WindsurfPrompt]
}

const (
	windsurfRuleExtension   = ".md"
	windsurfPromptExtension = ".md"

	windsurfRulesDir   = ".windsurf/rules"
	windsurfPromptsDir = ".windsurf/prompts"
)

func NewWindsurfAdapter() AgentFileAdapter {
	return &windsurfAdapter{
		bridge: bridge.NewWindsurfBridge(),
	}
}

func (adapter *windsurfAdapter) RuleExtension() string {
	return windsurfRuleExtension
}

func (adapter *windsurfAdapter) PromptExtension() string {
	return windsurfPromptExtension
}

func (adapter *windsurfAdapter) RulesDir() string {
	return windsurfRulesDir
}

func (adapter *windsurfAdapter) PromptsDir() string {
	return windsurfPromptsDir
}

func (adapter *windsurfAdapter) SerializeRule(rule *domain.RuleItem) (string, error) {
	agentRule, err := adapter.bridge.ToAgentRule(*rule)
	if err != nil {
		return "", err
	}

	return adapter.bridge.SerializeAgentRule(agentRule)
}

func (adapter *windsurfAdapter) SerializePrompt(prompt *domain.PromptItem) (string, error) {
	agentPrompt, err := adapter.bridge.ToAgentPrompt(*prompt)
	if err != nil {
		return "", err
	}

	return adapter.bridge.SerializeAgentPrompt(agentPrompt)
}
