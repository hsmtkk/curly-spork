package hybridanalysis

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type Parser struct {
	sugar *zap.SugaredLogger
}

func NewParser(sugar *zap.SugaredLogger) *Parser {
	return &Parser{sugar}
}

type responseSchema struct {
	JobID         string `json:"job_id"`
	SubmissionID  string `json:"submission_id"`
	EnvironmentID int    `json:"environment_id"`
	SHA256        string `json:"sha256"`
}

func (p *Parser) ParseSubmitFile(respBytes []byte) (string, error) {
	rs := responseSchema{}
	if err := json.Unmarshal(respBytes, &rs); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON; %s; %w", string(respBytes), err)
	}
	return rs.JobID, nil
}
