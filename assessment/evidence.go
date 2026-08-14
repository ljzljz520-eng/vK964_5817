package assessment

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrEvidenceRead = errors.New("证据读取失败")

type Evidence struct {
	Summary string `json:"summary"`
	Source  string `json:"source"`
}

type EvidenceReader interface {
	Read(path string) (Evidence, error)
}

type FileEvidenceReader struct{}

func (FileEvidenceReader) Read(path string) (Evidence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: %s: %v", ErrEvidenceRead, path, err)
	}

	var evidence Evidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		return Evidence{}, fmt.Errorf("%w: %s: %v", ErrEvidenceRead, path, err)
	}
	if strings.TrimSpace(evidence.Summary) == "" {
		return Evidence{}, fmt.Errorf("%w: %s: 缺少证据摘要", ErrEvidenceRead, path)
	}
	return evidence, nil
}

type RiskEvidenceAggregator struct {
	reader    EvidenceReader
	threshold int
}

func NewRiskEvidenceAggregator(reader EvidenceReader, threshold int) *RiskEvidenceAggregator {
	return &RiskEvidenceAggregator{reader: reader, threshold: threshold}
}

func (a *RiskEvidenceAggregator) Summarize(risks []RiskAssessment) ([]HighRiskItem, error) {
	items := make([]HighRiskItem, 0, len(risks))
	for _, risk := range risks {
		if risk.Score < a.threshold {
			continue
		}

		evidence, err := a.reader.Read(risk.EvidencePath)
		if err != nil {
			return nil, fmt.Errorf("汇总%s高风险证据: %w", risk.Area, err)
		}

		items = append(items, HighRiskItem{
			Area:            risk.Area,
			Score:           risk.Score,
			Risk:            risk.Risk,
			EvidencePath:    risk.EvidencePath,
			EvidenceSummary: evidence.Summary,
		})
	}
	return items, nil
}
