package aws

import (
	"context"

	"github.com/project-n-oss/sidekick/integration_tests/aws/utils"

	"github.com/stretchr/testify/suite"
)

type AwsSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *AwsSuite) SetupSuite() {
	t := s.T()
	s.ctx = context.Background()
	utils.InitVariables(t, s.ctx)
}
