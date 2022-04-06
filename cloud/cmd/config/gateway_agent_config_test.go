package config

import (
	"encoding/json"
	"log"
	"testing"
)

func TestRemoteGateway(t *testing.T) {
	receivedData := []string{
		`{
			"LogLevel": "3"
		}`,
		`{
			"GatewayPort": 2615,
			"LogLevel": "1"
		}`,
		`{
			"GatewayPort": 2616,
			"GenevePort": 6081,
			"LogLevel": "3"
		}`,
		`{
			"GatewayPort": 2616,
			"GenevePort": 6082,
			"LogLevel": "3"
		}`,
		`{
			"GrpcPort": 8090,
			"GatewayPort": 2616,
			"GenevePort": 6082,
			"LogLevel": "3"
		}`,
		`{
			"GrpcPort": 8090,
			"GatewayPort": 2616,
			"GenevePort": 6082,
			"LogLevel": "3"
		}`,
		`{
			"GatewayPort": 2616,
			"GenevePort": 6082,
			"LogLevel": "3",
			"GrpcPort": 8091,
			"GrpcTimeout": 10
		}`,
	}

	expectedConfigs := []GatewayAgentConfiguration{
		{
			LogLevel: "3",
		},
		{
			GatewayPort: 2615,
			LogLevel:    "1",
		},
		{
			GatewayPort: 2616,
			GenevePort:  6081,
			LogLevel:    "3",
		},
		{
			GatewayPort: 2616,
			GenevePort:  6082,
			LogLevel:    "3",
		},
		{
			GatewayPort: 2616,
			GenevePort:  6082,
			LogLevel:    "3",
			GrpcPort:    8090,
		},
		{
			GatewayPort: 2616,
			GenevePort:  6082,
			LogLevel:    "3",
			GrpcPort:    8090,
		},
		{
			GatewayPort: 2616,
			GenevePort:  6082,
			LogLevel:    "3",
			GrpcPort:    8091,
			GrpcTimeout: 10,
		},
	}

	for i, data := range receivedData {
		received := GatewayAgentConfiguration{}
		err := json.Unmarshal([]byte(data), &received)
		if err != nil {
			log.Fatal(err)
		}
		assertEquals(received, expectedConfigs[i], t)
	}
}

func assertEquals(received, expected GatewayAgentConfiguration, t *testing.T) {
	if received.GenevePort != expected.GenevePort {
		t.Errorf("Received %d geneve port, expected %d geneve port", received.GenevePort, expected.GenevePort)
	}
	if received.LogLevel != expected.LogLevel {
		t.Errorf("Received %s log level, expected %s local log level", received.LogLevel, expected.LogLevel)
	}
	if received.GatewayPort != expected.GatewayPort {
		t.Errorf("Received %d gateway port, expected %d gateway port", received.GatewayPort, expected.GatewayPort)
	}
	if received.GrpcPort != expected.GrpcPort {
		t.Errorf("Received %d grpc port, expected %d grpc port", received.GrpcPort, expected.GrpcPort)
	}
	if received.GrpcTimeout != expected.GrpcTimeout {
		t.Errorf("Received %d grpc timeout, expected %d grpc timeout", received.GrpcTimeout, expected.GrpcTimeout)
	}
}
