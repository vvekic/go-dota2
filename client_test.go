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

func (s *DotaSuite) SetupSuite() {
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

func (s *DotaSuite) TestProfileCard() {
	pc, err := s.c.ProfileCard(296049278)
	if !s.NoError(err) {
		s.T().FailNow()
	}
	s.Equal(uint32(296049278), pc.GetAccountId())
}

func TestDotaSuite(t *testing.T) {
	suite.Run(t, new(DotaSuite))
}
