package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIService AI服务接口
type AIService struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	mode       string // cloud, local, hybrid
}

// NewAIService 创建AI服务
func NewAIService(mode, apiKey, model, baseURL string) *AIService {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4"
	}

	return &AIService{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		mode:    mode,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // AI响应可能较慢
		},
	}
}

// GenerateCaseRequest 用例生成请求
type GenerateCaseRequest struct {
	Requirement      string            `json:"requirement"`
	PriorityStrategy string            `json:"priority_strategy"`
	CaseStyle        string            `json:"case_style"`
	Language         string            `json:"language"`
	MaxCases         int               `json:"max_cases"`
	CustomPrompt     string            `json:"custom_prompt"`
	Context          map[string]string `json:"context"`
}

// GenerateCaseResponse 用例生成响应
type GenerateCaseResponse struct {
	Cases     []GeneratedCase `json:"cases"`
	Summary   string          `json:"summary"`
	UsedModel string          `json:"used_model"`
}

// GeneratedCase 生成的测试用例
type GeneratedCase struct {
	Title        string       `json:"title"`
	TitleEn      string       `json:"title_en"`
	Priority     string       `json:"priority"`
	Precondition string       `json:"precondition"`
	Steps        []CaseStep   `json:"steps"`
	Tags         []string     `json:"tags"`
}

// CaseStep 测试步骤
type CaseStep struct {
	StepNum  int    `json:"step_num"`
	Action   string `json:"action"`
	ActionEn string `json:"action_en"`
	Expected string `json:"expected"`
}

// GenerateCases 生成测试用例
func (s *AIService) GenerateCases(req *GenerateCaseRequest) (*GenerateCaseResponse, error) {
	// 构建提示词
	systemPrompt := s.buildSystemPrompt(req)
	userPrompt := s.buildUserPrompt(req)

	// 调用AI API
	response, err := s.callAI(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// 解析响应
	cases, err := s.parseCasesFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 限制用例数量
	if len(cases) > req.MaxCases && req.MaxCases > 0 {
		cases = cases[:req.MaxCases]
	}

	return &GenerateCaseResponse{
		Cases:     cases,
		Summary:   fmt.Sprintf("成功生成%d个测试用例", len(cases)),
		UsedModel: s.model,
	}, nil
}

// OptimizeCaseRequest 用例优化请求
type OptimizeCaseRequest struct {
	Case         GeneratedCase `json:"case"`
	OptimizeType string        `json:"optimize_type"` // enhance, simplify, translate, format
	TargetLang   string        `json:"target_lang"`
	CustomPrompt string        `json:"custom_prompt"`
}

// OptimizeCaseResponse 用例优化响应
type OptimizeCaseResponse struct {
	Case         GeneratedCase `json:"case"`
	Changes      []string      `json:"changes"`
	UsedModel    string        `json:"used_model"`
}

// OptimizeCase 优化测试用例
func (s *AIService) OptimizeCase(req *OptimizeCaseRequest) (*OptimizeCaseResponse, error) {
	systemPrompt := s.buildOptimizeSystemPrompt(req)
	userPrompt := s.buildOptimizeUserPrompt(req)

	response, err := s.callAI(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	optimizedCase, err := s.parseCaseFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &OptimizeCaseResponse{
		Case:      *optimizedCase,
		Changes:   []string{"用例已优化"},
		UsedModel: s.model,
	}, nil
}

// AnalyzeDefect 分析缺陷
func (s *AIService) AnalyzeDefect(title, description string) (map[string]interface{}, error) {
	systemPrompt := `你是一个专业的软件测试专家。请分析给定的缺陷，提供以下信息：
1. 缺陷分类（功能缺陷/性能问题/UI问题/安全问题/其他）
2. 严重程度建议（fatal/critical/general/minor）
3. 可能的根本原因
4. 修复建议

请以JSON格式返回结果。`

	userPrompt := fmt.Sprintf("缺陷标题：%s\n缺陷描述：%s", title, description)

	response, err := s.callAI(systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 如果解析失败，返回原始响应
		result = map[string]interface{}{
			"analysis": response,
		}
	}

	return result, nil
}

// GenerateReportSummary 生成报告摘要
func (s *AIService) GenerateReportSummary(stats map[string]interface{}) (string, error) {
	systemPrompt := `你是一个专业的软件测试报告撰写专家。根据提供的测试数据，生成一份专业、简洁的测试报告摘要。
摘要应包括：
1. 测试概况
2. 关键发现
3. 风险提示
4. 建议
请用简洁、专业的语言撰写。`

	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	userPrompt := fmt.Sprintf("测试数据：\n%s", string(statsJSON))

	return s.callAI(systemPrompt, userPrompt)
}

// ============================================
// 内部方法
// ============================================

func (s *AIService) buildSystemPrompt(req *GenerateCaseRequest) string {
	langPrompt := "使用中文"
	if req.Language == "en-US" {
		langPrompt = "Use English"
	}

	return fmt.Sprintf(`你是一个专业的软件测试工程师，擅长编写高质量的测试用例。

要求：
1. %s
2. 每个用例包含：标题、优先级、前置条件、测试步骤（步骤编号、操作、预期结果）
3. 优先级：%s
4. 用例风格：%s
5. 返回JSON格式，包含cases数组

输出格式示例：
{
  "cases": [
    {
      "title": "用例标题",
      "title_en": "Case Title",
      "priority": "P1",
      "precondition": "前置条件",
      "steps": [
        {"step_num": 1, "action": "操作步骤", "action_en": "Action", "expected": "预期结果"}
      ],
      "tags": ["标签1", "标签2"]
    }
  ]
}`, langPrompt, req.PriorityStrategy, req.CaseStyle)
}

func (s *AIService) buildUserPrompt(req *GenerateCaseRequest) string {
	prompt := fmt.Sprintf("需求内容：\n%s\n", req.Requirement)

	if req.CustomPrompt != "" {
		prompt += fmt.Sprintf("\n额外要求：\n%s\n", req.CustomPrompt)
	}

	if req.MaxCases > 0 {
		prompt += fmt.Sprintf("\n最多生成%d个用例。", req.MaxCases)
	}

	return prompt
}

func (s *AIService) buildOptimizeSystemPrompt(req *OptimizeCaseRequest) string {
	langPrompt := "使用中文"
	if req.TargetLang == "en-US" {
		langPrompt = "Use English"
	}

	optimizeDesc := ""
	switch req.OptimizeType {
	case "enhance":
		optimizeDesc = "增强用例的完整性和覆盖率"
	case "simplify":
		optimizeDesc = "简化用例，去除冗余步骤"
	case "translate":
		optimizeDesc = "翻译用例内容到目标语言"
	case "format":
		optimizeDesc = "格式化用例，统一术语和表达"
	}

	return fmt.Sprintf(`你是一个专业的软件测试工程师。请优化给定的测试用例。

优化目标：%s
语言要求：%s

返回JSON格式的优化后用例。`, optimizeDesc, langPrompt)
}

func (s *AIService) buildOptimizeUserPrompt(req *OptimizeCaseRequest) string {
	caseJSON, _ := json.MarshalIndent(req.Case, "", "  ")
	return fmt.Sprintf("原始用例：\n%s", string(caseJSON))
}

func (s *AIService) callAI(systemPrompt, userPrompt string) (string, error) {
	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": userPrompt},
	}

	body := map[string]interface{}{
		"model":    s.model,
		"messages": messages,
		"temperature": 0.7,
	}

	bodyBytes, _ := json.Marshal(body)
	url := s.baseURL + "/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("AI API error: %d - %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return result.Choices[0].Message.Content, nil
}

func (s *AIService) parseCasesFromResponse(response string) ([]GeneratedCase, error) {
	// 尝试直接解析
	var result struct {
		Cases []GeneratedCase `json:"cases"`
	}

	if err := json.Unmarshal([]byte(response), &result); err == nil && len(result.Cases) > 0 {
		return result.Cases, nil
	}

	// 尝试提取JSON部分
	start := findJSONStart(response)
	end := findJSONEnd(response)
	if start >= 0 && end > start {
		jsonStr := response[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
			return result.Cases, nil
		}
	}

	// 返回默认用例
	return []GeneratedCase{
		{
			Title:        "默认测试用例",
			Priority:     "P2",
			Precondition: "系统正常运行",
			Steps: []CaseStep{
				{StepNum: 1, Action: "执行测试操作", Expected: "验证结果正确"},
			},
		},
	}, nil
}

func (s *AIService) parseCaseFromResponse(response string) (*GeneratedCase, error) {
	cases, err := s.parseCasesFromResponse(response)
	if err != nil || len(cases) == 0 {
		return nil, err
	}
	return &cases[0], nil
}

func findJSONStart(s string) int {
	for i, c := range s {
		if c == '{' || c == '[' {
			return i
		}
	}
	return -1
}

func findJSONEnd(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '}' || s[i] == ']' {
			return i
		}
	}
	return -1
}

var _ = fmt.Sprintf