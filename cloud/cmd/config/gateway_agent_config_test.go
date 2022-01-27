package config

import (
	"encoding/json"
	"log"
	"testing"
)

func TestRemoteGateway(t *testing.T) {
	receivedData := []string{
		`{
			"RemoteGateways": [
				{
					"RemoteGatewayIP": "1.0.0.1",
					"RemoteGatewayPort": 800
				}
			],
			"LogLevel": "3"
		}`,
		`{
			"RemoteGateways": [
				{
					"RemoteGatewayIP": "2.0.0.1",
					"RemoteGatewayPort": 800
				},
				{
					"RemoteGatewayIP": "172.15.2.3",
					"RemoteGatewayPort": 6081
				}
			],
			"LogLevel": "1"
		}`,
		`{
			"RemoteGateways": [
				{
					"RemoteGatewayIP": "1.0.0.1",
					"RemoteGatewayPort": 800
				}
			],
			"GenevePort": 100,
			"LogLevel": "3"
		}`,
		`{
			"RemoteGateways": [
				{
					"RemoteGatewayIP": "1.0.0.1",
					"RemoteGatewayPort": 800
				}
			],
			"GenevePort": 100,
			"LocalDividerIP": "127.0.0.1",
			"LogLevel": "3"
		}`,
	}

	expectedConfigs := []GatewayAgentConfiguration{
		{
			RemoteGateways: []RemoteGateway{
				{RemoteGatewayIP: "1.0.0.1", RemoteGatewayPort: 800},
			},
			LogLevel: "3",
		},
		{
			RemoteGateways: []RemoteGateway{
				{RemoteGatewayIP: "2.0.0.1", RemoteGatewayPort: 800},
				{RemoteGatewayIP: "172.15.2.3", RemoteGatewayPort: 6081},
			},
			LogLevel: "1",
		},
		{
			RemoteGateways: []RemoteGateway{
				{RemoteGatewayIP: "1.0.0.1", RemoteGatewayPort: 800},
			},
			GenevePort: 100,
			LogLevel:   "3",
		},
		{
			RemoteGateways: []RemoteGateway{
				{RemoteGatewayIP: "1.0.0.1", RemoteGatewayPort: 800},
			},
			GenevePort:     100,
			LocalDividerIP: "127.0.0.1",
			LogLevel:       "3",
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
	if len(received.RemoteGateways) != len(expected.RemoteGateways) {
		t.Errorf("Received %d gateway config, expected %d gateway config", len(received.RemoteGateways), len(expected.RemoteGateways))
	}
	for i, gateway := range received.RemoteGateways {
		if gateway.RemoteGatewayIP != expected.RemoteGateways[i].RemoteGatewayIP {
			t.Errorf("Received %s gateway ip, expected %s gateway ip", gateway.RemoteGatewayIP, expected.RemoteGateways[i].RemoteGatewayIP)
		}
		if gateway.RemoteGatewayPort != expected.RemoteGateways[i].RemoteGatewayPort {
			t.Errorf("Received %d gateway port, expected %d gateway port", gateway.RemoteGatewayPort, expected.RemoteGateways[i].RemoteGatewayPort)
		}
	}
	if received.LocalDividerIP != expected.LocalDividerIP {
		t.Errorf("Received %s local divider ip, expected %s local divider ip", received.LocalDividerIP, expected.LocalDividerIP)
	}
	if received.GenevePort != expected.GenevePort {
		t.Errorf("Received %d geneve port, expected %d geneve port", received.GenevePort, expected.GenevePort)
	}
	if received.LogLevel != expected.LogLevel {
		t.Errorf("Received %s log level, expected %s local log level", received.LogLevel, expected.LogLevel)
	}
}
