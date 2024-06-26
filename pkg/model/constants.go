package model

const (
	StateFileName                = "state.yaml"
	IncrementalBackupDirectory   = "incremental"
	FullBackupDirectory          = "backup"
	ConfigurationBackupDirectory = "configuration"
	DataDirectory                = "data"

	// max possible value https://aerospike.com/docs/server/reference/configuration#namespace__rack-id
	maxRack = 1000000
)
