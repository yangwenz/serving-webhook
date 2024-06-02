package storage

import (
	"fmt"
	"github.com/HyperGAI/serving-webhook/utils"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var config = utils.Config{
	AWSBucket:          "hypergai-upload-tmp",
	AWSRegion:          "us-east-2",
	AWSS3UseAccelerate: false,
	AWSAccessKeyID:     "xxx",
	AWSSecretAccessKey: "xxx",
}

func TestUpload(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store, err := NewS3Store(config)
	require.NoError(t, err)

	filepath := "/Users/ywz/Downloads/input_2.png"
	file, err := os.Open(filepath)
	require.NoError(t, err)
	defer file.Close()

	location, err := store.Upload(file, "test_input_1.jpg")
	require.NoError(t, err)
	fmt.Println(location)
}

func TestPutObject(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store, err := NewS3Store(config)
	require.NoError(t, err)

	filepath := "/Users/ywz/Downloads/input_2.png"
	file, err := os.Open(filepath)
	require.NoError(t, err)
	defer file.Close()

	location, err := store.PutObject(file, "test_input_2.jpg")
	require.NoError(t, err)
	fmt.Println(location)
}
