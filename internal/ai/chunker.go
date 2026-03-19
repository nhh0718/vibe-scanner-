package ai

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Chunker trích xuất code context cho AI
type Chunker struct {
	maxLines    int
	contextLines int
}

// NewChunker tạo chunker mới
func NewChunker() *Chunker {
	return &Chunker{
		maxLines:     50,
		contextLines: 5,
	}
}

// ExtractFunction trích xuất function chứa dòng lỗi
func (c *Chunker) ExtractFunction(filePath string, lineNum int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if lineNum > len(lines) {
		return "", fmt.Errorf("line number exceeds file length")
	}

	// Find function boundaries
	startLine := c.findFunctionStart(lines, lineNum)
	endLine := c.findFunctionEnd(lines, lineNum)

	// Limit chunk size
	if endLine-startLine > c.maxLines {
		// Take context around the error line
		startLine = max(0, lineNum-c.maxLines/2)
		endLine = min(len(lines), lineNum+c.maxLines/2)
	}

	// Build chunk
	var chunk strings.Builder
	for i := startLine; i < endLine && i < len(lines); i++ {
		lineNum := i + 1
		prefix := "  "
		if i+1 == lineNum {
			prefix = "> " // Mark error line
		}
		chunk.WriteString(fmt.Sprintf("%s%3d: %s\n", prefix, lineNum, lines[i]))
	}

	return chunk.String(), nil
}

// ExtractContext trích xuất context xung quanh dòng lỗi
func (c *Chunker) ExtractContext(filePath string, lineNum int, contextLines int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if lineNum > len(lines) {
		return "", fmt.Errorf("line number exceeds file length")
	}

	startLine := max(0, lineNum-contextLines-1)
	endLine := min(len(lines), lineNum+contextLines)

	var chunk strings.Builder
	for i := startLine; i < endLine; i++ {
		lineIndicator := "  "
		if i == lineNum-1 {
			lineIndicator = "> " // Mark error line
		}
		chunk.WriteString(fmt.Sprintf("%s%3d: %s\n", lineIndicator, i+1, lines[i]))
	}

	return chunk.String(), nil
}

// findFunctionStart tìm dòng bắt đầu function
func (c *Chunker) findFunctionStart(lines []string, errorLine int) int {
	// Simple heuristic: look for function declarations
	// This is a basic implementation - could be improved with AST parsing
	for i := errorLine - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		// Look for function patterns
		if strings.Contains(line, "func ") ||
			strings.Contains(line, "function ") ||
			strings.Contains(line, "def ") ||
			strings.Contains(line, "async ") ||
			strings.Contains(line, "const ") && strings.Contains(line, "= ") ||
			strings.Contains(line, "var ") && strings.Contains(line, "= ") ||
			strings.Contains(line, "let ") && strings.Contains(line, "= ") {
			return i
		}
		// Stop at empty lines or comments that might indicate new section
		if i < errorLine-10 && line == "" {
			return i + 1
		}
	}
	return max(0, errorLine-20)
}

// findFunctionEnd tìm dòng kết thúc function
func (c *Chunker) findFunctionEnd(lines []string, errorLine int) int {
	braceCount := 0
	inFunction := false

	for i := errorLine - 1; i < len(lines); i++ {
		line := lines[i]
		for _, char := range line {
			if char == '{' || char == '(' {
				braceCount++
				inFunction = true
			} else if char == '}' || char == ')' {
				braceCount--
			}
		}

		if inFunction && braceCount == 0 {
			return i + 1
		}

		// Safety limit
		if i > errorLine+c.maxLines {
			return errorLine + c.maxLines
		}
	}

	return len(lines)
}

// EstimateTokens ước tính số tokens
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token for code
	return len(text) / 4
}

// CountLinesInFile đếm số dòng trong file
func CountLinesInFile(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}

	return count, scanner.Err()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
