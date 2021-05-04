package magicbytes_test

import (
	"context"
	"testing"
	"time"

	"github.com/irukeru/binalyze-go-coding-challange/pkg/magicbytes"
)

const XLS_PATH = "../../test/file_example_XLS_10.xls"
const INVALID_XLS_PATH = "./file_example_XLS_10.xls"
const CURRENT_PATH = "./"
const CURRENT_INVALID_PATH = "./noDir"
const EMPTY_FILE_PATY = "../../test/empty_file.txt"

var XLS_FILE_META = magicbytes.Meta{"xls", []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, 0}
var XLS_FILE_META_WITH_OFFSET = magicbytes.Meta{"xls", []byte{0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, 1}

var META_ARRAY = []*magicbytes.Meta{
	{Type: "xls", Bytes: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, Offset: 0},
	{Type: "jpg", Bytes: []byte{0xFF, 0xD8}, Offset: 0},
	{Type: "png", Bytes: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, Offset: 0},
}

func TestWalkDirValidPathSuccess(t *testing.T) {

	t.Run("Should get filepaths successfully", func(t *testing.T) {

		testFilePathChan := make(chan string)
		defer close(testFilePathChan)

		go func() {
			for path := range testFilePathChan {
				if path == "" || len(path) == 0 {
					t.Error("path should not be empty")
				}
			}
		}()

		err := magicbytes.WalkDir(context.Background(), CURRENT_PATH, testFilePathChan)
		if err != nil {
			t.Error("WalkDir error: ", err)
		}
	})
}

func TestWalkDirInvalidPathSuccess(t *testing.T) {

	t.Run("Should return no error with invalid path", func(t *testing.T) {

		testFilePathChan := make(chan string)
		defer close(testFilePathChan)

		err := magicbytes.WalkDir(context.Background(), CURRENT_INVALID_PATH, testFilePathChan)
		if err != nil {
			t.Error("WalkDir error: ", err)
		}
	})
}

func TestWalkDirContextCancelSuccess(t *testing.T) {

	t.Run("Should cancel search successfully", func(t *testing.T) {

		testFilePathChan := make(chan string)
		defer close(testFilePathChan)

		ctx, _ := context.WithCancel(context.Background())

		err := magicbytes.WalkDir(ctx, CURRENT_INVALID_PATH, testFilePathChan)
		if err != nil {
			t.Error("WalkDir error: ", err)
		}
	})
}

func TestCheckMetaDataSuccess(t *testing.T) {

	result := magicbytes.CheckMetaData(XLS_PATH, XLS_FILE_META)

	if !result {
		t.Errorf(" file should have found")
	}
}

func TestCheckMetaDataWithOffetSuccess(t *testing.T) {

	result := magicbytes.CheckMetaData(XLS_PATH, XLS_FILE_META_WITH_OFFSET)

	if !result {
		t.Errorf(" file should have found")
	}
}

func TestCheckMetaDataFailureUnableToOpenFile(t *testing.T) {

	result := magicbytes.CheckMetaData(INVALID_XLS_PATH, XLS_FILE_META)

	if result {
		t.Errorf("file should not have found")
	}
}

func TestCheckMetaDataFailureFileSize(t *testing.T) {

	result := magicbytes.CheckMetaData(EMPTY_FILE_PATY, XLS_FILE_META)

	if result {
		t.Errorf("file should be empty")
	}
}

func TestFindMatchSuccessMatch(t *testing.T) {

	result, status := magicbytes.FindMatch(XLS_PATH, META_ARRAY)

	if result != "xls" {
		t.Error("file type should be xls")
	}

	if !status {
		t.Error("there should be a file match")
	}
}

func TestFindMatchSuccessNoMatch(t *testing.T) {

	t.Run("Find no match successfully", func(t *testing.T) {

		result, status := magicbytes.FindMatch(CURRENT_PATH, META_ARRAY)

		if result != "" {
			t.Error("file type should be empty string")
		}

		if status {
			t.Error("there should not be a file match")
		}

	})
}

func TestFindMatchWorkerOnMatchReturnFalseSuccess(t *testing.T) {

	t.Run("Should run Find Match Worker successfully", func(t *testing.T) {

		done := make(chan bool)

		testFilePathChan := make(chan string)
		defer close(testFilePathChan)

		go func() {
			magicbytes.FindMatchWorker(testFilePathChan, func(path, metaType string) bool {
				return false
			}, META_ARRAY)

			done <- true
		}()

		testFilePathChan <- XLS_PATH

		<-done
	})
}

func TestFindMatchWorkerOnMatchReturnTrueSuccess(t *testing.T) {

	t.Run("Should run Find Match Worker successfully", func(t *testing.T) {

		done := make(chan bool)

		testFilePathChan := make(chan string)

		go func() {
			magicbytes.FindMatchWorker(testFilePathChan, func(path, metaType string) bool {
				close(testFilePathChan)
				return true
			}, META_ARRAY)

			done <- true
		}()

		testFilePathChan <- XLS_PATH

		<-done
	})
}

func TestSearchSuccessfully(t *testing.T) {

	t.Run("Should search successfully", func(t *testing.T) {

		fileTypes := []*magicbytes.Meta{
			{Type: "xls", Bytes: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, Offset: 0},
			{Type: "jpg", Bytes: []byte{0xFF, 0xD8}, Offset: 0},
			{Type: "png", Bytes: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, Offset: 0},
		}

		magicbytes.Search(context.Background(), XLS_PATH, fileTypes, func(path, metaType string) bool {
			return true
		})
	})
}

func TestSearchCancelContextSuccessfully(t *testing.T) {

	t.Run("Should cancel context successfully", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		fileTypes := []*magicbytes.Meta{
			{Type: "xls", Bytes: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, Offset: 0},
			{Type: "jpg", Bytes: []byte{0xFF, 0xD8}, Offset: 0},
			{Type: "png", Bytes: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, Offset: 0},
		}

		go func() {
			cancel()

			err := magicbytes.Search(ctx, XLS_PATH, fileTypes, func(path, metaType string) bool {
				return true
			})

			if err == nil || err.Error() != context.Canceled.Error() {
				t.Errorf("unexpected or nil error received")
			}
		}()

		<-ctx.Done()
	})
}

func TestSearchDeadlineExceededSuccessfully(t *testing.T) {

	t.Run("Should handle deadline is exceeded successfully", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now())
		defer cancel()

		fileTypes := []*magicbytes.Meta{
			{Type: "xls", Bytes: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, Offset: 0},
			{Type: "jpg", Bytes: []byte{0xFF, 0xD8}, Offset: 0},
			{Type: "png", Bytes: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, Offset: 0},
		}

		err := magicbytes.Search(ctx, XLS_PATH, fileTypes, func(path, metaType string) bool {
			return true
		})

		if err == nil || err != context.DeadlineExceeded {
			t.Errorf("unexpected or nil error received")
		}
	})
}

func TestSearchMetaArraySizeExceededErrorSuccessfully(t *testing.T) {

	t.Run("Should return meta array size exceeded error successfully", func(t *testing.T) {

		fileTypes := make([]*magicbytes.Meta, magicbytes.MaxMetaArrayLength+1)

		err := magicbytes.Search(context.Background(), XLS_PATH, fileTypes, func(path, metaType string) bool {
			return true
		})

		if err == nil || err != magicbytes.ErrMetaArrayLengthExceeded {
			t.Errorf("unexpected or nil error received")
		}
	})
}

func TestSearchWithEmptyMetaArraySuccessfully(t *testing.T) {

	t.Run("Should make not search since meta array is empty", func(t *testing.T) {

		fileTypes := make([]*magicbytes.Meta, 0)

		err := magicbytes.Search(context.Background(), XLS_PATH, fileTypes, func(path, metaType string) bool {
			return true
		})

		if err != nil {
			t.Errorf("error received")
		}
	})
}
