package main

import (
    "testing"
    "os"
    "github.com/stretchr/testify/assert"
)

func TestValidateMsg(t *testing.T) {
    err1 := validateMsg("example.com:v1.0.0")
    assert.NoError(t, err1)

    err2 := validateMsg("")
    assert.Error(t, err2)

    err3 := validateMsg("example.com:1.0.0")
    assert.Error(t, err3)

    err4 := validateMsg("example.com:v100")
    assert.Error(t, err4)
}

func TestGetDeploymentInfo(t *testing.T) {
    os.Setenv("DEPLOY_BASE_DIR", "/tmp")
    os.Setenv("TMP_DIR", "/tmp")
    os.Setenv("DOWNLOAD_REPO", "https://example.com/repo")
    os.Setenv("DOWNLOAD_FILE", "sample.zip")
    ok, err := getDeploymentInfo("example.com:v1.0.0")
    assert.Equal(t, DeploymentInfo{
        SiteName:"example.com",
        Version:"v1.0.0",
        BaseDir:"/tmp/example.com",
        TmpFilePath:"/tmp/example.com.zip",
        URL:"https://example.com/repo/releases/download/v1.0.0/sample.zip",
    }, ok)
    assert.NoError(t, err)

    ng1, err := getDeploymentInfo("example.com:1.0.0")
    assert.Equal(t, ng1, DeploymentInfo{})
    assert.Error(t, err)

    ng2, err := getDeploymentInfo("")
    assert.Equal(t, ng2, DeploymentInfo{})
    assert.Error(t, err)
}

func TestValidateDeploymentInfo(t *testing.T) {
    info := DeploymentInfo{
        SiteName:"example.com",
        Version:"v1.0.0",
        BaseDir:"/tmp/example.com",
        TmpFilePath:"/tmp/example.com.zip",
        URL:"https://example.com/repo/releases/download/v1.0.0/sample.zip",
    }
    err := validateDeploymentInfo(info)
    assert.NoError(t, err)

    ngInfo1 := DeploymentInfo{
        SiteName:"example.com",
        Version:"v1.0.0",
        BaseDir:"/tmp/example.com/error",
        TmpFilePath:"/tmp/example.com.zip",
        URL:"https://example.com/repo/releases/download/v1.0.0/sample.zip",
    }
    err1 := validateDeploymentInfo(ngInfo1)
    assert.Error(t, err1)

    ngInfo2 := DeploymentInfo{
        SiteName:"example.com",
        Version:"v1.0.0",
        BaseDir:"/tmp/example.com",
        TmpFilePath:"/error/example.com.zip",
        URL:"https://example.com/repo/releases/download/v1.0.0/sample.zip",
    }
    err2 := validateDeploymentInfo(ngInfo2)
    assert.Error(t, err2)
}
