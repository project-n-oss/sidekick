package gcp

import (
	"context"

	"github.com/project-n-oss/sidekick/integration_tests/gcp/utils"
	"github.com/stretchr/testify/suite"
)

type GcpSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *GcpSuite) SetupSuite() {
	t := s.T()
	s.ctx = context.Background()
	utils.InitVariables(t, s.ctx)
}
