package assessment_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"printshop-safety/assessment"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestSubmitAssessmentPublishesCompanyRisksActionsAndTimeline(t *testing.T) {
	dir := t.TempDir()
	input := completeAssessment(t, dir)
	publisher := &assessment.MemoryPublisher{}
	now := time.Date(2026, 8, 15, 9, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	service := assessment.NewService(
		assessment.NewRiskEvidenceAggregator(assessment.FileEvidenceReader{}, assessment.HighRiskThreshold),
		publisher,
		fixedClock{now: now},
	)

	report, err := service.Submit(input)
	if err != nil {
		t.Fatalf("提交完整评估返回错误: %v", err)
	}
	if report.Status != "已提交" {
		t.Errorf("评估状态 = %q, want 已提交", report.Status)
	}
	if report.Company != input.Company {
		t.Errorf("报告企业档案 = %#v, want %#v", report.Company, input.Company)
	}
	if len(report.Assessments) != 4 {
		t.Errorf("报告区域评估数 = %d, want 4", len(report.Assessments))
	}
	if len(report.HighRiskItems) != 2 {
		t.Errorf("高风险项数 = %d, want 2", len(report.HighRiskItems))
	}
	if len(report.CorrectiveActions) != 2 {
		t.Errorf("整改项数 = %d, want 2", len(report.CorrectiveActions))
	}
	if len(report.Timeline) != 3 {
		t.Errorf("报告时间线事件数 = %d, want 3", len(report.Timeline))
	}
	if !report.SubmittedAt.Equal(now) {
		t.Errorf("报告提交时间 = %s, want %s", report.SubmittedAt, now)
	}
	if got := publisher.Reports(); len(got) != 1 {
		t.Errorf("已发布报告数 = %d, want 1", len(got))
	}
}

func TestSubmitAssessmentRejectsDamagedHighRiskPaperEvidence(t *testing.T) {
	dir := t.TempDir()
	input := completeAssessment(t, dir)
	input.Risks[2].Score = assessment.HighRiskThreshold
	input.Risks[2].Risk = "纸垛遮挡消防设施"
	if err := os.WriteFile(input.Risks[2].EvidencePath, []byte("damaged evidence"), 0o600); err != nil {
		t.Fatalf("准备纸库证据失败: %v", err)
	}
	publisher := &assessment.MemoryPublisher{}
	service := assessment.NewService(
		assessment.NewRiskEvidenceAggregator(assessment.FileEvidenceReader{}, assessment.HighRiskThreshold),
		publisher,
		fixedClock{now: time.Date(2026, 8, 15, 9, 30, 0, 0, time.UTC)},
	)

	_, err := service.Submit(input)
	if !errors.Is(err, assessment.ErrEvidenceRead) {
		t.Errorf("提交结果 = %v, want 证据读取失败", err)
	}
	if got := publisher.Reports(); len(got) != 0 {
		t.Errorf("已发布报告数 = %d, want 0", len(got))
	}
}

func TestSubmitAssessmentRejectsMissingWorkArea(t *testing.T) {
	dir := t.TempDir()
	input := completeAssessment(t, dir)
	input.Risks = input.Risks[:3]
	publisher := &assessment.MemoryPublisher{}
	service := assessment.NewService(
		assessment.NewRiskEvidenceAggregator(assessment.FileEvidenceReader{}, assessment.HighRiskThreshold),
		publisher,
		fixedClock{now: time.Date(2026, 8, 15, 9, 30, 0, 0, time.UTC)},
	)

	_, err := service.Submit(input)
	if !errors.Is(err, assessment.ErrInvalidAssessment) {
		t.Errorf("提交结果 = %v, want 评估信息无效", err)
	}
	if got := publisher.Reports(); len(got) != 0 {
		t.Errorf("已发布报告数 = %d, want 0", len(got))
	}
}

func completeAssessment(t *testing.T, dir string) assessment.AssessmentInput {
	t.Helper()
	paths := make([]string, 4)
	for i, summary := range []string{"防护罩缺失", "溶剂桶未接地", "纸垛间距正常", "疏散通道畅通"} {
		paths[i] = filepath.Join(dir, fmt.Sprintf("evidence-%d.json", i))
		data := []byte(fmt.Sprintf(`{"summary":%q,"source":"现场巡检"}`, summary))
		if err := os.WriteFile(paths[i], data, 0o600); err != nil {
			t.Fatalf("准备本地证据失败: %v", err)
		}
	}

	return assessment.AssessmentInput{
		Company: assessment.CompanyProfile{
			ID:                      "printer-001",
			Name:                    "华彩印刷有限公司",
			UnifiedSocialCreditCode: "91310000TEST000001",
			Address:                 "上海市示例路 18 号",
			SafetyManager:           "陈安",
		},
		Risks: []assessment.RiskAssessment{
			{Area: assessment.AreaMachine, Score: 86, Risk: "设备防护罩缺失", EvidencePath: paths[0]},
			{Area: assessment.AreaInk, Score: 78, Risk: "易燃溶剂桶未接地", EvidencePath: paths[1]},
			{Area: assessment.AreaPaperStorage, Score: 42, Risk: "纸垛间距不足", EvidencePath: paths[2]},
			{Area: assessment.AreaFinishedGoods, Score: 30, Risk: "疏散标识褪色", EvidencePath: paths[3]},
		},
	}
}
