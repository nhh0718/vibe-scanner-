package ai

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// PromptTemplate data structure for prompts
type PromptData struct {
	ProjectType  string
	Language     string
	Framework    string
	FilePath     string
	IssueType    string
	CodeChunk    string
	Severity     string
	RuleMessage  string
}

// SecurityPrompt template cho security issues
const SecurityPrompt = `Bạn là security expert giải thích cho người không có kiến thức kỹ thuật.

Dự án: {{.ProjectType}} viết bằng {{.Language}}, framework {{.Framework}}
File: {{.FilePath}}
Loại lỗi: {{.IssueType}}
Severity: {{.Severity}}

Code có vấn đề:
` + "```" + `{{.Language}}
{{.CodeChunk}}
` + "```" + `

Trả lời ngắn gọn theo đúng format này:

NGUY_HIỂM: [1-2 câu giải thích tại sao lỗi này nguy hiểm, ví dụ cụ thể hacker có thể làm gì]

SỬA_NHƯ_NÀY:
` + "```" + `{{.Language}}
[code đã sửa, chỉ trả về phần code cần thay đổi]
` + "```" + `

ƯU_TIÊN: [NGAY_BÂY_GIỜ / TUẦN_NÀY / KHI_CÓ_THỜI_GIAN]

LƯU_Ý_THÊM: [mẹo bảo mật bổ sung nếu có]`

// QualityPrompt template cho code quality issues
const QualityPrompt = `Bạn là senior developer giải thích vấn đề chất lượng code.

Dự án: {{.ProjectType}} viết bằng {{.Language}}
File: {{.FilePath}}
Vấn đề: {{.IssueType}}

Code hiện tại:
` + "```" + `{{.Language}}
{{.CodeChunk}}
` + "```" + `

Trả lời theo format:

VẤN_ĐỀ: [giải thích tại sao code này không tốt, ảnh hưởng gì]

CÁCH_SỬA:
` + "```" + `{{.Language}}
[code đã refactor]
` + "```" + `

LỢI_ÍCH: [cải thiện gì sau khi sửa]`

// ArchitecturePrompt template cho architecture issues
const ArchitecturePrompt = `Bạn là software architect tư vấn về kiến trúc code.

Dự án: {{.ProjectType}} ({{.Language}})
File: {{.FilePath}}
Vấn đề kiến trúc: {{.IssueType}}

Context:
` + "```" + `{{.Language}}
{{.CodeChunk}}
` + "```" + `

Trả lời theo format:

VẤN_ĐỀ_KIẾN_TRÚC: [giải thích vấn đề kiến trúc và tác động dài hạn]

GIẢI_PHÁP: [đề xuất refactor hoặc design pattern áp dụng]

` + "```" + `{{.Language}}
[ví dụ code sau khi áp dụng giải pháp]
` + "```"

// SecretPrompt template cho secret detection
const SecretPrompt = `Bạn là security consultant cảnh báo về việc lộ thông tin nhạy cảm.

PHÁT_HIỆN: {{.IssueType}}
File: {{.FilePath}}

` + "```" + `
{{.CodeChunk}}
` + "```" + `

Trả lời theo format:

MỨC_ĐỘ_NGUY_HIỂM: [tại sao việc này cực kỳ nguy hiểm]

HẬU_QUẢ: [điều gì có thể xảy ra nếu bị lộ]

HÀNH_ĐỘNG_KHẨN_CẤP:
1. [bước 1]
2. [bước 2]
3. [bước 3]

CÁCH_BẢO_VỆ_ĐÚNG:
` + "```" + `{{.Language}}
[cách lưu trữ secret đúng]
` + "```" + `

LƯU_Ý: Nếu đã push lên git, cần rotate credential NGAY LẬP TỨC.`

// GetPromptForFinding trả về prompt phù hợp cho finding
func GetPromptForFinding(finding *models.Finding, data PromptData) (string, error) {
	var templateStr string

	switch finding.Category {
	case models.Security:
		templateStr = SecurityPrompt
	case models.Quality:
		templateStr = QualityPrompt
	case models.Architecture:
		templateStr = ArchitecturePrompt
	case models.Secrets:
		templateStr = SecretPrompt
	default:
		templateStr = SecurityPrompt
	}

	tmpl, err := template.New("prompt").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("lỗi parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("lỗi execute template: %w", err)
	}

	return buf.String(), nil
}

// BuildPromptData tạo PromptData từ finding và context
func BuildPromptData(finding *models.Finding, codeChunk string, projectType string) PromptData {
	// Detect language from file extension
	language := detectLanguageFromPath(finding.File)
	framework := detectFramework(codeChunk, language)

	return PromptData{
		ProjectType: projectType,
		Language:    language,
		Framework:   framework,
		FilePath:    finding.File,
		IssueType:   finding.Title,
		CodeChunk:   codeChunk,
		Severity:    string(finding.Severity),
		RuleMessage: finding.Message,
	}
}

// detectLanguageFromPath detect ngôn ngữ từ file path
func detectLanguageFromPath(filePath string) string {
	ext := strings.ToLower(filePath)
	if strings.HasSuffix(ext, ".go") {
		return "go"
	}
	if strings.HasSuffix(ext, ".js") {
		return "javascript"
	}
	if strings.HasSuffix(ext, ".ts") {
		return "typescript"
	}
	if strings.HasSuffix(ext, ".py") {
		return "python"
	}
	if strings.HasSuffix(ext, ".rb") {
		return "ruby"
	}
	if strings.HasSuffix(ext, ".php") {
		return "php"
	}
	if strings.HasSuffix(ext, ".java") {
		return "java"
	}
	if strings.HasSuffix(ext, ".cs") {
		return "csharp"
	}
	return "unknown"
}

// detectFramework detect framework từ code
func detectFramework(code string, language string) string {
	lower := strings.ToLower(code)

	switch language {
	case "javascript", "typescript":
		if strings.Contains(lower, "react") || strings.Contains(lower, "usestate") || strings.Contains(lower, "useeffect") {
			return "React"
		}
		if strings.Contains(lower, "next") {
			return "Next.js"
		}
		if strings.Contains(lower, "express") {
			return "Express"
		}
		if strings.Contains(lower, "vue") {
			return "Vue"
		}
		if strings.Contains(lower, "angular") {
			return "Angular"
		}
	case "python":
		if strings.Contains(lower, "django") {
			return "Django"
		}
		if strings.Contains(lower, "flask") {
			return "Flask"
		}
		if strings.Contains(lower, "fastapi") {
			return "FastAPI"
		}
	case "go":
		if strings.Contains(lower, "gin") {
			return "Gin"
		}
		if strings.Contains(lower, "echo") {
			return "Echo"
		}
		if strings.Contains(lower, "fiber") {
			return "Fiber"
		}
	case "php":
		if strings.Contains(lower, "laravel") {
			return "Laravel"
		}
		if strings.Contains(lower, "symfony") {
			return "Symfony"
		}
	}

	return "unknown"
}
