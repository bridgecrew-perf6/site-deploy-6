package main

import (
    "github.com/google/logger"
    "github.com/nats-io/stan.go"
    "os"
    "regexp"
    "errors"
    "fmt"
    "path"
    "net/http"
    "io"
    "path/filepath"
)

type DeploymentInfo struct {
    SiteName string
    Version string
    BaseDir string
    TmpFilePath string
    URL string
}

func deploySite(message *stan.Msg) {
    msg := string(message.Data)
    if err := validateMsg(msg); err != nil {
        logger.Errorf("validation failed: %s", err)
        return
    }

    depInfo, err := getDeploymentInfo(msg)
    if err != nil {
        logger.Errorf("failed to parse message: %s", err)
        return
    }

    if err := validateDeploymentInfo(depInfo); err != nil {
        logger.Errorf("validation failed: %s", err)
        return
    }

    if err := downloadFile(depInfo.TmpFilePath, depInfo.URL); err != nil {
        logger.Errorf("failed to download file: %s", err)
        return
    }

/*
    if err := deploy(depInfo); err != nil {
        logger.Errorf("failed to download file: %s", err)
    }
*/
    return
}

func validateMsg(msg string) error {
    if len(msg) == 0 {
        return errors.New("message is empty")
    }

    r := regexp.MustCompile(`^[a-z0-9\.]+:v[0-9]+\.[0-9]+\.[0-9]+$`)
    if !r.MatchString(msg) {
        return fmt.Errorf("message does not match format: %s", msg)
    }

    return nil
}

func getDeploymentInfo(msg string) (DeploymentInfo, error) {
    r := regexp.MustCompile(`^([a-z0-9\.]+):(v[0-9]+\.[0-9]+\.[0-9]+)$`)
    match := r.FindStringSubmatch(msg)
    if match == nil {
        return DeploymentInfo{}, errors.New("could not get deployment info from received message")
    }

    depBaseDir := os.Getenv("DEPLOY_BASE_DIR")
    tmpDir := os.Getenv("TMP_DIR")
    url := fmt.Sprintf("%s/releases/download/%s/%s", os.Getenv("DOWNLOAD_REPO"), match[2], os.Getenv("DOWNLOAD_FILE"))
    tmpFileName := fmt.Sprintf("%s%s", match[1], filepath.Ext(os.Getenv("DOWNLOAD_FILE")))

    return DeploymentInfo{
        SiteName: match[1],
        Version: match[2],
        BaseDir: path.Join(path.Dir(depBaseDir), path.Base(depBaseDir), match[1]),
        TmpFilePath: path.Join(path.Dir(tmpDir), path.Base(tmpDir), tmpFileName),
        URL: url,
    }, nil
}

func validateDeploymentInfo(info DeploymentInfo) error {
    if _, err := os.Stat(info.BaseDir); err != nil {
        return fmt.Errorf("%s does not exist", info.BaseDir)
    }

    if _, err := os.Stat(path.Dir(info.TmpFilePath)); err != nil {
        return fmt.Errorf("%s does not exist", info.TmpFilePath)
    }

    return nil
}

func downloadFile(filepath string, url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    if _, err := io.Copy(out, resp.Body); err != nil {
        return err
    }
    return nil
}

