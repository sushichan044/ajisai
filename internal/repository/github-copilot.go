package repository

import (
	"github.com/sushichan044/ajisai/internal/bridge"
	"github.com/sushichan044/ajisai/internal/domain"
)

type gitHubCopilotAdapter struct {
	bridge domain.AgentBridge[bridge.GitHubCopilotInstruction, bridge.GitHubCopilotPrompt]
}

const (
	gitHubCopilotInstructionExtension = ".instructions.md"
	gitHubCopilotPromptExtension      = ".prompt.md"

	githubCopilotInstructionsDir = ".github/instructions"
	githubCopilotPromptsDir      = ".github/prompts"
)

func NewGitHubCopilotAdapter() AgentFileAdapter {
	return &gitHubCopilotAdapter{
		bridge: bridge.NewGitHubCopilotBridge(),
	}
}

func (adapter *gitHubCopilotAdapter) RuleExtension() string {
	return gitHubCopilotInstructionExtension
}

func (adapter *gitHubCopilotAdapter) PromptExtension() string {
	return gitHubCopilotPromptExtension
}

func (adapter *gitHubCopilotAdapter) RulesDir() string {
	return githubCopilotInstructionsDir
}

func (adapter *gitHubCopilotAdapter) PromptsDir() string {
	return githubCopilotPromptsDir
}

func (adapter *gitHubCopilotAdapter) SerializeRule(rule *domain.RuleItem) (string, error) {
	agentRule, err := adapter.bridge.ToAgentRule(*rule)
	if err != nil {
		return "", err
	}

	return adapter.bridge.SerializeAgentRule(agentRule)
}

func (adapter *gitHubCopilotAdapter) SerializePrompt(prompt *domain.PromptItem) (string, error) {
	agentPrompt, err := adapter.bridge.ToAgentPrompt(*prompt)
	if err != nil {
		return "", err
	}

	return adapter.bridge.SerializeAgentPrompt(agentPrompt)
}
