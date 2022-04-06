package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
)

type GatewayAgentConfiguration struct {
	GatewayPort int
	GenevePort  int
	LogLevel    string
	GrpcPort    int
	GrpcTimeout int
}

func NewGatewayConfiguration(fileName string) (GatewayAgentConfiguration, error) {
	_, runningfile, _, ok := runtime.Caller(1)
	configuration := GatewayAgentConfiguration{}
	if !ok {
		return configuration, fmt.Errorf("failed to open the given config file %s", fileName)
	}
	filepath := path.Join(path.Dir(runningfile), fileName)
	file, err := os.Open(filepath)
	if err != nil {
		return configuration, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&configuration)
	return configuration, err
}
