package adapter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ai-rules-manager/internal/adapter"
)

const longDescriptionContent = `This is a test rule with description This is a test rule with description This is a test rule with description This is a test rule with description This is a test rule with description `

func TestCursorRule_String(t *testing.T) {
	testCases := []struct {
		name     string
		rule     adapter.CursorRule
		expected string
	}{
		{
			name: "TS Snapshot - AlwaysApply true",
			rule: adapter.CursorRule{
				Content: "# Always Apply Rule\n\nThis rule is always applied.",
				Metadata: adapter.CursorRuleMetadata{
					AlwaysApply: true,
					Description: "", // Empty, expect 'description:'
					Globs:       "", // Empty, expect 'globs:'
				},
			},
			expected: `---
alwaysApply: true
description:
globs:
---
# Always Apply Rule

This rule is always applied.
`,
		},
		{
			name: "TS Snapshot - GlobRule (no quotes, empty description)",
			rule: adapter.CursorRule{
				Content: "# Glob Rule\n\nThis rule applies to specific file patterns.",
				Metadata: adapter.CursorRuleMetadata{
					AlwaysApply: false,
					Description: "", // Empty, expect 'description:'
					Globs:       "*.ts,src/**/*.{ts,tsx}",
				},
			},
			expected: `---
alwaysApply: false
description:
globs: *.ts,src/**/*.{ts,tsx}
---
# Glob Rule

This rule applies to specific file patterns.
`,
		},
		{
			name: "TS Snapshot - LongDescription (single quotes, trailing space, empty globs)",
			rule: adapter.CursorRule{
				Content: "# Description Rule\n\nThis rule has a description.",
				Metadata: adapter.CursorRuleMetadata{
					AlwaysApply: false,
					Description: longDescriptionContent,
					Globs:       "", // Empty, expect 'globs:'
				},
			},
			expected: `---
alwaysApply: false
description: 'This is a test rule with description This is a test rule with description This is a test rule with description This is a test rule with description This is a test rule with description '
globs:
---
# Description Rule

This rule has a description.
`,
		},
		{
			name: "Content normalization - no trailing newline",
			rule: adapter.CursorRule{
				Content: "# Test Rule No Newline", // Normalized to end with one \n
				Metadata: adapter.CursorRuleMetadata{
					AlwaysApply: true, Description: "", Globs: "",
				},
			},
			expected: `---
alwaysApply: true
description:
globs:
---
# Test Rule No Newline
`,
		},
		{
			name: "Content normalization - one trailing newline",
			rule: adapter.CursorRule{
				Content: "# Test Rule One Newline\n", // Normalized to end with one \n
				Metadata: adapter.CursorRuleMetadata{
					AlwaysApply: true, Description: "", Globs: "",
				},
			},
			expected: `---
alwaysApply: true
description:
globs:
---
# Test Rule One Newline
`,
		},
		{
			name: "Content normalization - multiple trailing newlines",
			rule: adapter.CursorRule{
				Content: "# Test Rule Multiple Newlines\n\n\n", // Normalized to end with one \n
				Metadata: adapter.CursorRuleMetadata{
					AlwaysApply: true, Description: "", Globs: "",
				},
			},
			expected: `---
alwaysApply: true
description:
globs:
---
# Test Rule Multiple Newlines
`,
		},
		{
			name: "Content normalization - empty content",
			rule: adapter.CursorRule{
				Content: "", // Normalized to end with one \n (becomes just "\n")
				Metadata: adapter.CursorRuleMetadata{
					AlwaysApply: true, Description: "", Globs: "",
				},
			},
			expected: `---
alwaysApply: true
description:
globs:
---
`,
		},
		{
			name: "Slug is not part of the string output (uses alwaysApply case for structure)",
			rule: adapter.CursorRule{
				Slug:    "my-test-rule",
				Content: "Rule content.",
				Metadata: adapter.CursorRuleMetadata{
					AlwaysApply: true, Description: "", Globs: "",
				},
			},
			expected: `---
alwaysApply: true
description:
globs:
---
Rule content.
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.rule.String()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
