package integration

import (
	"github.com/sushichan044/ajisai/internal/bridge"
	"github.com/sushichan044/ajisai/internal/domain"
)

type cursorAdapter struct {
	bridge domain.AgentBridge[bridge.CursorRule, bridge.CursorPrompt]
}

const (
	cursorRuleExtension   = ".mdc"
	cursorPromptExtension = ".md"

	cursorRulesDir   = ".cursor/rules"
	cursorPromptsDir = ".cursor/prompts"
)

func NewCursorAdapter() agentSpecificationAdapter {
	return &cursorAdapter{
		bridge: bridge.NewCursorBridge(),
	}
}

func (adapter *cursorAdapter) RuleExtension() string {
	return cursorRuleExtension
}

func (adapter *cursorAdapter) PromptExtension() string {
	return cursorPromptExtension
}

func (adapter *cursorAdapter) RulesDir() string {
	return cursorRulesDir
}

func (adapter *cursorAdapter) PromptsDir() string {
	return cursorPromptsDir
}

func (adapter *cursorAdapter) SerializeRule(rule *domain.RuleItem) (string, error) {
	cursorRule, ruleConversionErr := adapter.bridge.ToAgentRule(*rule)
	if ruleConversionErr != nil {
		return "", ruleConversionErr
	}

	return adapter.bridge.SerializeAgentRule(cursorRule)
}

func (adapter *cursorAdapter) SerializePrompt(prompt *domain.PromptItem) (string, error) {
	cursorPrompt, promptConversionErr := adapter.bridge.ToAgentPrompt(*prompt)
	if promptConversionErr != nil {
		return "", promptConversionErr
	}

	return adapter.bridge.SerializeAgentPrompt(cursorPrompt)
}
