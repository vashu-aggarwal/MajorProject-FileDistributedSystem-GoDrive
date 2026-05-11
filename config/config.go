package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// PipelineConfig holds settings for the optional Python pipeline microservice.
// Controlled entirely from config.yaml — set enabled: false to bypass the pipeline.
type PipelineConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ServiceURL string `yaml:"service_url"`
	AESKeyHex  string `yaml:"aes_key_hex"`
	Port       int    `yaml:"port"`
}

type MasterConfig struct {
	Host              string `yaml:"host"`
	HttpPort          int    `yaml:"http_port"`
	TcpPort           int    `yaml:"tcp_port"`
	WebSocketPort     int    `yaml:"ws_port"`
	HeartBeatInterval int    `yaml:"heartbeat_interval"`
	LockTimeout       int    `yaml:"lock_timeout"`
	ChunkSize         int    `yaml:"chunk_size"`
	ReplicationFactor int    `yaml:"replication_factor"`
	ReadQuorum        int    `yaml:"read_quorum"`
	WriteQuorum       int    `yaml:"write_quorum"`
}

type Node struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Config struct {
	Master      MasterConfig   `yaml:"master"`
	SlaveNodes  []Node         `yaml:"slaveDataNodes"`
	BackupNodes []Node         `yaml:"backupNodes"`
	Pipeline    PipelineConfig `yaml:"pipeline"`
}

var ReadConfig Config

func LoadConfig() {
	file, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Println("Couldn't load config", err)
	}
	err = yaml.Unmarshal(file, &ReadConfig)
	if err != nil {
		log.Println("Couldn't read config")
	}
	log.Println("Config Loaded Successfully!")
}
