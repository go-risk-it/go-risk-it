package config

type OtelConfig struct {
	Enabled      bool        `koanf:"enabled"`
	TraceQueries bool        `koanf:"trace_queries"`
	Batch        BatchConfig `koanf:"batch"`
}

type BatchConfig struct {
	MaxQueueSize       int `koanf:"max_queue_size"`
	MaxExportBatchSize int `koanf:"max_export_batch_size"`
}
