package main

import (
	"os"
	"testing"

	"github.com/saxypandabear/digimonql/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIntOrDefault(t *testing.T) {
	const key = "TEST_GET_INT_OR_DEFAULT"

	t.Run("unset falls back to default", func(t *testing.T) {
		os.Unsetenv(key)
		assert.Equal(t, 42, getIntOrDefault(key, 42))
	})

	t.Run("valid int overrides default", func(t *testing.T) {
		t.Setenv(key, "7")
		assert.Equal(t, 7, getIntOrDefault(key, 42))
	})

	t.Run("non-numeric value falls back to default", func(t *testing.T) {
		t.Setenv(key, "not-a-number")
		assert.Equal(t, 42, getIntOrDefault(key, 42))
	})

	t.Run("empty value falls back to default", func(t *testing.T) {
		t.Setenv(key, "")
		assert.Equal(t, 42, getIntOrDefault(key, 42))
	})
}

func TestLoadLocalData(t *testing.T) {
	digimons := loadLocalData()

	require.NotEmpty(t, digimons, "expected the bundled data/digimon.json to contain entries")
	for _, d := range digimons {
		assert.NotEmpty(t, d.ID)
		assert.NotEmpty(t, d.Name)
	}
}

func TestInstantiateDatabase_FallsBackToLocalRepository(t *testing.T) {
	previous, wasSet := os.LookupEnv(MongoUrlKey)
	os.Unsetenv(MongoUrlKey)
	defer func() {
		if wasSet {
			os.Setenv(MongoUrlKey, previous)
		}
	}()

	repo := instantiateDatabase()

	localRepo, ok := repo.(*db.LocalDigimonRepository)
	require.True(t, ok, "expected a LocalDigimonRepository when MONGO_URL is unset")
	assert.NotEmpty(t, localRepo.Digimons)
}
