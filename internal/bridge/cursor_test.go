package bridge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/aisync/internal/bridge"
)

const longDescriptionContent = `This is a test rule with description This is a test rule with description This is a test rule with description This is a test rule with description This is a test rule with description`

func TestCursorRule_String(t *testing.T) {
	testCases := []struct {
		name     string
		rule     bridge.CursorRule
		expected string
	}{
		{
			name: "TS Snapshot - AlwaysApply true",
			rule: bridge.CursorRule{
				Content: "# Always Apply Rule\n\nThis rule is always applied.",
				Metadata: bridge.CursorRuleMetadata{
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
			rule: bridge.CursorRule{
				Content: "# Glob Rule\n\nThis rule applies to specific file patterns.",
				Metadata: bridge.CursorRuleMetadata{
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
			rule: bridge.CursorRule{
				Content: "# Description Rule\n\nThis rule has a description.",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: false,
					Description: longDescriptionContent,
					Globs:       "", // Empty, expect 'globs:'
				},
			},
			expected: `---
alwaysApply: false
description: This is a test rule with description This is a test rule with description This is a test rule with description This is a test rule with description This is a test rule with description
globs:
---
# Description Rule

This rule has a description.
`,
		},
		{
			name: "Content normalization - no trailing newline",
			rule: bridge.CursorRule{
				Content: "# Test Rule No Newline", // Normalized to end with one \n
				Metadata: bridge.CursorRuleMetadata{
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
			rule: bridge.CursorRule{
				Content: "# Test Rule One Newline\n", // Normalized to end with one \n
				Metadata: bridge.CursorRuleMetadata{
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
			rule: bridge.CursorRule{
				Content: "# Test Rule Multiple Newlines\n\n\n", // Normalized to end with one \n
				Metadata: bridge.CursorRuleMetadata{
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
			rule: bridge.CursorRule{
				Content: "", // Normalized to end with one \n (becomes just "\n")
				Metadata: bridge.CursorRuleMetadata{
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
			rule: bridge.CursorRule{
				Slug:    "my-test-rule",
				Content: "Rule content.",
				Metadata: bridge.CursorRuleMetadata{
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
