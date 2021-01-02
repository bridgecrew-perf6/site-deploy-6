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
    "archive/zip"
    "strings"
    "encoding/json"
    "bytes"
)

type DeploymentInfo struct {
    SiteName string
    Version string
    BaseDir string
    TmpFilePath string
    URL string
}

type SlackParams struct {
    Text string `json:"text"`
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

    if err := deleteCurrentContents(depInfo.BaseDir); err != nil {
        logger.Errorf("failed to delete file: %s", err)
        return
    }

    if err := unzip(depInfo.TmpFilePath, depInfo.BaseDir); err != nil {
        logger.Errorf("failed to unzip file: %s", err)
        return
    }

    if err := notifySlack(msg); err != nil {
        logger.Errorf("failed to notify slack: %s", err)
        return
    }

    logger.Infof("deploy completed: %s", msg)
    return
}

func notifySlack(msg string) error {
    if os.Getenv("SLACK_WEBHOOK") == "" {
        return errors.New("SLACK_WEBHOOK is not defined")
    }
    URL := os.Getenv("SLACK_WEBHOOK")

    hostname, err := os.Hostname()
    if err != nil {
        hostname = "<error>"
    }
    params := SlackParams{
        Text: fmt.Sprintf("deploy completed: %s, host: %s", msg, hostname),
    }

    js, err := json.Marshal(params)
    if err != nil {
        return err
    }

    res, err := http.Post(URL, "application/json", bytes.NewBuffer(js))
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.StatusCode < 200 || res.StatusCode >= 400 {
        return fmt.Errorf("status code NG: %v", res.StatusCode)
    }

    return nil
}

func deleteCurrentContents(dir string) error {
    d, err := os.Open(dir)
    if err != nil {
        return err
    }
    defer d.Close()

    fnames, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }

    for _, n := range fnames {
        err := os.RemoveAll(filepath.Join(dir, n))
        if err != nil {
            return err
        }
    }

    return nil
}

func unzip(src string, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        fpath := filepath.Join(dest, f.Name)

        if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
            return fmt.Errorf("%s: illegal file path", fpath)
        }

        if f.FileInfo().IsDir() {
            os.MkdirAll(fpath, os.ModePerm)
            continue
        }

        if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
            return err
        }

        outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }

        rc, err := f.Open()
        if err != nil {
            return err
        }

        _, err = io.Copy(outFile, rc)

        outFile.Close()
        rc.Close()

        if err != nil {
            return err
        }
    }

    return nil
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

