package utils

import (
	"fmt"
	"io/fs"
	"testing"
	"path"
	//"path/filepath"
	"testing/fstest"
)

func createTestFs(root string, failing bool) fstest.MapFS {
	type filePair struct {
		name string
		isDir bool
	}
	internalFiles := []filePair{
		//{ "go.mod", false },
		{ "go.sum", false },
		{ "cmd", true },
		{ ".env", false },
		{ "internal/bot", true },
		{ "internal/cache", true },
		{ "internal/sql", true },
		{ "internal/utils", true },
		{ "internal/utils/utils_test.go", false },
	}

	if !failing {
		internalFiles = append(internalFiles, filePair{
			name: "go.mod",
			isDir: false,
		})
	}

	fsMap := make(fstest.MapFS)

	join := func(base string) string {
		return path.Join(root, base)
	}

	for _, pair := range internalFiles {
		mFile := new(fstest.MapFile)

		if pair.isDir {
			mFile.Mode = fs.ModeDir
		} else {
			mFile.Data = []byte{ '\x00' }
		}

		fsMap[join(pair.name)] = mFile
	}

	return fsMap
}

func createInputPath(root, current string) string {
	return path.Join(root, current)
}

func TestFindGoProjectRootFs(t *testing.T) {
	expectedRoot := "nba-daily-fantasy-bot"

	fs := createTestFs(expectedRoot, false)
	failingFs := createTestFs(expectedRoot, true)

	testCases := []struct{
		expected string
		input string
	}{
		{ expectedRoot, "internal/utils" },
		{ expectedRoot, "internal/bot" },
		{ expectedRoot, "." },
	}

	for i, tt := range testCases {
		testName := fmt.Sprintf("test %d", i)
		t.Run(testName, func(t *testing.T) {
			input := createInputPath(expectedRoot, tt.input)

			root, ok, err := FindGoProjectRootFs(fs, input)

			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			if !ok {
				t.Error("expected to find root")
				t.FailNow()
			}

			if root != expectedRoot {
				t.Errorf("incorrect root returned. expected %s got %s", expectedRoot, root)
			}
		})

		t.Run(testName + "_failing", func(t *testing.T) {
			input := createInputPath(expectedRoot, tt.input)

			_, ok, err := FindGoProjectRootFs(failingFs, input)

			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			if ok {
				t.Errorf("expected to fail the test")
			}

		})
	}
}
