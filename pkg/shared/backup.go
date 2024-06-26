//go:build !ci

package shared

/*
#cgo CFLAGS: -I../../modules/aerospike-tools-backup/include
#cgo darwin CFLAGS: -I../../modules/aerospike-tools-backup/modules/c-client/target/Darwin-x86_64/include
#cgo darwin CFLAGS: -I../../modules/aerospike-tools-backup/modules/secret-agent-client/target/Darwin-x86_64/include
#cgo linux CFLAGS: -I../../modules/aerospike-tools-backup/modules/c-client/target/Linux-x86_64/include
#cgo linux CFLAGS: -I../../modules/aerospike-tools-backup/modules/secret-agent-client/target/Linux-x86_64/include
#cgo LDFLAGS: -L${SRCDIR}/../../lib -lasbackup

#include <stddef.h>
#include <stdio.h>
#include <stdint.h>

#include <backup.h>
#include <utils.h>
*/
import "C"
import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/model"
	"github.com/aws/smithy-go/ptr"
)

// BackupShared implements the Backup interface.
type BackupShared struct {
	sync.Mutex
}

var _ Backup = (*BackupShared)(nil)

// NewBackup returns a new BackupShared instance.
func NewBackup() *BackupShared {
	return &BackupShared{}
}

// BackupRun calls the backup_run function from the asbackup shared library.
//
//nolint:funlen,gocritic
func (b *BackupShared) BackupRun(backupRoutine *model.BackupRoutine, backupPolicy *model.BackupPolicy,
	cluster *model.AerospikeCluster, storage *model.Storage, secretAgent *model.SecretAgent,
	opts BackupOptions, namespace *string, path *string) (*BackupStat, error) {
	// lock to restrict parallel execution (shared library limitation)
	b.Lock()
	defer b.Unlock()

	backupConfig := C.backup_config_t{}
	C.backup_config_init(&backupConfig)
	defer C.backup_config_destroy(&backupConfig)

	setCString(&backupConfig.host, cluster.SeedNodesAsString())
	setCBool(&backupConfig.use_services_alternate, cluster.UseServicesAlternate)

	setCString(&backupConfig.user, cluster.GetUser())
	setCString(&backupConfig.password, cluster.GetPassword())
	setCString(&backupConfig.auth_mode, cluster.GetAuthMode())

	parseSetList(&backupConfig.set_list, &backupRoutine.SetList)
	if backupRoutine.BinList != nil {
		setCString(&backupConfig.bin_list, ptr.String(strings.Join(backupRoutine.BinList, ",")))
	}
	if backupRoutine.PreferRacks != nil {
		setCString(&backupConfig.prefer_racks, joinInts(backupRoutine.PreferRacks))
	}
	setCUint(&backupConfig.socket_timeout, backupPolicy.SocketTimeout)
	setCUint(&backupConfig.total_timeout, backupPolicy.TotalTimeout)

	// namespace list configuration
	nsCharArray := C.CString(*namespace)
	C.strcpy((*C.char)(unsafe.Pointer(&backupConfig.ns)), nsCharArray)

	setCInt(&backupConfig.parallel, backupPolicy.Parallel)

	setCBool(&backupConfig.remove_files, ptr.Bool(true))
	setCBool(&backupConfig.no_bins, backupPolicy.NoBins)
	setCBool(&backupConfig.no_records, backupPolicy.NoRecords)
	setCBool(&backupConfig.no_indexes, backupPolicy.NoIndexes)
	setCBool(&backupConfig.no_udfs, backupPolicy.NoUdfs)

	setCUlong(&backupConfig.bandwidth, backupPolicy.Bandwidth)
	setCUlong(&backupConfig.max_records, backupPolicy.MaxRecords)
	setCUint(&backupConfig.records_per_second, backupPolicy.RecordsPerSecond)
	setCUlong(&backupConfig.file_limit, backupPolicy.FileLimit)
	setCString(&backupConfig.partition_list, backupRoutine.PartitionList)

	// S3 configuration
	setCString(&backupConfig.s3_endpoint_override, storage.S3EndpointOverride)
	setCString(&backupConfig.s3_region, storage.S3Region)
	setCString(&backupConfig.s3_profile, storage.S3Profile)
	setS3LogLevel(&backupConfig.s3_log_level, storage.S3LogLevel)

	// Secret Agent configuration
	backupSecretAgent(&backupConfig, secretAgent)

	// TLS configuration
	setTLSOptions(&backupConfig.tls_name, &backupConfig.tls, cluster.TLS)

	// Encryption configuration
	err := configureEncryption(&backupConfig.encrypt_mode, &backupConfig.pkey, backupPolicy.EncryptionPolicy)
	if err != nil {
		return nil, fmt.Errorf("error configuring encryption: %w", err)
	}

	// Compression configuration
	configureCompression(&backupConfig.compress_mode, &backupConfig.compression_level,
		backupPolicy.CompressionPolicy)

	setCString(&backupConfig.directory, path)
	setCLong(&backupConfig.mod_after, opts.ModAfter)
	setCLong(&backupConfig.mod_before, opts.ModBefore)

	backupStatus := C.backup_run(&backupConfig)

	if unsafe.Pointer(backupStatus) == C.RUN_BACKUP_FAILURE {
		return nil, fmt.Errorf("backup failure")
	}

	result := &BackupStat{}
	if unsafe.Pointer(backupStatus) == C.RUN_BACKUP_SUCCESS {
		return result, nil
	}

	setStatistics(result, backupStatus)

	C.backup_status_destroy(backupStatus)
	C.cf_free(unsafe.Pointer(backupStatus))

	return result, nil
}

func setStatistics(result *BackupStat, status *C.backup_status_t) {
	result.RecordCount = uint64(status.rec_count_total)
	result.ByteCount = uint64(status.byte_count_total)
	result.FileCount = uint64(status.file_count)
	result.IndexCount = uint64(status.index_count)
	result.UDFCount = uint64(status.udf_count)
}

func backupSecretAgent(config *C.backup_config_t, secretsAgent *model.SecretAgent) {
	if secretsAgent != nil {
		config.secret_cfg.addr = C.CString(secretsAgent.Address)
		config.secret_cfg.port = C.CString(secretsAgent.Port)
		config.secret_cfg.timeout = C.int(secretsAgent.Timeout)
		config.secret_cfg.tls.ca_string = C.CString(secretsAgent.TLSCAString)
		setCBool(&config.secret_cfg.tls.enabled, &secretsAgent.TLSEnabled)
	}
}

// parseSetList parses the configured set list for backup
func parseSetList(setVector *C.as_vector, setList *[]string) {
	if setList != nil && len(*setList) > 0 {
		concatenatedSetList := strings.Join(*setList, ",")
		C.parse_set_list(setVector, C.CString(concatenatedSetList))
	}
}

func joinInts(nums []int) *string {
	strNums := make([]string, len(nums))
	for i, num := range nums {
		strNums[i] = strconv.Itoa(num)
	}
	join := strings.Join(strNums, ",")
	return &join
}
