package fly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"os"
	"strings"
	"io"

	"crypto/tls"
	"net/http"

	"github.com/concourse/concourse-pipeline-resource/logger"
)

//go:generate counterfeiter . Command

type Command interface {
	Login(url string, teamName string, username string, password string, insecure bool) ([]byte, error)
	Pipelines() ([]string, error)
	GetPipeline(pipelineName string) ([]byte, error)
	SetPipeline(pipelineName string, configFilepath string, varsFilepaths []string, vars map[string]interface{}) ([]byte, error)
	DestroyPipeline(pipelineName string) ([]byte, error)
	UnpausePipeline(pipelineName string) ([]byte, error)
	ExposePipeline(pipelineName string) ([]byte, error)
}

type command struct {
	target        string
	logger        logger.Logger
	flyBinaryPath string
}

func NewCommand(target string, logger logger.Logger, flyBinaryPath string) Command {
	return &command{
		target:        target,
		logger:        logger,
		flyBinaryPath: flyBinaryPath,
	}
}

func (f command) Login(
	url string,
	teamName string,
	username string,
	password string,
	insecure bool,
) ([]byte, error) {
	args := []string{
		"login",
		"-c", url,
		"-n", teamName,
	}

	if username != "" && password != "" {
		args = append(args, "-u", username, "-p", password)
	}

	if insecure {
		args = append(args, "-k")
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		http.DefaultClient.Transport = tr
	}

	versionOut, err := f.versionCheck()
	if err != nil {
		return nil, err
	}
	f.logger.Debugf("Fly Version: %s\n", string(versionOut))

	loginOut, err := f.run(args...)
	if err != nil {
		return nil, err
	}

	syncOut, err := f.run("sync")
	if err != nil {
		return nil, err
	}

	return append(loginOut, syncOut...), nil
}

func (f command) Pipelines() ([]string, error) {
	psOut, err := f.run("pipelines", "--json")
	if err != nil {
		return nil, err
	}

	var ps []struct {
		Name string `json:"name"`
	}

	err = json.Unmarshal(psOut, &ps)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
	}

	return names, nil
}

func (f command) GetPipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"get-pipeline",
		"-p", pipelineName,
	)
}

func (f command) SetPipeline(
	pipelineName string,
	configFilepath string,
	varsFilepaths []string,
	vars map[string]interface{},
) ([]byte, error) {
	allArgs := []string{
		"set-pipeline",
		"-n",
		"-p", pipelineName,
		"-c", configFilepath,
	}

	for _, vf := range varsFilepaths {
		allArgs = append(allArgs, "-l", vf)
	}

	for key, value := range vars {
		payload, err := json.Marshal(value)

		if err != nil {
			return nil, err
		}

		allArgs = append(allArgs, "-y", fmt.Sprintf("%s=%s", key, payload))
	}

	return f.run(allArgs...)
}

func (f command) UnpausePipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"unpause-pipeline",
		"-p", pipelineName,
	)
}

func (f command) DestroyPipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"destroy-pipeline",
		"-n",
		"-p", pipelineName,
	)
}

func (f command) ExposePipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"expose-pipeline",
		"-p", pipelineName,
	)
}

func (f command) run(args ...string) ([]byte, error) {
	if f.target == "" {
		return nil, fmt.Errorf("target cannot be empty in command.run")
	}

	defaultArgs := []string{
		"-t", f.target,
	}
	allArgs := append(defaultArgs, args...)
	cmd := exec.Command(f.flyBinaryPath, allArgs...)

	outbuf := bytes.NewBuffer(nil)
	errbuf := bytes.NewBuffer(nil)

	cmd.Stdout = outbuf
	cmd.Stderr = errbuf

	f.logger.Debugf("Starting fly command: %v\n", allArgs)
	err := cmd.Start()
	if err != nil {
		// If the command was never started, there will be nothing in the buffers
		return nil, err
	}

	f.logger.Debugf("Waiting for fly command: %v\n", allArgs)
	err = cmd.Wait()
	if err != nil {
		if len(errbuf.Bytes()) > 0 {
			err = fmt.Errorf("%v - %s", err, string(errbuf.Bytes()))
		}
		return outbuf.Bytes(), err
	}

	return outbuf.Bytes(), nil
}

func (f command) versionCheck() ([]byte, error) {
	if f.target == "" {
		return nil, fmt.Errorf("target cannot be empty in command.run")
	}

	args := []string{
		"--version",
	}
	cmd := exec.Command(f.flyBinaryPath, args...)

	outbuf := bytes.NewBuffer(nil)
	errbuf := bytes.NewBuffer(nil)

	cmd.Stdout = outbuf
	cmd.Stderr = errbuf

	f.logger.Debugf("Starting fly command: %v\n", args)
	err := cmd.Start()
	if err != nil {
		// If the command was never started, there will be nothing in the buffers
		return nil, err
	}

	f.logger.Debugf("Waiting for fly command: %v\n", args)
	err = cmd.Wait()
	if err != nil {
		if len(errbuf.Bytes()) > 0 {
			err = fmt.Errorf("%v - %s", err, string(errbuf.Bytes()))
		}
		return outbuf.Bytes(), err
	}

	versionUrl := f.target + "/api/v1/info"

	res, err := http.Get(versionUrl)
	if err != nil {
		return nil, err
	}

	var targetVersion struct {
		Version        string `json:"version"`
		Worker_version string `json:"worker_version"`
	}

	err = json.NewDecoder(res.Body).Decode(&targetVersion)
	if err != nil {
		return nil, err
	}

	bakedVersion := string(outbuf.Bytes())

	if targetVersion.Version == bakedVersion{
		return []byte(targetVersion.Version), nil
	}
	if strings.Split(targetVersion.Version, ".")[0] == strings.Split(bakedVersion, ".")[0] {
		return []byte(targetVersion.Version), nil
	}

	file, err := os.OpenFile(f.flyBinaryPath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
	    return []byte("Unable to open fly binary"), err
	}
	defer file.Close()

	binaryUrl := f.target + "/api/v1/cli?arch=amd64&platform=linux"

	binaryResp, err := http.Get(binaryUrl)
	if err != nil {
		return []byte("Unable to download new fly binary"), err
	}
	defer binaryResp.Body.Close()
	_, err = io.Copy(file, binaryResp.Body)
	if err != nil {
		return []byte("Unable to update fly binary"), err
	}
	return []byte(targetVersion.Version), nil
}
