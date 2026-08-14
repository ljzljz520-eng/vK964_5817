package assessment

import "time"

type Area string

const (
	AreaMachine       Area = "机台"
	AreaInk           Area = "油墨区"
	AreaPaperStorage  Area = "纸库"
	AreaFinishedGoods Area = "成品区"
)

const HighRiskThreshold = 70

type CompanyProfile struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	UnifiedSocialCreditCode string `json:"unifiedSocialCreditCode"`
	Address                 string `json:"address"`
	SafetyManager           string `json:"safetyManager"`
}

type RiskAssessment struct {
	Area         Area   `json:"area"`
	Score        int    `json:"score"`
	Risk         string `json:"risk"`
	EvidencePath string `json:"evidencePath"`
}

type AssessmentInput struct {
	Company CompanyProfile   `json:"company"`
	Risks   []RiskAssessment `json:"risks"`
}

type HighRiskItem struct {
	Area            Area   `json:"area"`
	Score           int    `json:"score"`
	Risk            string `json:"risk"`
	EvidencePath    string `json:"evidencePath"`
	EvidenceSummary string `json:"evidenceSummary"`
}

type CorrectiveAction struct {
	Area        Area      `json:"area"`
	Issue       string    `json:"issue"`
	Requirement string    `json:"requirement"`
	Priority    string    `json:"priority"`
	DueAt       time.Time `json:"dueAt"`
}

type TimelineEvent struct {
	At     time.Time `json:"at"`
	Event  string    `json:"event"`
	Detail string    `json:"detail"`
}

type Report struct {
	ID                string             `json:"id"`
	Company           CompanyProfile     `json:"company"`
	Assessments       []RiskAssessment   `json:"assessments"`
	HighRiskItems     []HighRiskItem     `json:"highRiskItems"`
	CorrectiveActions []CorrectiveAction `json:"correctiveActions"`
	Timeline          []TimelineEvent    `json:"timeline"`
	Status            string             `json:"status"`
	SubmittedAt       time.Time          `json:"submittedAt"`
}

func RequiredAreas() []Area {
	return []Area{AreaMachine, AreaInk, AreaPaperStorage, AreaFinishedGoods}
}
