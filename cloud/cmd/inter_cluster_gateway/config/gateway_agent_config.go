/*
Copyright 2022 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
)

type GatewayAgentConfiguration struct {
	RemoteGateways []RemoteGateway
	LocalDividerIP string
	GenevePort     int
	LogLevel       int
}

type RemoteGateway struct {
	RemoteGatewayIP   string
	RemoteGatewayPort int
}

func NewGatewayConfiguration(fileName string) (GatewayAgentConfiguration, error) {
	_, runningfile, _, ok := runtime.Caller(1)
	configuration := GatewayAgentConfiguration{}
	if !ok {
		return configuration, fmt.Errorf("failed to open the given config file %s", fileName)
	}
	filepath := path.Join(path.Dir(runningfile), fileName)
	fmt.Println(filepath)
	file, err := os.Open(filepath)
	if err != nil {
		return configuration, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&configuration)
	return configuration, err
}
