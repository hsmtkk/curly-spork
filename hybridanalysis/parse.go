package hybridanalysis

import "go.uber.org/zap"

type Parser struct {
	sugar *zap.SugaredLogger
}

func NewParser(sugar *zap.SugaredLogger) *Parser {
	return &Parser{sugar}
}

func (p *Parser) ParseSubmitFile(respBytes []byte) (string, error) {
	return "taskID", nil
}
