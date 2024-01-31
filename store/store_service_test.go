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
	assert.True(t, testingStoreService.redisClient != nil)
}

func TestInsertionAndRetrieval(t *testing.T) {
	initialLink := "https://www.guru3d.com/news-story/spotted-ryzen-threadripper-pro-3995wx-processor-with-8-channel-ddr4,2.html"
	userUUID := "e0dba740-fc4b-4977-872c-d360239e6b1a"
	shortURL := "Jsz4k57oAX"
	//Persist data mapping
	saveUrlMapping(shortURL, initialLink, userUUID)
	//Retrieve initial link
	retrievedLink := RetrieveLongUrl(shortURL)
	assert.Equal(t, initialLink, retrievedLink)
}
