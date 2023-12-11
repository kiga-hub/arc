package tests

import (
	"context"
	"testing"

	"github.com/kiga-hub/common/micro"
	"github.com/kiga-hub/common/micro/component"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

type ServerTestSuite struct {
	suite.Suite
	component micro.IComponent
}

func (suite *ServerTestSuite) SetupTest() {
	m := &component.MockComponent{}
	m.On("PreInit", mock.Anything).Return(func(ctx context.Context) error {
		return nil
	})
	m.On("Init", mock.AnythingOfType("*micro.Server")).Return(func(server *micro.Server) error {
		return nil
	})
	m.On("PostInit", mock.Anything).Return(func(ctx context.Context) error {
		return nil
	})

	suite.component = m
}

func (suite *ServerTestSuite) TestExample() {
	appname := "test"
	appversion := "v1.0.0"
	s, err := micro.NewServer(appname, appversion, []micro.IComponent{suite.component})
	assert.NotNil(suite.T(), s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), appname, s.AppName)
	assert.Equal(suite.T(), appversion, s.AppVersion)
	err = s.Init()
	assert.NotNil(suite.T(), err)
}
