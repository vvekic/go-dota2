package dota2

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DotaSuite struct {
	suite.Suite
	c *Client
}

func (s *DotaSuite) SetupTest() {
	s.c = NewClient()
	err := s.c.Connect(os.Getenv("STEAM_USERNAME"), os.Getenv("STEAM_PASSWORD"), os.Getenv("STEAM_SENTRY"), os.Getenv("STEAM_AUTH_CODE"))
	if !s.NoError(err) {
		s.T().FailNow()
	}
}

func (s *DotaSuite) TestMatchDetails() {
	md, err := s.c.MatchDetails(854233753)
	if !s.NoError(err) {
		s.T().FailNow()
	}
	s.Equal(uint64(854233753), md.GetMatch().GetMatchId())
}

func TestDotaSuite(t *testing.T) {
	suite.Run(t, new(DotaSuite))
}
