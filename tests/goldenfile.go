package tests

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var shouldUpdate = flag.Bool("update", false, "")

func GetGoldenFilePath(filePath string) string {
	return path.Join("testdata", fmt.Sprintf("%s.json", filePath))
}

func AssertJSONResponse(t *testing.T, fileName string, actual string) {
	t.Helper()

	filePath := GetGoldenFilePath(fileName)
	if *shouldUpdate {
		err := os.WriteFile(filePath, []byte(actual), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	var actualMap map[string]interface{}
	err := json.Unmarshal([]byte(actual), &actualMap)
	if err != nil {
		t.Fatal(err)
	}
	var expectedMap map[string]interface{}
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(fileContent, &expectedMap)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, expectedMap, actualMap)
}
