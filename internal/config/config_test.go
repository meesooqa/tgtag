package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	c, err := Load("testdata/config.yml")

	require.NoError(t, err)

	assert.IsType(t, &MongoConfig{}, c.Mongo)
	assert.Equal(t, "mongodb://localhost:27017", c.Mongo.URI)
	assert.Equal(t, "database_name", c.Mongo.Database)
	assert.Equal(t, "messages_collection_name", c.Mongo.CollectionMessages)

	assert.IsType(t, &SystemConfig{}, c.System)
	assert.Equal(t, "test/data", c.System.DataPath)

	assert.IsType(t, &ServerConfig{}, c.Server)
	assert.Equal(t, 12345, c.Server.Port)
}

func TestLoadConfigNotFoundFile(t *testing.T) {
	r, err := Load("/tmp/1b154315-b263-4952-b2b8-fad031f0df4f.txt")
	assert.Nil(t, r)
	assert.EqualError(t, err, "open /tmp/1b154315-b263-4952-b2b8-fad031f0df4f.txt: no such file or directory")
}

func TestLoadConfigInvalidYaml(t *testing.T) {
	r, err := Load("testdata/file.txt")

	assert.Nil(t, r)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `Not Yaml` into config.Conf")
}
