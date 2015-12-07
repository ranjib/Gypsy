// Package util provides utility methods for gypsy
package util

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/lxc/go-lxc.v2"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func MinimalEnv() []string {
	return []string{
		"SHELL=/bin/bash",
		"USER=root",
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/go/bin:/opt/gospace/bin",
		"PWD=/root",
		"EDITOR=vim",
		"LANG=en_US.UTF-8",
		"HOME=/root",
		"LANGUAGE=en_US",
		"LOGNAME=root",
		"GOPATH=/opt/gospace",
		"GOROOT=/opt/go",
	}
}

func PostFileFromContainer(ct *lxc.Container, src, url string) error {
	uuid, err := UUID()
	if err != nil {
		log.Errorf("Failed to generate uuid for temporary file name. Error: %v", err)
		return err
	}
	dst := filepath.Join("/tmp", uuid)
	cmd := []string{"cp", src, dst}
	_, e1 := ct.RunCommandStatus(cmd, lxc.DefaultAttachOptions)
	if e1 != nil {
		log.Errorf("Failed to execute: '%s' inside container '%s'", strings.Join(cmd, " "), ct.Name())
		return e1
	}
	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)
	fw, err := bodyWriter.CreateFormFile("artifact", dst)
	if err != nil {
		log.Errorf("Failed to write to buffer. Error: %v", err)
		return err
	}
	rootfs := ct.ConfigItem("lxc.rootfs")[0]
	fh, e := os.Open(filepath.Join(rootfs, dst))
	if e != nil {
		log.Errorf("Failed to open file %s. Error: %v", dst, e)
		return e
	}
	_, err = io.Copy(fw, fh)
	if err != nil {
		log.Errorf("Failed to copy file. Error: %v", err)
		return err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	resp, e2 := http.Post(url, contentType, bodyBuf)
	if e2 != nil {
		log.Errorf("Failed to perform http post. Error: %v", e1)
		return e2
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Errorf("Server responded with non 200 status code.")
		return fmt.Errorf("Non 200 response from server. Return code: %d", resp.StatusCode)
	}
	return nil
}
func CloneAndStartContainer(original, cloned string) (*lxc.Container, error) {
	orig, err := lxc.NewContainer(original)
	if err != nil {
		log.Errorf("Failed to initialize container object. Error: %v", err)
		return nil, err
	}
	if err := orig.Clone(cloned, lxc.CloneOptions{}); err != nil {
		log.Errorf("Failed to clone container %s as %s. Error: %v", original, cloned, err)
		return nil, err
	}
	ct, err := lxc.NewContainer(cloned)
	if err != nil {
		log.Errorf("Failed to clone container %s as %s. Error: %v", original, cloned, err)
	}
	if err := ct.Start(); err != nil {
		log.Errorf("Failed to start cloned container %s. Error: %v", cloned, err)
		return nil, err
	}
	log.Infof("Created container named: %s. Waiting for ip allocation", cloned)
	if _, err := ct.WaitIPAddresses(30 * time.Second); err != nil {
		log.Errorf("Failed to while waiting to start the container %s. Error: %v", cloned, err)
		return nil, err
	}
	return ct, nil
}
