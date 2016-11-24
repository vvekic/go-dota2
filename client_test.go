package dota2

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DotaSuite struct {
	suite.Suite
	c *Client
}

func (s *DotaSuite) SetupSuite() {
	s.c = NewClient()
	err := s.c.Connect(os.Getenv("TEST_STEAM_USERNAME"), os.Getenv("TEST_STEAM_PASSWORD"), os.Getenv("TEST_STEAM_SENTRY"), os.Getenv("TEST_STEAM_AUTH_CODE"))
	if !s.NoError(err) {
		s.T().FailNow()
	}
}

func (s *DotaSuite) TestMatchDetails() {
	// s.T().SkipNow()
	for i := 0; i < 200; i++ {
		_, err := s.c.MatchDetails(2411900220)
		if !s.NoError(err) {
			s.T().FailNow()
		}

		log.Printf("i: %d", i)
		time.Sleep(time.Second)
	}
	// s.Equal(uint64(854233753), md.GetMatch().GetMatchId())
}

func (s *DotaSuite) TestProfileCard() {
	s.T().SkipNow()
	pc, err := s.c.ProfileCard(296049278)
	if !s.NoError(err) {
		s.T().FailNow()
	}
	s.Equal(uint32(296049278), pc.GetAccountId())
}

func (s *DotaSuite) TestMatches() {
	s.T().SkipNow()
	for i := 0; i < 200; i++ {
		ms, err := s.c.Matches(0, 100)
		if !s.NoError(err) {
			s.T().FailNow()
		}
		log.Printf("i: %d, matches: %d", i, len(ms.GetMatches()))
		time.Sleep(time.Second)
	}
}

func (s *DotaSuite) TestMatchMinimalDetails() {
	s.T().SkipNow()
	md, err := s.c.MatchMinimalDetails(2411939053, 2411960491, 2411900220, 2411948816, 2411924671)
	if !s.NoError(err) {
		s.T().FailNow()
	}
	log.Print(md)
}

func TestDotaSuite(t *testing.T) {
	suite.Run(t, new(DotaSuite))
}
