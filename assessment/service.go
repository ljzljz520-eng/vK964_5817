package assessment

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidAssessment = errors.New("评估信息无效")

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

type ReportPublisher interface {
	Publish(Report) error
}

type Service struct {
	aggregator *RiskEvidenceAggregator
	publisher  ReportPublisher
	clock      Clock
}

func NewService(aggregator *RiskEvidenceAggregator, publisher ReportPublisher, clock Clock) *Service {
	return &Service{aggregator: aggregator, publisher: publisher, clock: clock}
}

func (s *Service) Submit(input AssessmentInput) (Report, error) {
	if err := validateAssessment(input); err != nil {
		return Report{}, err
	}

	now := s.clock.Now()
	highRiskItems, err := s.aggregator.Summarize(input.Risks)
	if err != nil {
		return Report{}, err
	}

	actions := make([]CorrectiveAction, 0, len(highRiskItems))
	for _, item := range highRiskItems {
		actions = append(actions, CorrectiveAction{
			Area:        item.Area,
			Issue:       item.Risk,
			Requirement: "消除风险并上传复核证据",
			Priority:    "高",
			DueAt:       now.Add(7 * 24 * time.Hour),
		})
	}

	report := Report{
		ID:                fmt.Sprintf("%s-%s", input.Company.ID, now.UTC().Format("20060102T150405Z")),
		Company:           input.Company,
		Assessments:       append([]RiskAssessment(nil), input.Risks...),
		HighRiskItems:     highRiskItems,
		CorrectiveActions: actions,
		Timeline: []TimelineEvent{
			{At: now, Event: "评估已接收", Detail: "四个作业区域的风险评估已接收"},
			{At: now, Event: "高风险已汇总", Detail: fmt.Sprintf("汇总高风险项 %d 条", len(highRiskItems))},
			{At: now, Event: "报告已发布", Detail: "整改清单已随评估报告发布"},
		},
		Status:      "已提交",
		SubmittedAt: now,
	}
	if err := s.publisher.Publish(report); err != nil {
		return Report{}, fmt.Errorf("发布评估报告: %w", err)
	}
	return report, nil
}

func validateAssessment(input AssessmentInput) error {
	if strings.TrimSpace(input.Company.ID) == "" || strings.TrimSpace(input.Company.Name) == "" {
		return fmt.Errorf("%w: 企业档案缺少编号或名称", ErrInvalidAssessment)
	}
	if len(input.Risks) != len(RequiredAreas()) {
		return fmt.Errorf("%w: 必须填写四个作业区域", ErrInvalidAssessment)
	}

	required := make(map[Area]bool, len(RequiredAreas()))
	for _, area := range RequiredAreas() {
		required[area] = true
	}
	seen := make(map[Area]bool, len(input.Risks))
	for _, risk := range input.Risks {
		if !required[risk.Area] || seen[risk.Area] {
			return fmt.Errorf("%w: 作业区域重复或未知", ErrInvalidAssessment)
		}
		if risk.Score < 0 || risk.Score > 100 {
			return fmt.Errorf("%w: %s风险分值必须在 0 到 100 之间", ErrInvalidAssessment, risk.Area)
		}
		if strings.TrimSpace(risk.Risk) == "" || strings.TrimSpace(risk.EvidencePath) == "" {
			return fmt.Errorf("%w: %s缺少风险或本地证据", ErrInvalidAssessment, risk.Area)
		}
		seen[risk.Area] = true
	}
	return nil
}
