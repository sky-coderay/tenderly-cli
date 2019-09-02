package truffle

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"path/filepath"
)

const (
	NewTruffleConfigFile = "truffle-config.js"
	OldTruffleConfigFile = "truffle.js"
)

type NetworkConfig struct {
	Host      string      `json:"host"`
	Port      int         `json:"port"`
	NetworkID interface{} `json:"network_id"`
}

type Compiler struct {
	Version  string            `json:"version"`
	Settings *CompilerSettings `json:"settings"`
}

type CompilerSettings struct {
	Optimizer  *Optimizer `json:"optimizer"`
	EvmVersion *string    `json:"evmVersion"`
}

type Optimizer struct {
	Enabled *bool `json:"enabled"`
	Runs    *int  `json:"runs"`
}

type Config struct {
	ProjectDirectory string                   `json:"project_directory"`
	BuildDirectory   string                   `json:"contracts_build_directory"`
	Networks         map[string]NetworkConfig `json:"networks"`
	Solc             map[string]Optimizer     `json:"solc"`
	Compilers        map[string]Compiler      `json:"compilers"`
}

func (c *Config) AbsoluteBuildDirectoryPath() string {
	if c.BuildDirectory == "" {
		c.BuildDirectory = filepath.Join(".", "build", "contracts")
	}

	switch c.BuildDirectory[0] {
	case '.':
		return filepath.Join(c.ProjectDirectory, c.BuildDirectory)
	default:
		return c.BuildDirectory
	}
}

func GetTruffleConfig(configName string, projectDir string) (*Config, error) {
	trufflePath := filepath.Join(projectDir, configName)
	divider := getDivider()

	logrus.Debugf("Trying truffle config path: %s", trufflePath)

	data, err := exec.Command("node", "-e", fmt.Sprintf(`
		var config = require('%s');

		console.log("%s" + JSON.stringify(config) + "%s");
	`, trufflePath, divider, divider)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cannot find %s, tried path: %s, error: %s", configName, trufflePath, err)
	}

	configString, err := extractConfigWithDivider(string(data), divider)
	if err != nil {
		logrus.Debugf("failed extracting config with divider: %s", err)
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	fmt.Println(string(data))
	fmt.Println(string(configString))

	var truffleConfig Config
	err = json.Unmarshal([]byte(configString), &truffleConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s", configName)
	}

	truffleConfig.ProjectDirectory = projectDir

	return &truffleConfig, nil
}

func getDivider() string {
	return fmt.Sprintf("======%s======", randSeq(10))
}
