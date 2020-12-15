package main

import (
    "os"
    "errors"
)

func validateEnv() error {
    if os.Getenv("DEPLOY_BASE_DIR") == "" {
        return errors.New("environment variable DEPLOY_BASE_DIR is required")
    }

    if os.Getenv("TMP_DIR") == "" {
        return errors.New("environment variable TMP_DIR is required")
    }

    if os.Getenv("DOWNLOAD_REPO") == "" {
        return errors.New("environment variable DOWNLOAD_REPO is required")
    }

    if os.Getenv("DOWNLOAD_FILE") == "" {
        return errors.New("environment variable DOWNLOAD_FILE is required")
    }

    return nil
}
