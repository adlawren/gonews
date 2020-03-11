package lib

import (
	"errors"
	"gonews/db"
	"gonews/fs"
	"gonews/item"
	"gonews/lib"
	"gonews/mock_config"
	"gonews/mock_db"
	"gonews/mock_fs"
	"gonews/mock_lib"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mmcdole/gofeed"
)

func TestTimestampFile_ParseReturnsNilTimeWhenFileDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)

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

func TestTimestampFile_ParseReturnsNilWhenFileOpenFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)
	mockFile := mock_fs.NewMockFile(ctrl)

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

func TestTimestampFile_ParseReturnsNilWhenFileReadFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)
	mockFile := mock_fs.NewMockFile(ctrl)

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

func TestTimestampFile_ParseReturnsNilWhenTimeParseFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)
	mockFile := mock_fs.NewMockFile(ctrl)

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

func TestTimestampFile_ParseReturnsTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)
	mockFile := mock_fs.NewMockFile(ctrl)

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

func TestTimestampFile_UpdateReturnsErrorWhenOpenFileFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)

	newTime := time.Now()
	expectedError := os.ErrPermission
	var expectedFileMode os.FileMode = 0644

	mockFS.EXPECT().OpenFile(gomock.Eq(testTimestampFilePath), gomock.Eq(os.O_RDWR|os.O_CREATE), gomock.Eq(expectedFileMode)).Return(nil, expectedError)

	if err := testTimestampFile.Update(mockFS, &newTime); err != expectedError {
		t.Errorf("update error is %v; expected %v\n", err, expectedError)
		t.Fail()

	}
}

func TestTimestampFile_UpdateReturnsErrorWhenWriteStringFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)
	mockFile := mock_fs.NewMockFile(ctrl)

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

func TestTimestampFile_UpdateReturnsErrorWhenCloseFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)
	mockFile := mock_fs.NewMockFile(ctrl)

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

func TestTimestampFile_UpdateReturnsNilWhenTimeUpdated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTimestampFilePath := "test_file"
	testTimestampFile := &lib.TimestampFile{Path: testTimestampFilePath}

	mockFS := mock_fs.NewMockFS(ctrl)
	mockFile := mock_fs.NewMockFile(ctrl)

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
	testFeed1["url"] = "TestUrl"
	testFeed1["tags"] = make([]interface{}, 2, 2)

	testFeed1TagsArray := testFeed1["tags"].([]interface{})
	testFeed1TagsArray[0] = "tag1"
	testFeed1TagsArray[1] = "tag2"

	feeds = append(feeds, testFeed1)

	return feeds
}

func mockFeedItem() *item.Item {
	var nilTime time.Time
	return &item.Item{
		Person: gofeed.Person{
			Name:  "TestAuthorName",
			Email: "TestAuthorEmail",
		},
		Title:       "TestTitle",
		Description: "TestDescription",
		Link:        "TestLink",
		Published:   nilTime,
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

func TestFetchFeeds_CreatesFeedItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	feedURL := "TestUrl"
	feedsConfig := mockFeedsConfig()
	f := mockGoFeed()
	feedItem := mockFeedItem()

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)
	mockGofeedURLParser.EXPECT().ParseURL(gomock.Eq(feedURL)).Return(f, nil)

	mockDB.EXPECT().FirstOrCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(out interface{}, where ...interface{}) db.DB {
		newFeedItem := where[0].(*item.Item)
		if *newFeedItem != *feedItem {
			t.Errorf("FeedItem is %v; expected %v\n", newFeedItem, feedItem)
			t.Fail()
		}

		return mockDB
	})

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockGofeedURLParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeeds_ReturnsNoErrorWhenFeedTagsMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	feedURL := "TestUrl"
	feedsConfig := mockFeedsConfig()
	f := mockGoFeed()
	feedItem := mockFeedItem()

	delete(feedsConfig[0].(map[string]interface{}), "tags")

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)
	mockGofeedURLParser.EXPECT().ParseURL(gomock.Eq(feedURL)).Return(f, nil)

	mockDB.EXPECT().FirstOrCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(out interface{}, where ...interface{}) db.DB {
		newFeedItem := where[0].(*item.Item)
		if *newFeedItem != *feedItem {
			t.Errorf("FeedItem is %v; expected %v\n", newFeedItem, feedItem)
			t.Fail()
		}

		return mockDB
	})

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockGofeedURLParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeeds_ReturnsErrorWhenFeedUrlMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	feedsConfig := mockFeedsConfig()
	delete(feedsConfig[0].(map[string]interface{}), "url")

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockGofeedURLParser, mockDB); err == nil {
		t.Errorf("FetchFeeds error is nil; expected an error\n")
		t.Fail()
	}
}

func TestFetchFeeds_DoesNotCreateFeedItemsWhenUrlUnparsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	feedURL := "TestUrl"
	feedsConfig := mockFeedsConfig()

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)
	mockGofeedURLParser.EXPECT().ParseURL(gomock.Eq(feedURL)).Return(nil, errors.New(""))
	mockDB.EXPECT().FirstOrCreate(gomock.Any(), gomock.Any()).Times(0)

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockGofeedURLParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeeds_DoesNotCreateFeedItemsWhenFeedIsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	feedURL := "TestUrl"
	feedsConfig := mockFeedsConfig()
	feed := mockGoFeed()
	feed.Items = []*gofeed.Item{}

	mockAppConfig.EXPECT().Get(gomock.Eq("feeds")).Return(feedsConfig)
	mockGofeedURLParser.EXPECT().ParseURL(gomock.Eq(feedURL)).Return(feed, nil)
	mockDB.EXPECT().FirstOrCreate(gomock.Any(), gomock.Any()).Times(0)

	testFeedFetcher := &lib.DefaultFeedFetcher{}
	if err := testFeedFetcher.FetchFeeds(mockAppConfig, mockGofeedURLParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayFetchesFeedsWhenFetchPeriodElapsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockFS := mock_fs.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	currentTime := time.Now()

	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(&currentTime, nil)
	mockTimestamp.EXPECT().Update(gomock.Eq(mockFS), gomock.Any()).DoAndReturn(func(fs fs.FS, newTime *time.Time) error {
		if !newTime.After(currentTime) {
			t.Errorf("new time is %v, expected to be after %v\n", newTime, currentTime)
			t.Fail()
		}

		return nil
	})

	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_period")).Return("5s")
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_delay")).Return("1s")
	mockFeedFetcher.EXPECT().FetchFeeds(gomock.Eq(mockAppConfig), gomock.Eq(mockGofeedURLParser), gomock.Eq(mockDB)).Return(nil)

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockGofeedURLParser, mockDB); err != nil {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, nil)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayReturnsErrorWhenTimestampUnparsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockFS := mock_fs.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	expectedError := errors.New("")
	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(nil, expectedError)

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockGofeedURLParser, mockDB); err != expectedError {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, expectedError)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayReturnsErrorWhenFetchPeriodUnparsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockFS := mock_fs.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	var nilTime time.Time

	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(&nilTime, nil)
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_period")).Return("invalid-string")

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockGofeedURLParser, mockDB); err == nil {
		t.Errorf("FetchFeeds returned %v; expected error\n", err)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayReturnsErrorWhenFetchDelayUnparsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockFS := mock_fs.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	var nilTime time.Time

	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(&nilTime, nil)
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_period")).Return("1s")
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_delay")).Return("invalid-string")

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockGofeedURLParser, mockDB); err == nil {
		t.Errorf("FetchFeeds returned %v; expected error\n", err)
		t.Fail()
	}
}

func TestFetchFeedsAfterDelayReturnsErrorWhenFeedsUnfetched(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mock_config.NewMockConfig(ctrl)
	mockFS := mock_fs.NewMockFS(ctrl)
	mockTimestamp := mock_lib.NewMockPersistentTimestamp(ctrl)
	mockFeedFetcher := mock_lib.NewMockFeedFetcher(ctrl)
	mockGofeedURLParser := mock_lib.NewMockGofeedURLParser(ctrl)
	mockDB := mock_db.NewMockDB(ctrl)

	var nilTime time.Time

	mockTimestamp.EXPECT().Parse(gomock.Eq(mockFS)).Return(&nilTime, nil)
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_period")).Return("1s")
	mockAppConfig.EXPECT().GetString(gomock.Eq("feed_fetch_delay")).Return("1s")

	expectedError := errors.New("")
	mockFeedFetcher.EXPECT().FetchFeeds(gomock.Eq(mockAppConfig), gomock.Eq(mockGofeedURLParser), gomock.Eq(mockDB)).Return(expectedError)

	if err := lib.FetchFeedsAfterDelay(mockAppConfig, mockFS, mockTimestamp, mockFeedFetcher, mockGofeedURLParser, mockDB); err != expectedError {
		t.Errorf("FetchFeeds error is %v; expected %v\n", err, expectedError)
		t.Fail()
	}
}
