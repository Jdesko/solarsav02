package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const (
	Host           = "https://openapi.tuyaeu.com"
	ClientID       = "hskykhepdrxhmygcg5cd"
	Secret         = "c28b99733f7c4c6cb4fe3b3f7d8e283f"
	UpdateInterval = 2 * time.Second
	TokenExpiry    = 4 * time.Hour
	DataFile       = "data/energy.json"
)

var (
	Token       string
	tokenExpiry time.Time
)

type DeviceStatus struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

type DeviceResponse struct {
	Result  []DeviceStatus `json:"result"`
	Success bool           `json:"success"`
	T       int64          `json:"t"`
}

type EnergyData struct {
	PainelSolar    float64 `json:"painel_solar"`
	ImportacaoRede float64 `json:"importacao_rede"`
	CarregadorEV   float64 `json:"carregador_ev"`
	ConsumoCasa    float64 `json:"consumo_casa"`
	VoltagemMedia  float64 `json:"voltagem_media"`
	Timestamp      int64   `json:"timestamp"`
}

var devices = map[string]map[string]string{
	"bfe2a8f17ebd1d39d1hctm": {
		"cur_power2":   "importacao_rede",
		"cur_voltage2": "voltagem",
	},
	"bfcaa7d812363ec7eeimjs": {
		"cur_power1":   "carregador_ev",
		"cur_power2":   "painel_solar",
		"cur_voltage1": "voltagem",
	},
}

func main() {
	// Criar diretório data se não existir
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}

	if err := refreshToken(); err != nil {
		log.Fatal("Initial token refresh failed:", err)
	}

	go tokenRefresher()
	updateDataPeriodically()
}

func updateDataPeriodically() {
	ticker := time.NewTicker(UpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := updateEnergyData(); err != nil {
			log.Printf("Error updating data: %v", err)
		}
	}
}

func updateEnergyData() error {
	data, err := fetchEnergyData()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(DataFile, jsonData, 0644)
}

func fetchEnergyData() (*EnergyData, error) {
	data := &EnergyData{
		Timestamp: time.Now().Unix(),
	}

	var voltageSum float64
	voltageCount := 0

	for deviceID, mappings := range devices {
		status, err := getDeviceStatus(deviceID)
		if err != nil {
			return nil, fmt.Errorf("device %s error: %v", deviceID, err)
		}

		for _, item := range status.Result {
			if mapping, exists := mappings[item.Code]; exists {
				value := convertToFloat(item.Value) * 0.1

				switch mapping {
				case "painel_solar":
					data.PainelSolar = value
				case "importacao_rede":
					data.ImportacaoRede = value
				case "carregador_ev":
					data.CarregadorEV = value
				case "voltagem":
					voltageSum += value
					voltageCount++
				}
			}
		}
	}

	if voltageCount > 0 {
		data.VoltagemMedia = voltageSum / float64(voltageCount)
	}

	data.ConsumoCasa = data.PainelSolar + data.ImportacaoRede - data.CarregadorEV
	return data, nil
}

// (Mantenha todas as outras funções do seu código original)
// getDeviceStatus, refreshToken, tokenRefresher, buildHeader,
// buildSign, sha256Hash, hmacSha256, getUrlStr, getHeaderStr,
// convertToFloat - todas idênticas ao seu código original
