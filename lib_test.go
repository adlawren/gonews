package main

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/mmcdole/gofeed"
	"gonews/lib"
	"gonews/mock_lib"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	exitStatus := m.Run()

	os.Exit(exitStatus)
}

func TestTimestampFile_ParseShouldReturnNilTimeWhenFileDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)

	mockFS.EXPECT().Stat(gomock.Eq(testTimestampFilePath)).Return(nil, os.ErrNotExist)

	var nilTime time.Time
	if res, err := testTimestampFile.Parse(mockFS); !res.Equal(nilTime) {
		t.Errorf("parsed time is %v; expected %v\n", res, nilTime)
		t.Fail()

	} else if err != nil {
		t.Errorf("parse error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestTimestampFile_ParseShouldReturnNilWhenFileOpenFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)
	mockFile := mock_lib.NewMockFile(ctrl)

	expectedError := errors.New("test")

	mockFS.EXPECT().Stat(gomock.Eq(testTimestampFilePath)).Return(nil, nil)
	mockFS.EXPECT().Open(gomock.Eq(testTimestampFilePath)).Return(mockFile, expectedError)

	if res, err := testTimestampFile.Parse(mockFS); res != nil {
		t.Errorf("parsed time is %v; expected %v\n", res, nil)
		t.Fail()

	} else if err != expectedError {
		t.Errorf("parse error is %v; expected %v\n", err, expectedError)
		t.Fail()
	}
}

func TestTimestampFile_ParseShouldReturnNilWhenFileReadFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)
	mockFile := mock_lib.NewMockFile(ctrl)

	expectedError := errors.New("test")

	mockFS.EXPECT().Stat(gomock.Eq(testTimestampFilePath)).Return(nil, nil)
	mockFS.EXPECT().Open(gomock.Eq(testTimestampFilePath)).Return(mockFile, nil)
	mockFile.EXPECT().Read(gomock.Any()).Return(0, expectedError)

	if res, err := testTimestampFile.Parse(mockFS); res != nil {
		t.Errorf("parsed time is %v; expected %v\n", res, nil)
		t.Fail()

	} else if err != expectedError {
		t.Errorf("parse error is %v; expected %v\n", err, expectedError)
		t.Fail()
	}
}

func TestTimestampFile_ParseShouldReturnNilWhenTimeParseFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)
	mockFile := mock_lib.NewMockFile(ctrl)

	mockFS.EXPECT().Stat(gomock.Eq(testTimestampFilePath)).Return(nil, nil)
	mockFS.EXPECT().Open(gomock.Eq(testTimestampFilePath)).Return(mockFile, nil)
	mockFile.EXPECT().Read(gomock.Any()).DoAndReturn(func(bytes []byte) (int, error) {
		return 0, nil
	})

	if res, err := testTimestampFile.Parse(mockFS); res != nil {
		t.Errorf("parsed time is %v; expected %v\n", res, nil)
		t.Fail()

	} else if err == nil {
		t.Error("parse error is nil; expected non-nil")
		t.Fail()
	}
}

func TestTimestampFile_ParseShouldReturnTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)
	mockFile := mock_lib.NewMockFile(ctrl)

	expectedTime, _ := time.Parse(time.UnixDate, time.Now().Format(time.UnixDate))

	mockFS.EXPECT().Stat(gomock.Eq(testTimestampFilePath)).Return(nil, nil)
	mockFS.EXPECT().Open(gomock.Eq(testTimestampFilePath)).Return(mockFile, nil)
	mockFile.EXPECT().Read(gomock.Any()).DoAndReturn(func(bytes []byte) (int, error) {
		timeString := expectedTime.Format(time.UnixDate)
		for i := 0; i < len(timeString); i++ {
			bytes[i] = timeString[i]
		}
		return len(timeString), nil
	})

	if res, err := testTimestampFile.Parse(mockFS); !res.Equal(expectedTime) {
		t.Errorf("parsed time is %v; expected %v\n", res, expectedTime)
		t.Fail()

	} else if err != nil {
		t.Errorf("parse error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestTimestampFile_UpdateShouldReturnErrorWhenOpenFileFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)

	newTime := time.Now()
	expectedError := os.ErrPermission
	var expectedFileMode os.FileMode = 0644

	mockFS.EXPECT().OpenFile(gomock.Eq(testTimestampFilePath), gomock.Eq(os.O_RDWR|os.O_CREATE), gomock.Eq(expectedFileMode)).Return(nil, expectedError)

	if err := testTimestampFile.Update(mockFS, &newTime); err != expectedError {
		t.Errorf("update error is %v; expected %v\n", err, expectedError)
		t.Fail()

	}
}

func TestTimestampFile_UpdateShouldReturnErrorWhenWriteStringFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)
	mockFile := mock_lib.NewMockFile(ctrl)

	newTime := time.Now()
	expectedError := os.ErrPermission
	var expectedFileMode os.FileMode = 0644

	mockFS.EXPECT().OpenFile(gomock.Eq(testTimestampFilePath), gomock.Eq(os.O_RDWR|os.O_CREATE), gomock.Eq(expectedFileMode)).Return(mockFile, nil)
	mockFile.EXPECT().WriteString(gomock.Eq(newTime.Format(time.UnixDate))).Return(len(newTime.Format(time.UnixDate)), expectedError)

	if err := testTimestampFile.Update(mockFS, &newTime); err != expectedError {
		t.Errorf("update error is %v; expected %v\n", err, expectedError)
		t.Fail()

	}
}

func TestTimestampFile_UpdateShouldReturnErrorWhenCloseFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)
	mockFile := mock_lib.NewMockFile(ctrl)

	newTime := time.Now()
	expectedError := os.ErrPermission
	var expectedFileMode os.FileMode = 0644

	mockFS.EXPECT().OpenFile(gomock.Eq(testTimestampFilePath), gomock.Eq(os.O_RDWR|os.O_CREATE), gomock.Eq(expectedFileMode)).Return(mockFile, nil)
	mockFile.EXPECT().WriteString(gomock.Eq(newTime.Format(time.UnixDate))).Return(len(newTime.Format(time.UnixDate)), nil)
	mockFile.EXPECT().Close().Return(expectedError)

	if err := testTimestampFile.Update(mockFS, &newTime); err != expectedError {
		t.Errorf("update error is %v; expected %v\n", err, expectedError)
		t.Fail()

	}
}

func TestTimestampFile_UpdateShouldReturnNilWhenTimeUpdated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_lib.NewMockFS(ctrl)
	mockFile := mock_lib.NewMockFile(ctrl)

	newTime := time.Now()
	var expectedFileMode os.FileMode = 0644

	mockFS.EXPECT().OpenFile(gomock.Eq(testTimestampFilePath), gomock.Eq(os.O_RDWR|os.O_CREATE), gomock.Eq(expectedFileMode)).Return(mockFile, nil)
	mockFile.EXPECT().WriteString(gomock.Eq(newTime.Format(time.UnixDate))).Return(len(newTime.Format(time.UnixDate)), nil)
	mockFile.EXPECT().Close().Return(nil)

	if err := testTimestampFile.Update(mockFS, &newTime); err != nil {
		t.Errorf("update error is %v; expected %v\n", err, nil)
		t.Fail()

	}
}

// TODO: don't use 'mockXYZ' functions? Mock the modules directly using Gomock?

func mockFeedsConfig() []interface{} {
	var feeds []interface{}

	testFeed1 := make(map[string]interface{})

	var itemLimit int64 = 5
	testFeed1["item_limit"] = itemLimit

	testFeed1["url"] = "TestUrl"

	feeds = append(feeds, testFeed1)

	return feeds
}

func mockFeedItem() *lib.FeedItem {
	var nilTime time.Time
	return &lib.FeedItem{
		Title:       "TestTitle",
		Description: "TestDescription",
		Link:        "TestLink",
		Published:   nilTime,
		Url:         "TestUrl",
		AuthorName:  "TestAuthorName",
		AuthorEmail: "TestAuthorEmail",
		Hide:        false,
	}
}

func mockGoFeed() *gofeed.Feed {
	var nilTime time.Time
	return &gofeed.Feed{
		Items: []*gofeed.Item{
			&gofeed.Item{
				Title:           "TestTitle",
				Description:     "TestDescription",
				Link:            "TestLink",
				PublishedParsed: &nilTime,
				Author: &gofeed.Person{
					Name:  "TestAuthorName",
					Email: "TestAuthorEmail",
				},
			},
		},
	}
}

func TestFetchFeeds_ShouldCreateFeedItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	feedUrl := "TestUrl"
	feedsConfig := mockFeedsConfig()
	feed := mockGoFeed()
	feedItem := mockFeedItem()

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)
	mockFeedParser.EXPECT().ParseURL(gomock.Eq(feedUrl)).Return(feed, nil)

	mockDB.EXPECT().FirstOrCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(out interface{}, where ...interface{}) lib.DB {
		newFeedItem := where[0].(*lib.FeedItem)
		if *newFeedItem != *feedItem {
			t.Errorf("FeedItem is %v; expected %v\n", newFeedItem, feedItem)
			t.Fail()
		}

		return mockDB
	})

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockFeedParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeeds_ShouldReturnErrorWhenFeedUrlMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	feedsConfig := mockFeedsConfig()
	delete(feedsConfig[0].(map[string]interface{}), "url")

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockFeedParser, mockDB); err == nil {
		t.Errorf("FetchFeeds error is nil; expected an error\n")
		t.Fail()
	}
}

func TestFetchFeeds_ShouldNotCreateFeedItemsWhenUrlUnparsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	feedUrl := "TestUrl"
	feedsConfig := mockFeedsConfig()

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)
	mockFeedParser.EXPECT().ParseURL(gomock.Eq(feedUrl)).Return(nil, errors.New(""))
	mockDB.EXPECT().FirstOrCreate(gomock.Any(), gomock.Any()).Times(0)

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockFeedParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeeds_ShouldNotCreateFeedItemsWhenFeedIsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	feedUrl := "TestUrl"
	feedsConfig := mockFeedsConfig()
	feed := mockGoFeed()
	feed.Items = []*gofeed.Item{}

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)
	mockFeedParser.EXPECT().ParseURL(gomock.Eq(feedUrl)).Return(feed, nil)
	mockDB.EXPECT().FirstOrCreate(gomock.Any(), gomock.Any()).Times(0)

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockFeedParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayShouldFetchFeedsWhenFetchPeriodElapsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFS := mock_lib.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	currentTime := time.Now()

	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(&currentTime, nil)
	mockTimestamp.EXPECT().Update(gomock.Eq(mockFS), gomock.Any()).DoAndReturn(func(fs lib.FS, newTime *time.Time) lib.DB {
		if *newTime == currentTime {
			t.Errorf("new time is %v, expected to be unequal to old time\n", newTime)
			t.Fail()
		}

		return nil
	})

	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_period")).Return("5s")
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_delay")).Return("1s")
	mockFeedFetcher.EXPECT().FetchFeeds(gomock.Eq(mockAppConfig), gomock.Eq(mockFeedParser), gomock.Eq(mockDB)).Return(nil)

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockFeedParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayShouldReturnErrorWhenTimestampUnparsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFS := mock_lib.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	expectedError := errors.New("")
	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(nil, expectedError)

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockFeedParser, mockDB); err != expectedError {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, expectedError)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayShouldReturnErrorWhenFetchPeriodUnparsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFS := mock_lib.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	var nilTime time.Time

	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(&nilTime, nil)
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_period")).Return("invalid-string")

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockFeedParser, mockDB); err == nil {
		t.Errorf("FetchFeeds returned %v; expected error\n", err)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayShouldReturnErrorWhenFetchDelayUnparsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFS := mock_lib.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	var nilTime time.Time

	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(&nilTime, nil)
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_period")).Return("1s")
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_delay")).Return("invalid-string")

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockFeedParser, mockDB); err == nil {
		t.Errorf("FetchFeeds returned %v; expected error\n", err)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayShouldReturnErrorWhenFeedsUnfetched(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_lib.NewMockAppConfig(ctrl)
	mockFS := mock_lib.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockFeedParser := mock_lib.NewMockFeedParser(ctrl)
	mockDB := mock_lib.NewMockDB(ctrl)

	var nilTime time.Time

	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(&nilTime, nil)
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_period")).Return("1s")
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_delay")).Return("1s")

	expectedError := errors.New("")
	mockFeedFetcher.EXPECT().FetchFeeds(gomock.Eq(mockAppConfig), gomock.Eq(mockFeedParser), gomock.Eq(mockDB)).Return(expectedError)

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockFeedParser, mockDB); err != expectedError {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, expectedError)
		t.Fail()
	}
}
