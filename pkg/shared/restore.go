//go:build !ci

package shared

/*
#cgo CFLAGS: -I../../modules/aerospike-tools-backup/include
#cgo darwin CFLAGS: -I../../modules/aerospike-tools-backup/modules/c-client/target/Darwin-x86_64/include
#cgo darwin CFLAGS: -I../../modules/aerospike-tools-backup/modules/secret-agent-client/target/Darwin-x86_64/include
#cgo linux CFLAGS: -I../../modules/aerospike-tools-backup/modules/c-client/target/Linux-x86_64/include
#cgo linux CFLAGS: -I../../modules/aerospike-tools-backup/modules/secret-agent-client/target/Linux-x86_64/include
#cgo LDFLAGS: -L${SRCDIR}/../../lib -lasrestore

#include <stddef.h>
#include <stdio.h>
#include <stdint.h>

#include <restore.h>
*/
import "C"
import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"log/slog"

	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/model"
)

// RestoreShared implements the Restore interface.
type RestoreShared struct {
	sync.Mutex
}

var _ Restore = (*RestoreShared)(nil)

// NewRestore returns a new RestoreShared instance.
func NewRestore() *RestoreShared {
	return &RestoreShared{}
}

// RestoreRun calls the restore_run function from the asrestore shared library.
//
//nolint:funlen,gocritic
func (r *RestoreShared) RestoreRun(restoreRequest *model.RestoreRequestInternal) (*model.RestoreResult, error) {
	// lock to restrict parallel execution (shared library limitation)
	r.Lock()
	defer r.Unlock()

	slog.Debug("Starting restore operation")

	restoreConfig := C.restore_config_t{}
	C.restore_config_init(&restoreConfig)
	defer C.restore_config_destroy(&restoreConfig)

	setCString(&restoreConfig.host, restoreRequest.DestinationCuster.SeedNodesAsString())
	setCBool(&restoreConfig.use_services_alternate, restoreRequest.DestinationCuster.UseServicesAlternate)

	setCString(&restoreConfig.user, restoreRequest.DestinationCuster.GetUser())
	setCString(&restoreConfig.password, restoreRequest.DestinationCuster.GetPassword())
	setCString(&restoreConfig.auth_mode, restoreRequest.DestinationCuster.GetAuthMode())

	setCUint(&restoreConfig.parallel, restoreRequest.Policy.Parallel)
	setCBool(&restoreConfig.no_records, restoreRequest.Policy.NoRecords)
	setCBool(&restoreConfig.no_indexes, restoreRequest.Policy.NoIndexes)
	setCBool(&restoreConfig.no_udfs, restoreRequest.Policy.NoUdfs)

	setCUint(&restoreConfig.timeout, restoreRequest.Policy.Timeout)

	setCBool(&restoreConfig.disable_batch_writes, restoreRequest.Policy.DisableBatchWrites)
	setCUint(&restoreConfig.max_async_batches, restoreRequest.Policy.MaxAsyncBatches)
	setCUint(&restoreConfig.batch_size, restoreRequest.Policy.BatchSize)

	if restoreRequest.Policy.Namespace != nil {
		nsList := *restoreRequest.Policy.Namespace.Source + "," + *restoreRequest.Policy.Namespace.Destination
		setCString(&restoreConfig.ns_list, &nsList)
	}
	if len(restoreRequest.Policy.SetList) > 0 {
		setList := strings.Join(restoreRequest.Policy.SetList, ",")
		setCString(&restoreConfig.set_list, &setList)
	}
	if len(restoreRequest.Policy.BinList) > 0 {
		binList := strings.Join(restoreRequest.Policy.BinList, ",")
		setCString(&restoreConfig.bin_list, &binList)
	}

	// S3 configuration
	setCString(&restoreConfig.s3_endpoint_override, restoreRequest.SourceStorage.S3EndpointOverride)
	setCString(&restoreConfig.s3_region, restoreRequest.SourceStorage.S3Region)
	setCString(&restoreConfig.s3_profile, restoreRequest.SourceStorage.S3Profile)
	setS3LogLevel(&restoreConfig.s3_log_level, restoreRequest.SourceStorage.S3LogLevel)

	// Secret Agent configuration
	restoreSecretAgent(&restoreConfig, restoreRequest.SecretAgent)

	// TLS configuration
	setTLSOptions(&restoreConfig.tls_name, &restoreConfig.tls, restoreRequest.DestinationCuster.TLS)

	// Encryption configuration
	err := configureEncryption(&restoreConfig.encrypt_mode, &restoreConfig.pkey, restoreRequest.Policy.EncryptionPolicy)
	if err != nil {
		return nil, fmt.Errorf("error configuring encryption: %w", err)
	}

	// Compression configuration
	configureCompression(&restoreConfig.compress_mode, nil, restoreRequest.Policy.CompressionPolicy)

	// restore source configuration
	setCString(&restoreConfig.directory, restoreRequest.Dir)

	setCBool(&restoreConfig.replace, restoreRequest.Policy.Replace)
	setCBool(&restoreConfig.unique, restoreRequest.Policy.Unique)
	setCBool(&restoreConfig.no_generation, restoreRequest.Policy.NoGeneration)

	setCUlong(&restoreConfig.bandwidth, restoreRequest.Policy.Bandwidth)
	setCUint(&restoreConfig.tps, restoreRequest.Policy.Tps)

	restoreStatus := C.restore_run(&restoreConfig)

	if unsafe.Pointer(restoreStatus) == C.RUN_RESTORE_FAILURE {
		return nil, fmt.Errorf("failed restore operation")
	}

	result := getRestoreResult(restoreStatus)

	C.restore_status_destroy(restoreStatus)
	C.cf_free(unsafe.Pointer(restoreStatus))

	return result, nil
}

func getRestoreResult(status *C.restore_status_t) *model.RestoreResult {
	return &model.RestoreResult{
		TotalRecords:    uint64(status.total_records),
		TotalBytes:      uint64(status.total_bytes),
		ExpiredRecords:  uint64(status.expired_records),
		SkippedRecords:  uint64(status.skipped_records),
		IgnoredRecords:  uint64(status.ignored_records),
		InsertedRecords: uint64(status.inserted_records),
		ExistedRecords:  uint64(status.existed_records),
		FresherRecords:  uint64(status.fresher_records),
		IndexCount:      uint64(status.index_count),
		UDFCount:        uint64(status.udf_count),
	}
}

func restoreSecretAgent(config *C.restore_config_t, secretsAgent *model.SecretAgent) {
	if secretsAgent != nil {
		config.secret_cfg.addr = C.CString(secretsAgent.Address)
		config.secret_cfg.port = C.CString(secretsAgent.Port)
		config.secret_cfg.timeout = C.int(secretsAgent.Timeout)
		config.secret_cfg.tls.ca_string = C.CString(secretsAgent.TLSCAString)
		setCBool(&config.secret_cfg.tls.enabled, &secretsAgent.TLSEnabled)
	}
}
