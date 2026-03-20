package rules

import (
	"github.com/odvcencio/gotreesitter"
)

// Severity level
type Severity int

const (
	Info Severity = iota
	Low
	Medium
	High
	Critical
)

// Category type
type Category string

const (
	Security     Category = "security"
	Quality      Category = "quality"
	Architecture Category = "architecture"
	Performance  Category = "performance"
	Secrets      Category = "secrets"
)

func (s Severity) String() string {
	switch s {
	case Critical:
		return "critical"
	case High:
		return "high"
	case Medium:
		return "medium"
	case Low:
		return "low"
	default:
		return "info"
	}
}

// Finding represents a security/quality issue found
// (compatible với models.Finding để dễ chuyển đổi)
type Finding struct {
	RuleID      string
	Title       string
	Description string
	Fix         string
	File        string
	Line        int
	Col         int
	Snippet     string
	Severity    Severity
	Category    string
	Tags        []string
}

// Rule interface cho tất cả rules
type Rule interface {
	ID() string
	Title() string
	Languages() []string
	Check(file *ParsedFile) []Finding
}

// ParsedFile chứa mọi thứ engine cần
type ParsedFile struct {
	Path     string
	Content  []byte
	Language string
	Tree     *gotreesitter.Tree
}

// AllRules trả về tất cả rules được đăng ký
func AllRules() []Rule {
	return []Rule{
		// Security Critical (AST-based)
		&SQLInjectionRule{},
		&CommandInjectionRule{},
		&HardcodedSecretRule{},
		&WeakJWTSecretRule{},
		&JWTNoVerifyRule{},
		&PlainPasswordRule{},
		&WeakBcryptRule{},
		&EvalUserInputRule{},
		&PathTraversalRule{},
		&EnvInGitignoreRule{},

		// Security High
		&CORSWildcardRule{},
		&DangerousHTMLRule{},
		&XSSInnerHTMLRule{},
		&HTTPNotHTTPSRule{},
		&MathRandomSecurityRule{},
		&ExposeStackTraceRule{},

		// Security Advanced
		&RateLimitingRule{},
		&CSRFRule{},
		&AuthMiddlewareRule{},
		&PickleDeserializationRule{},
		&DebugModeRule{},
		&ConsoleLogSensitiveRule{},
		&EnvCheckRule{},
		&InputValidationRule{},

		// Performance
		&NPlusOneQueryRule{},
		&ReadFileSyncInAsyncRule{},
		&LodashFullImportRule{},
		&JSONParseInLoopRule{},

		// Architecture
		&CircularImportRule{},
		&BusinessLogicInRouteRule{},
		&DirectDBCallInControllerRule{},
		&DynamicRequireRule{},

		// Vibe-specific
		&EnvLoadOrderRule{},
		&DuplicateMiddlewareRule{},
		&SessionSecretRule{},
		&MongooseValidateRule{},
		&ResponseStatusCodeRule{},
		&SelectStarRule{},
		&APIKeyFrontendRule{},
		&HardcodePortRule{},

		// Quality
		&ComplexityRule{Threshold: 10},
		&LongFunctionRule{},
		&LongFileRule{},
		&DeepNestingRule{},
		&TooManyParamsRule{},
		&EmptyCatchRule{},
		&ConsoleLogRule{},
		&UnusedVarRule{},
		&DeadCodeRule{},
		&MagicNumberRule{},
		&TodoCommentRule{},
		&VarInsteadOfConstRule{},
	}
}

// ruleApplies kiểm tra rule có áp dụng cho ngôn ngữ không
func RuleApplies(rule Rule, lang string) bool {
	langs := rule.Languages()
	if len(langs) == 1 && langs[0] == "*" {
		return true
	}
	for _, l := range langs {
		if l == lang {
			return true
		}
	}
	return false
}
