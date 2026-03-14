package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testingStoreService = &StorageService{}

func init() {
	testingStoreService = InitializeStore()
}

func TestStoreInit(t *testing.T) {
	assert.NotNil(t, testingStoreService.redisClient)
}

func TestInsertionAndRetrieval(t *testing.T) {
	initialLink := "https://www.guru3d.com/news-story/spotted-ryzen-threadripper-pro-3995wx-processor-with-8-channel-ddr4,2.html"
	userUUID := "e0dba740-fc4b-4977-872c-d360239e6b1a"
	shortURL := "Jsz4k57oAX"

	err := SaveUrlMapping(shortURL, initialLink, userUUID)
	assert.NoError(t, err)

	retrievedLink, err := RetrieveLongUrl(shortURL)
	assert.NoError(t, err)
	assert.Equal(t, initialLink, retrievedLink)
}
