package main

import (
    "testing"
    "os"
    "github.com/stretchr/testify/assert"
)

func TestValidateEnv(t *testing.T) {
    err := validateEnv()
    assert.NoError(t, err)

    os.Setenv("DEPLOY_BASE_DIR", "/home/kkodama/work/tmp")
    os.Setenv("TMP_DIR", "/tmp")
    os.Setenv("DOWNLOAD_REPO", "https://example.com/repo")
    os.Setenv("DOWNLOAD_FILE", "sample.zip")
    err2 := validateEnv()
    assert.NoError(t, err2)
}
