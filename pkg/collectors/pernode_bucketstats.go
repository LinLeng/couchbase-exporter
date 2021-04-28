//  Copyright (c) 2019 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package collectors

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/couchbase/couchbase-exporter/pkg/log"
	"github.com/couchbase/couchbase-exporter/pkg/objects"
	"github.com/couchbase/couchbase-exporter/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	client = http.Client{}
)

const (
	subsystem = "pernodebucket"
)

var (
	AvgDiskUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "avg_disk_update_time",
		Help: "Average disk update time in microseconds as from disk_update histogram of timings",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	AvgDiskCommitTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "avg_disk_commit_time",
		Help: "Average disk commit time in seconds as from disk_update histogram of timings",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	AvgBgWaitTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "avg_bg_wait_seconds",
		Help: " ",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	AvgActiveTimestampDrift = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "avg_active_timestamp_drift",
		Help: "Average drift (in seconds) between mutation timestamps and the local time for active vBuckets. (measured from ep_active_hlc_drift and ep_active_hlc_drift_count)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	AvgReplicaTimestampDrift = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "avg_replica_timestamp_drift",
		Help: "Average drift (in seconds) between mutation timestamps and the local time for replica vBuckets. (measured from ep_replica_hlc_drift and ep_replica_hlc_drift_count)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CouchTotalDiskSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_total_disk_size",
		Help: "The total size on disk of all data and view files for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchDocsFragmentation = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_docs_fragmentation",
		Help: "How much fragmented data there is to be compacted compared to real data for the data files in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchViewsFragmentation = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_views_fragmentation",
		Help: "How much fragmented data there is to be compacted compared to real data for the view index files in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchDocsActualDiskSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_docs_actual_disk_size",
		Help: "The size of all data files for this bucket, including the data itself, meta data and temporary files",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchDocsDataSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_docs_data_size",
		Help: "The size of active data in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchDocsDiskSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_docs_disk_size",
		Help: "The size of all data files for this bucket, including the data itself, meta data and temporary files",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchSpatialDataSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_spatial_data_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchSpatialDiskSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_spatial_disk_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchSpatialOps = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_spatial_ops",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchViewsActualDiskSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_views_actual_disk_size",
		Help: "The size of all active items in all the indexes for this bucket on disk",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchViewsDataSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_views_data_size",
		Help: "The size of active data on for all the indexes in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchViewsDiskSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_views_disk_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	CouchViewsOps = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "couch_views_ops",
		Help: "All the view reads for all design documents including scatter gather",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	HitRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "hit_ratio",
		Help: "Hit ratio",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpCacheMissRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_cache_miss_rate",
		Help: "Percentage of reads per second to this bucket from disk as opposed to RAM",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	EpResidentItemsRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_resident_items_rate",
		Help: "Percentage of all items cached in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsIndexesCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_indexes_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	EpDcpViewsIndexesItemsRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_indexes_items_remaining",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	EpDcpViewsIndexesProducerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_indexes_producer_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
	EpDcpViewsIndexesTotalBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_indexes_total_backlog_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsIndexesItemsSent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_indexes_items_sent",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsIndexesTotalBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_indexes_total_bytes",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsIndexesBackoff = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_indexes_backoff",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	BgWaitCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "bg_wait_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	BgWaitTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "bg_wait_total",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	BytesRead = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "bytes_read",
		Help: "Bytes Read",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	BytesWritten = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "bytes_written",
		Help: "Bytes written",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CasBadVal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "cas_bad_val",
		Help: "Compare and Swap bad values",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CasHits = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "cas_hits",
		Help: "Number of operations with a CAS id per second for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CasMisses = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "cas_misses",
		Help: "Compare and Swap misses",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CmdGet = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "cmd_get",
		Help: "Number of reads (get operations) per second from this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CmdSet = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "cmd_set",
		Help: "Number of writes (set operations) per second to this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CurrConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "curr_connections",
		Help: "Number of connections to this server including connections from external client SDKs, proxies, DCP requests and internal statistic gathering",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CurrItems = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "curr_items",
		Help: "Number of items in active vBuckets in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CurrItemsTot = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "curr_items_tot",
		Help: "Total number of items in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DecrHits = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "decr_hits",
		Help: "Decrement hits",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DecrMisses = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "decr_misses",
		Help: "Decrement misses",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DeleteHits = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "delete_hits",
		Help: "Number of delete operations per second for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DeleteMisses = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "delete_misses",
		Help: "Number of delete operations per second for data that this bucket does not contain. (measured from delete_misses)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DiskCommitCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "disk_commit_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DiskCommitTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "disk_commit_total",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DiskUpdateCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "disk_update_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DiskUpdateTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "disk_update_total",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	DiskWriteQueue = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "disk_write_queue",
		Help: "Number of items waiting to be written to disk in this bucket. (measured from ep_queue_size+ep_flusher_todo)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpActiveAheadExceptions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_active_ahead_exceptions",
		Help: "Total number of ahead exceptions (when timestamp drift between mutations and local time has exceeded 5000000 μs) per second for all active vBuckets.",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpActiveHlcDrift = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_active_hlc_drift",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpActiveHlcDriftCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_active_hlc_drift_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpBgFetched = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_bg_fetched",
		Help: "Number of reads per second from disk for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpClockCasDriftTheresholExceeded = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_clock_cas_drift_threshold_exceeded",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDataReadFailed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_data_read_failed",
		Help: "Number of disk read failures. (measured from ep_data_read_failed)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDataWriteFailed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_data_write_failed",
		Help: "Number of disk write failures. (measured from ep_data_write_failed)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcp2iBackoff = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_2i_backoff",
		Help: "Number of backoffs for indexes DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcp2iCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_2i_count",
		Help: "Number of indexes DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcp2iItemsRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_2i_items_remaining",
		Help: "Number of indexes items remaining to be sent",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcp2iItemsSent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_2i_items_sent",
		Help: "Number of indexes items sent",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcp2iProducerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_2i_producers",
		Help: "Number of indexes producers",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcp2iTotalBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_2i_total_backlog_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcp2iTotalBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_2i_total_bytes",
		Help: "Number of bytes per second being sent for indexes DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpCbasBackoff = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_cbas_backoff",
		Help: "Number of backoffs per second for analytics DCP connections (measured from ep_dcp_cbas_backoff)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpCbasCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_cbas_count",
		Help: "Number of internal analytics DCP connections in this bucket (measured from ep_dcp_cbas_count)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpCbasItemsRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_cbas_items_remaining",
		Help: "Number of items remaining to be sent to consumer in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpCbasItemsSent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_cbas_items_sent",
		Help: "Number of items per second being sent for a producer for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpCbasProducerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_cbas_producer_count",
		Help: "Number of analytics senders for this bucket (measured from ep_dcp_cbas_producer_count)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpCbasTotalBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_cbas_total_backlog_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpCbasTotalBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_total_bytes",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpFtsBackoff = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_fts_backoff",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpFtsCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_fts_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpFtsItemsRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_fts_items_remaining",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpFtsItemsSent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_fts_items_sent",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpFtsProducerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_fts_producer_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpFtsTotalBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_fts_backlog_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpFtsTotalBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_fts_total_bytes",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpOtherBackoff = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_other_backoff",
		Help: "Number of backoffs for other DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpOtherCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_other_count",
		Help: "Number of other DCP connections in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpOtherItemsRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_other_items_remaining",
		Help: "Number of items remaining to be sent to consumer in this bucket (measured from ep_dcp_other_items_remaining)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpOtherItemsSent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_other_items_sent",
		Help: "Number of items per second being sent for a producer for this bucket (measured from ep_dcp_other_items_sent)",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpOtherProducerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_other_producer_count",
		Help: "Number of other senders for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpOtherTotalBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_other_total_backlog_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpOtherTotalBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_other_total_bytes",
		Help: "Number of bytes per second being sent for other DCP connections for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpReplicaBackoff = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_replica_backoff",
		Help: "Number of backoffs for replication DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpReplicaCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_replica_count",
		Help: "Number of internal replication DCP connections in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpReplicaItemsRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_replica_items_remaining",
		Help: "Number of items remaining to be sent to consumer in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpReplicaItemsSent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_replica_items_sent",
		Help: "Number of items per second being sent for a producer for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpReplicaProducerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_replica_producer_count",
		Help: "Number of replication senders for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpReplicaTotalBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_replica_total_backlog_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpReplicaTotalBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_replica_total_bytes",
		Help: "Number of bytes per second being sent for replication DCP connections for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsBackoff = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_backoff",
		Help: "Number of backoffs for views DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_count",
		Help: "Number of views DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsItemsRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_items_remaining",
		Help: "Number of views items remaining to be sent",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsItemsSent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_items_sent",
		Help: "Number of views items sent",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsProducerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_producer_count",
		Help: "Number of views producers",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsTotalBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_total_backlog_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpViewsTotalBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_views_total_bytes",
		Help: "Number bytes per second being sent for views DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpXdcrBackoff = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_xdcr_backoff",
		Help: "Number of backoffs for XDCR DCP connections",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpXdcrCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_xdcr_count",
		Help: "Number of internal XDCR DCP connections in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpXdcrItemsRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_xdcr_items_remaining",
		Help: "Number of items remaining to be sent to consumer in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpXdcrItemsSent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_xdcr_items_sent",
		Help: "Number of items per second being sent for a producer for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpXdcrProducerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_xdcr_producer_count",
		Help: "Number of XDCR senders for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpXdcrTotalBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_xdcr_total_backlog_size",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDcpXdcrTotalBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_dcp_xdcr_total_bytes",
		Help: "Number of bytes per second being sent for XDCR DCP connections for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDiskqueueDrain = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_diskqueue_drain",
		Help: "Total number of items per second being written to disk in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDiskqueueFill = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_diskqueue_fill",
		Help: "Total number of items per second being put on the disk queue in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpDiskqueueItems = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_diskqueue_items",
		Help: "Total number of items waiting to be written to disk in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpFlusherTodo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_flusher_todo",
		Help: "Number of items currently being written",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpItemCommitFailed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_item_commit_failed",
		Help: "Number of times a transaction failed to commit due to storage errors",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpKvSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_kv_size",
		Help: "Total amount of user data cached in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpMaxSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_max_size",
		Help: "The maximum amount of memory this bucket can use",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpMemHighWat = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_mem_high_wat",
		Help: "High water mark for auto-evictions",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpMemLowWat = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_mem_low_wat",
		Help: "Low water mark for auto-evictions",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpMetaDataMemory = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_meta_data_memory",
		Help: "Total amount of item metadata consuming RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpNumNonResident = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_num_non_resident",
		Help: "Number of non-resident items",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpNumOpsDelMeta = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_num_ops_del_meta",
		Help: "Number of delete operations per second for this bucket as the target for XDCR",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpNumOpsDelRetMeta = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_num_ops_del_ret_meta",
		Help: "Number of delRetMeta operations per second for this bucket as the target for XDCR",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpNumOpsGetMeta = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_num_ops_get_meta",
		Help: "Number of metadata read operations per second for this bucket as the target for XDCR",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpNumOpsSetMeta = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_num_ops_set_meta",
		Help: "Number of set operations per second for this bucket as the target for XDCR",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpNumOpsSetRetMeta = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_num_ops_set_ret_meta",
		Help: "Number of setRetMeta operations per second for this bucket as the target for XDCR",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpNumValueEjects = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_num_value_ejects",
		Help: "Total number of items per second being ejected to disk in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpOomErrors = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_oom_errors",
		Help: "Number of times unrecoverable OOMs happened while processing operations",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpOpsCreate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_ops_create",
		Help: "Total number of new items being inserted into this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpOpsUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_ops_update",
		Help: "Number of items updated on disk per second for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpOverhead = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_overhead",
		Help: "Extra memory used by transient data like persistence queues or checkpoints",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpQueueSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_queue_size",
		Help: "Number of items queued for storage",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpReplicaAheadExceptions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_replica_ahead_exceptions",
		Help: "Percentage of all items cached in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpReplicaHlcDrift = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_replica_hlc_drift",
		Help: "The sum of the total Absolute Drift, which is the accumulated drift observed by the vBucket. Drift is always accumulated as an absolute value.",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpReplicaHlcDriftCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_replica_hlc_drift_count",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpTmpOomErrors = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_tmp_oom_errors",
		Help: "Number of back-offs sent per second to client SDKs due to OOM situations from this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	EpVbTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ep_vb_total",
		Help: "Total number of vBuckets for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	Evictions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "evictions",
		Help: "Number of evictions",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	GetHits = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "get_hits",
		Help: "Number of get hits",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	GetMisses = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "get_misses",
		Help: "Number of get misses",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	IncrHits = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "incr_hits",
		Help: "Number of increment hits",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	IncrMisses = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "incr_misses",
		Help: "Number of increment misses",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	MemUsed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "mem_used",
		Help: "Amount of memory used",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	Misses = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "misses",
		Help: "Number of misses",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	Ops = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "ops",
		Help: "Total amount of operations per second to this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	// lol Timestamp

	VbActiveEject = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_eject",
		Help: "Number of items per second being ejected to disk from active vBuckets in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveItmMemory = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_itm_memory",
		Help: "Amount of active user data cached in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveMetaDataMemory = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_meta_data_memory",
		Help: "Amount of active item metadata consuming RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveNum = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_num",
		Help: "Number of vBuckets in the active state for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveNumNonresident = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_num_non_resident",
		Help: "Number of non resident vBuckets in the active state for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveOpsCreate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_ops_create",
		Help: "New items per second being inserted into active vBuckets in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveOpsUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_ops_update",
		Help: "Number of items updated on active vBucket per second for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveQueueAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_queue_age",
		Help: "Sum of disk queue item age in milliseconds",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveQueueDrain = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_queue_drain",
		Help: "Number of active items per second being written to disk in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveQueueFill = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_queue_fill",
		Help: "Number of active items per second being put on the active item disk queue in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveQueueSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_queue_size",
		Help: "Number of active items waiting to be written to disk in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveQueueItems = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_queue_items",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingCurrItems = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_curr_items",
		Help: "Number of items in pending vBuckets in this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingEject = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_eject",
		Help: "Number of items per second being ejected to disk from pending vBuckets in this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingItmMemory = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_itm_memory",
		Help: "Amount of pending user data cached in RAM in this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingMetaDataMemory = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_meta_data_memory",
		Help: "Amount of pending item metadata consuming RAM in this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingNum = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_num",
		Help: "Number of vBuckets in the pending state for this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingNumNonResident = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_num_non_resident",
		Help: "Number of non resident vBuckets in the pending state for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingOpsCreate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_ops_create",
		Help: "New items per second being instead into pending vBuckets in this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingOpsUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_ops_update",
		Help: "Number of items updated on pending vBucket per second for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingQueueAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_queue_age",
		Help: "Sum of disk pending queue item age in milliseconds",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingQueueDrain = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_queue_drain",
		Help: "Number of pending items per second being written to disk in this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingQueueFill = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_queue_fill",
		Help: "Number of pending items per second being put on the pending item disk queue in this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingQueueSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_queue_size",
		Help: "Number of pending items waiting to be written to disk in this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaCurrItems = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_curr_items",
		Help: "Number of items in replica vBuckets in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaEject = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_eject",
		Help: "Number of items per second being ejected to disk from replica vBuckets in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaItmMemory = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_itm_memory",
		Help: "Amount of replica user data cached in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaMetaDataMemory = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_meta_data_memory",
		Help: "Amount of replica item metadata consuming in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaNum = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_num",
		Help: "Number of vBuckets in the replica state for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaNumNonResident = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_num_non_resident",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaOpsCreate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_ops_create",
		Help: "New items per second being inserted into replica vBuckets in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaOpsUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_ops_update",
		Help: "Number of items updated on replica vBucket per second for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaQueueAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_queue_age",
		Help: "Sum of disk replica queue item age in milliseconds",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaQueueDrain = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_queue_drain",
		Help: "Number of replica items per second being written to disk in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaQueueFill = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_queue_fill",
		Help: "Number of replica items per second being put on the replica item disk queue in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaQueueSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_queue_size",
		Help: "Number of replica items waiting to be written to disk in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbTotalQueueAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_total_queue_age",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbAvgActiveQueueAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_avg_active_queue_age",
		Help: "Sum of disk queue item age in milliseconds",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbAvgReplicaQueueAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_avg_replica_queue_age",
		Help: "Average age in seconds of replica items in the replica item queue for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbAvgPendingQueueAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_avg_pending_queue_age",
		Help: "Average age in seconds of pending items in the pending item queue for this bucket and should be transient during rebalancing",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbAvgTotalQueueAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_avg_total_queue_age",
		Help: "Average age in seconds of all items in the disk write queue for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbActiveResidentItemsRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_active_resident_items_ratio",
		Help: "Percentage of active items cached in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbReplicaResidentItemsRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_replica_resident_items_ratio",
		Help: "Percentage of active items cached in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	VbPendingResidentItemsRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "vb_pending_resident_items_ratio",
		Help: "Percentage of items in pending state vbuckets cached in RAM in this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	XdcOps = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "xdc_ops",
		Help: "Total XDCR operations per second for this bucket",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CpuIdleMs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "cpu_idle_ms",
		Help: "CPU idle milliseconds",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CpuLocalMs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "cpu_local_ms",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	CpuUtilizationRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "cpu_utilization_rate",
		Help: "Percentage of CPU in use across all available cores on this server",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	HibernatedRequests = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "hibernated_requests",
		Help: "Number of streaming requests on port 8091 now idle",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	HibernatedWaked = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "hibernated_waked",
		Help: "Rate of streaming request wakeups on port 8091",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	MemActualFree = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "mem_actual_free",
		Help: "Amount of RAM available on this server",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	MemActualUsed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "mem_actual_used",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	MemFree = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "mem_free",
		Help: "Amount of Memory free",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	MemTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "mem_total",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	MemUsedSys = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "mem_used_sys",
		Help: "",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	RestRequests = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "rest_requests",
		Help: "Rate of http requests on port 8091",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	SwapTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "swap_total",
		Help: "Total amount of swap available",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)

	SwapUsed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: FQ_NAMESPACE + subsystem,
		Subsystem: "",
		Name: "swap_used",
		Help: "Amount of swap space in use on this server",
		ConstLabels: nil,
	},
		[]string{"bucket", "node", "cluster"},
	)
)

func strToFloatArr(floatsStr string) []float64 {
	floatsStrArr := strings.Split(floatsStr, " ")
	var floatsArr []float64

	for _, f := range floatsStrArr {
		i, err := strconv.ParseFloat(f, 64)
		if err == nil {
			floatsArr = append(floatsArr, i)
		}
	}

	return floatsArr
}

func setGaugeVec(vec prometheus.GaugeVec, stats []float64, labelValues ...string) {
	if len(stats) > 0 {
		vec.WithLabelValues(labelValues...).Set(stats[len(stats)-1])
	}
}

func getClusterBalancedStatus(c util.Client) (bool, error) {
	node, err := c.Nodes()
	if err != nil {
		return false, fmt.Errorf("unable to retrieve nodes, %s", err)
	}

	return node.Counters.RebalanceSuccess > 0 || (node.Balanced && node.RebalanceStatus == "none"), nil
}

func getCurrentNode(c util.Client) (string, error) {
	nodes, err := c.Nodes()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve nodes: %s", err)
	}

	for _, node := range nodes.Nodes {
		if node.ThisNode { // "ThisNode" is a boolean value indicating that it is the current node
			return node.Hostname, nil // hostname seems to work? just don't use for single node setups
		}
	}

	return "", err
}

func getPerNodeBucketStats(client util.Client, bucketName, nodeName string) map[string]interface{} {
	url := getSpecificNodeBucketStatsURL(client, bucketName, nodeName)

	var bucketStats objects.PerNodeBucketStats
	err := client.Get(url, &bucketStats)
	if err != nil {
		log.Error("unable to GET PerNodeBucketStats %s", err)
	}

	return bucketStats.Op.Samples
}

// /pools/default/buckets/<bucket-name>/nodes/<node-name>/stats
func getSpecificNodeBucketStatsURL(client util.Client, bucket, node string) string {
	servers, err := client.Servers(bucket)
	if err != nil {
		log.Error("unable to retrieve Servers %s", err)
	}

	correctURI := ""
	for _, server := range servers.Servers {
		if server.Hostname == node {
			correctURI = server.Stats["uri"]
		}
	}

	return correctURI
}

func collectPerNodeBucketMetrics(client util.Client, node string, refreshTime int) {

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	clusterName, err := client.ClusterName()
	if err != nil {
		log.Error("%s", err)
		return
	}

	outerErr := util.Retry(ctx, 20*time.Second, 10, func() (bool, error) {

		rebalanced, err := getClusterBalancedStatus(client)
		if err != nil {
			log.Error("Unable to get rebalance status %s", err)
		}

		if !rebalanced {
			log.Info("Waiting for Rebalance... retrying...")
			return false, err
		} else {
			go func() {
				for {
					buckets, err := client.Buckets()
					if err != nil {
						log.Error("Unable to get buckets %s", err)
					}

					for _, bucket := range buckets {
						log.Debug("Collecting per-node bucket stats, node=%s, bucket=%s", node, bucket.Name)

						samples := getPerNodeBucketStats(client, bucket.Name, node)

						setGaugeVec(*AvgDiskUpdateTime, strToFloatArr(fmt.Sprint(samples["avg_disk_update_time"])), bucket.Name, node, clusterName)
						setGaugeVec(*AvgDiskCommitTime, strToFloatArr(fmt.Sprint(samples["avg_disk_commit_time"])), bucket.Name, node, clusterName)
						setGaugeVec(*AvgBgWaitTime, strToFloatArr(fmt.Sprint(samples["avg_bg_wait_seconds"])), bucket.Name, node, clusterName)
						setGaugeVec(*AvgActiveTimestampDrift, strToFloatArr(fmt.Sprint(samples["avg_active_timestamp_drift"])), bucket.Name, node, clusterName)
						setGaugeVec(*AvgReplicaTimestampDrift, strToFloatArr(fmt.Sprint(samples["avg_replica_timestamp_drift"])), bucket.Name, node, clusterName)

						setGaugeVec(*CouchTotalDiskSize, strToFloatArr(fmt.Sprint(samples["couch_total_disk_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchDocsFragmentation, strToFloatArr(fmt.Sprint(samples["couch_docs_fragmentation"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchViewsFragmentation, strToFloatArr(fmt.Sprint(samples["couch_views_fragmentation"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchDocsActualDiskSize, strToFloatArr(fmt.Sprint(samples["couch_docs_actual_disk_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchDocsDataSize, strToFloatArr(fmt.Sprint(samples["couch_docs_data_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchDocsDiskSize, strToFloatArr(fmt.Sprint(samples["couch_docs_disk_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchSpatialDataSize, strToFloatArr(fmt.Sprint(samples["couch_docs_spatial_data_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchSpatialDiskSize, strToFloatArr(fmt.Sprint(samples["couch_docs_spatial_disk_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchSpatialOps, strToFloatArr(fmt.Sprint(samples["couch_spatial_ops"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchViewsActualDiskSize, strToFloatArr(fmt.Sprint(samples["couch_views_actual_disk_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchViewsDataSize, strToFloatArr(fmt.Sprint(samples["couch_views_data_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchViewsDiskSize, strToFloatArr(fmt.Sprint(samples["couch_views_disk_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*CouchViewsOps, strToFloatArr(fmt.Sprint(samples["couch_views_ops"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpCacheMissRate, strToFloatArr(fmt.Sprint(samples["ep_cache_miss_rate"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpResidentItemsRate, strToFloatArr(fmt.Sprint(samples["ep_resident_items_rate"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpActiveAheadExceptions, strToFloatArr(fmt.Sprint(samples["ep_active_ahead_exceptions"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpActiveHlcDrift, strToFloatArr(fmt.Sprint(samples["ep_active_hlc_drift"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpActiveHlcDriftCount, strToFloatArr(fmt.Sprint(samples["ep_active_hlc_drift_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpBgFetched, strToFloatArr(fmt.Sprint(samples["ep_bg_fetched"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpClockCasDriftTheresholExceeded, strToFloatArr(fmt.Sprint(samples["ep_clock_cas_drift_threshold_exceeded"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDataReadFailed, strToFloatArr(fmt.Sprint(samples["ep_data_read_failed"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDataWriteFailed, strToFloatArr(fmt.Sprint(samples["ep_data_write_failed"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDcp2iBackoff, strToFloatArr(fmt.Sprint(samples["ep_dcp_2i_backoff"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcp2iCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_2i_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcp2iItemsRemaining, strToFloatArr(fmt.Sprint(samples["ep_dcp_2i_items_remaining"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcp2iItemsSent, strToFloatArr(fmt.Sprint(samples["ep_dcp_2i_items_sent"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcp2iProducerCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_2i_producers"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcp2iTotalBacklogSize, strToFloatArr(fmt.Sprint(samples["ep_dcp_2i_total_backlog_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcp2iTotalBytes, strToFloatArr(fmt.Sprint(samples["ep_dcp_2i_total_bytes"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDcpCbasBackoff, strToFloatArr(fmt.Sprint(samples["ep_dcp_cbas_backoff"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpCbasCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_cbas_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpCbasItemsRemaining, strToFloatArr(fmt.Sprint(samples["ep_dcp_cbas_items_remaining"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpCbasItemsSent, strToFloatArr(fmt.Sprint(samples["ep_dcp_cbas_items_sent"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpCbasProducerCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_cbas_items_producer_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpCbasTotalBacklogSize, strToFloatArr(fmt.Sprint(samples["ep_dcp_cbas_items_total_backlog_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpCbasTotalBytes, strToFloatArr(fmt.Sprint(samples["ep_dcp_cbas_items_total_bytes"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDcpFtsBackoff, strToFloatArr(fmt.Sprint(samples["ep_dcp_fts_backoff"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpFtsCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_fts_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpFtsItemsRemaining, strToFloatArr(fmt.Sprint(samples["ep_dcp_fts_items_remaining"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpFtsItemsSent, strToFloatArr(fmt.Sprint(samples["ep_dcp_fts_items_sent"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpFtsProducerCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_fts_producer_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpFtsTotalBacklogSize, strToFloatArr(fmt.Sprint(samples["ep_dcp_fts_backlog_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpFtsTotalBytes, strToFloatArr(fmt.Sprint(samples["ep_dcp_fts_total_bytes"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDcpOtherBackoff, strToFloatArr(fmt.Sprint(samples["ep_dcp_other_backoff"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpOtherCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_other_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpOtherItemsRemaining, strToFloatArr(fmt.Sprint(samples["ep_dcp_other_items_remaining"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpOtherItemsSent, strToFloatArr(fmt.Sprint(samples["ep_dcp_other_items_sent"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpOtherProducerCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_other_producer_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpOtherTotalBacklogSize, strToFloatArr(fmt.Sprint(samples["ep_dcp_other_total_backlog_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpOtherTotalBytes, strToFloatArr(fmt.Sprint(samples["ep_dcp_other_total_bytes"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDcpReplicaBackoff, strToFloatArr(fmt.Sprint(samples["ep_dcp_replica_backoff"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpReplicaCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_replica_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpReplicaItemsRemaining, strToFloatArr(fmt.Sprint(samples["ep_dcp_replica_items_remaining"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpReplicaItemsSent, strToFloatArr(fmt.Sprint(samples["ep_dcp_replica_items_sent"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpReplicaProducerCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_replica_producer_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpReplicaTotalBacklogSize, strToFloatArr(fmt.Sprint(samples["ep_dcp_replica_total_backlog_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpReplicaTotalBytes, strToFloatArr(fmt.Sprint(samples["ep_dcp_replica_total_bytes"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDcpViewsBackoff, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_backoff"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsItemsRemaining, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_items_remaining"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsItemsSent, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_items_sent"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsProducerCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_producer_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsTotalBacklogSize, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_total_backlog_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsTotalBytes, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_total_bytes"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDcpViewsIndexesBackoff, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_indexes_backoff"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsIndexesCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_indexes_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsIndexesItemsRemaining, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_indexes_items_remaining"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsIndexesItemsSent, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_indexes_items_sent"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsIndexesProducerCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_indexes_producer_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsIndexesTotalBacklogSize, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_indexes_total_backlog_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpViewsIndexesTotalBytes, strToFloatArr(fmt.Sprint(samples["ep_dcp_views_indexes_total_bytes"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDcpXdcrBackoff, strToFloatArr(fmt.Sprint(samples["ep_dcp_xdcr_backoff"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpXdcrCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_xdcr_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpXdcrItemsRemaining, strToFloatArr(fmt.Sprint(samples["ep_dcp_xdcr_items_remaining"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpXdcrItemsSent, strToFloatArr(fmt.Sprint(samples["ep_dcp_xdcr_items_sent"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpXdcrProducerCount, strToFloatArr(fmt.Sprint(samples["ep_dcp_xdcr_producer_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpXdcrTotalBacklogSize, strToFloatArr(fmt.Sprint(samples["ep_dcp_xdcr_total_backlog_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDcpXdcrTotalBytes, strToFloatArr(fmt.Sprint(samples["ep_dcp_xdcr_total_bytes"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpDiskqueueDrain, strToFloatArr(fmt.Sprint(samples["ep_diskqueue_drain"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDiskqueueFill, strToFloatArr(fmt.Sprint(samples["ep_diskqueue_fill"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpDiskqueueItems, strToFloatArr(fmt.Sprint(samples["ep_diskqueue_items"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpFlusherTodo, strToFloatArr(fmt.Sprint(samples["ep_flusher_todo"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpItemCommitFailed, strToFloatArr(fmt.Sprint(samples["ep_item_commit_failed"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpKvSize, strToFloatArr(fmt.Sprint(samples["ep_kv_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpMaxSize, strToFloatArr(fmt.Sprint(samples["ep_max_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpMemHighWat, strToFloatArr(fmt.Sprint(samples["ep_mem_high_wat"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpMemLowWat, strToFloatArr(fmt.Sprint(samples["ep_mem_low_wat"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpMetaDataMemory, strToFloatArr(fmt.Sprint(samples["ep_meta_data_memory"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpNumNonResident, strToFloatArr(fmt.Sprint(samples["ep_num_non_resident"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpNumOpsDelMeta, strToFloatArr(fmt.Sprint(samples["ep_num_ops_del_meta"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpNumOpsDelRetMeta, strToFloatArr(fmt.Sprint(samples["ep_num_ops_del_ret_meta"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpNumOpsGetMeta, strToFloatArr(fmt.Sprint(samples["ep_num_ops_get_meta"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpNumOpsSetMeta, strToFloatArr(fmt.Sprint(samples["ep_num_ops_set_meta"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpNumOpsSetRetMeta, strToFloatArr(fmt.Sprint(samples["ep_num_ops_set_ret_meta"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpNumValueEjects, strToFloatArr(fmt.Sprint(samples["ep_num_value_ejects"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpOomErrors, strToFloatArr(fmt.Sprint(samples["ep_oom_errors"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpOpsCreate, strToFloatArr(fmt.Sprint(samples["ep_ops_create"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpOpsUpdate, strToFloatArr(fmt.Sprint(samples["ep_ops_update"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpOverhead, strToFloatArr(fmt.Sprint(samples["ep_overhead"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpQueueSize, strToFloatArr(fmt.Sprint(samples["ep_queue_size"])), bucket.Name, node, clusterName)

						setGaugeVec(*EpReplicaAheadExceptions, strToFloatArr(fmt.Sprint(samples["ep_replica_ahead_exceptions"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpReplicaHlcDrift, strToFloatArr(fmt.Sprint(samples["ep_replica_hlc_drift"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpReplicaHlcDriftCount, strToFloatArr(fmt.Sprint(samples["ep_replica_hlc_drift_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpTmpOomErrors, strToFloatArr(fmt.Sprint(samples["ep_tmp_oom_errors"])), bucket.Name, node, clusterName)
						setGaugeVec(*EpVbTotal, strToFloatArr(fmt.Sprint(samples["ep_vb_total"])), bucket.Name, node, clusterName)

						setGaugeVec(*VbAvgActiveQueueAge, strToFloatArr(fmt.Sprint(samples["vb_avg_active_queue_age"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbAvgReplicaQueueAge, strToFloatArr(fmt.Sprint(samples["vb_avg_replica_queue_age"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbAvgPendingQueueAge, strToFloatArr(fmt.Sprint(samples["vb_avg_pending_queue_age"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbAvgTotalQueueAge, strToFloatArr(fmt.Sprint(samples["vb_avg_total_queue_age"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveResidentItemsRatio, strToFloatArr(fmt.Sprint(samples["vb_active_resident_items_ratio"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaResidentItemsRatio, strToFloatArr(fmt.Sprint(samples["vb_replica_resident_items_ratio"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingResidentItemsRatio, strToFloatArr(fmt.Sprint(samples["vb_pending_resident_items_ratio"])), bucket.Name, node, clusterName)

						setGaugeVec(*VbActiveEject, strToFloatArr(fmt.Sprint(samples["vb_active_eject"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveItmMemory, strToFloatArr(fmt.Sprint(samples["vb_active_itm_memory"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveMetaDataMemory, strToFloatArr(fmt.Sprint(samples["vb_active_meta_data_memory"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveNum, strToFloatArr(fmt.Sprint(samples["vb_active_num"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveNumNonresident, strToFloatArr(fmt.Sprint(samples["vb_active_num_non_resident"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveOpsCreate, strToFloatArr(fmt.Sprint(samples["vb_active_ops_create"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveOpsUpdate, strToFloatArr(fmt.Sprint(samples["vb_active_ops_update"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveQueueAge, strToFloatArr(fmt.Sprint(samples["vb_active_queue_age"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveQueueDrain, strToFloatArr(fmt.Sprint(samples["vb_active_queue_drain"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveQueueFill, strToFloatArr(fmt.Sprint(samples["vb_active_queue_fill"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveQueueSize, strToFloatArr(fmt.Sprint(samples["vb_active_queue_size"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbActiveQueueItems, strToFloatArr(fmt.Sprint(samples["vb_active_queue_items"])), bucket.Name, node, clusterName)

						setGaugeVec(*VbPendingCurrItems, strToFloatArr(fmt.Sprint(samples["vb_pending_curr_items"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingEject, strToFloatArr(fmt.Sprint(samples["vb_pending_eject"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingItmMemory, strToFloatArr(fmt.Sprint(samples["vb_pending_itm_memory"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingMetaDataMemory, strToFloatArr(fmt.Sprint(samples["vb_pending_meta_data_memory"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingNum, strToFloatArr(fmt.Sprint(samples["vb_pending_num"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingNumNonResident, strToFloatArr(fmt.Sprint(samples["vb_pending_num_non_resident"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingOpsCreate, strToFloatArr(fmt.Sprint(samples["vb_pending_ops_create"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingOpsUpdate, strToFloatArr(fmt.Sprint(samples["vb_pending_ops_update"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingQueueAge, strToFloatArr(fmt.Sprint(samples["vb_pending_queue_age"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingQueueDrain, strToFloatArr(fmt.Sprint(samples["vb_pending_queue_drain"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingQueueFill, strToFloatArr(fmt.Sprint(samples["vb_pending_queue_fill"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbPendingQueueSize, strToFloatArr(fmt.Sprint(samples["vb_pending_queue_size"])), bucket.Name, node, clusterName)

						setGaugeVec(*VbReplicaCurrItems, strToFloatArr(fmt.Sprint(samples["vb_replica_curr_items"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaEject, strToFloatArr(fmt.Sprint(samples["vb_replica_eject"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaItmMemory, strToFloatArr(fmt.Sprint(samples["vb_replica_itm_memory"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaMetaDataMemory, strToFloatArr(fmt.Sprint(samples["vb_replica_meta_data_memory"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaNum, strToFloatArr(fmt.Sprint(samples["vb_replica_num"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaNumNonResident, strToFloatArr(fmt.Sprint(samples["vb_replica_num_non_resident"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaOpsCreate, strToFloatArr(fmt.Sprint(samples["vb_replica_ops_create"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaOpsUpdate, strToFloatArr(fmt.Sprint(samples["vb_replica_ops_update"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaQueueAge, strToFloatArr(fmt.Sprint(samples["vb_replica_queue_age"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaQueueDrain, strToFloatArr(fmt.Sprint(samples["vb_replica_queue_drain"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaQueueFill, strToFloatArr(fmt.Sprint(samples["vb_replica_queue_fill"])), bucket.Name, node, clusterName)
						setGaugeVec(*VbReplicaQueueSize, strToFloatArr(fmt.Sprint(samples["vb_replica_queue_size"])), bucket.Name, node, clusterName)

						setGaugeVec(*VbTotalQueueAge, strToFloatArr(fmt.Sprint(samples["vb_total_queue_age"])), bucket.Name, node, clusterName)
						setGaugeVec(*HibernatedRequests, strToFloatArr(fmt.Sprint(samples["hibernated_requests"])), bucket.Name, node, clusterName)
						setGaugeVec(*HibernatedRequests, strToFloatArr(fmt.Sprint(samples["hibernated_waked"])), bucket.Name, node, clusterName)
						setGaugeVec(*XdcOps, strToFloatArr(fmt.Sprint(samples["xdc_ops"])), bucket.Name, node, clusterName)
						setGaugeVec(*CpuIdleMs, strToFloatArr(fmt.Sprint(samples["cpu_idle_ms"])), bucket.Name, node, clusterName)
						setGaugeVec(*CpuLocalMs, strToFloatArr(fmt.Sprint(samples["cpu_local_ms"])), bucket.Name, node, clusterName)
						setGaugeVec(*CpuUtilizationRate, strToFloatArr(fmt.Sprint(samples["cpu_utilization_rate"])), bucket.Name, node, clusterName)

						setGaugeVec(*BgWaitCount, strToFloatArr(fmt.Sprint(samples["bg_wait_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*BgWaitTotal, strToFloatArr(fmt.Sprint(samples["bg_wait_total"])), bucket.Name, node, clusterName)
						setGaugeVec(*BytesRead, strToFloatArr(fmt.Sprint(samples["bytes_read"])), bucket.Name, node, clusterName)
						setGaugeVec(*BytesWritten, strToFloatArr(fmt.Sprint(samples["bytes_written"])), bucket.Name, node, clusterName)
						setGaugeVec(*CasBadVal, strToFloatArr(fmt.Sprint(samples["cas_bad_val"])), bucket.Name, node, clusterName)
						setGaugeVec(*CasHits, strToFloatArr(fmt.Sprint(samples["cas_hits"])), bucket.Name, node, clusterName)
						setGaugeVec(*CasMisses, strToFloatArr(fmt.Sprint(samples["cas_misses"])), bucket.Name, node, clusterName)
						setGaugeVec(*CmdGet, strToFloatArr(fmt.Sprint(samples["cmd_get"])), bucket.Name, node, clusterName)
						setGaugeVec(*CmdSet, strToFloatArr(fmt.Sprint(samples["cmd_set"])), bucket.Name, node, clusterName)
						setGaugeVec(*HitRatio, strToFloatArr(fmt.Sprint(samples["hit_ratio"])), bucket.Name, node, clusterName)

						setGaugeVec(*CurrConnections, strToFloatArr(fmt.Sprint(samples["curr_connections"])), bucket.Name, node, clusterName)
						setGaugeVec(*CurrItems, strToFloatArr(fmt.Sprint(samples["curr_items"])), bucket.Name, node, clusterName)
						setGaugeVec(*CurrItemsTot, strToFloatArr(fmt.Sprint(samples["curr_items_tot"])), bucket.Name, node, clusterName)

						setGaugeVec(*DecrHits, strToFloatArr(fmt.Sprint(samples["decr_hits"])), bucket.Name, node, clusterName)
						setGaugeVec(*DecrMisses, strToFloatArr(fmt.Sprint(samples["decr_misses"])), bucket.Name, node, clusterName)
						setGaugeVec(*DeleteHits, strToFloatArr(fmt.Sprint(samples["delete_hits"])), bucket.Name, node, clusterName)
						setGaugeVec(*DeleteMisses, strToFloatArr(fmt.Sprint(samples["delete_misses"])), bucket.Name, node, clusterName)

						setGaugeVec(*DiskCommitCount, strToFloatArr(fmt.Sprint(samples["disk_commit_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*DiskCommitTotal, strToFloatArr(fmt.Sprint(samples["disk_commit_total"])), bucket.Name, node, clusterName)
						setGaugeVec(*DiskUpdateCount, strToFloatArr(fmt.Sprint(samples["disk_update_count"])), bucket.Name, node, clusterName)
						setGaugeVec(*DiskUpdateTotal, strToFloatArr(fmt.Sprint(samples["disk_update_total"])), bucket.Name, node, clusterName)
						setGaugeVec(*DiskWriteQueue, strToFloatArr(fmt.Sprint(samples["disk_write_queue"])), bucket.Name, node, clusterName)

						setGaugeVec(*Evictions, strToFloatArr(fmt.Sprint(samples["evictions"])), bucket.Name, node, clusterName)
						setGaugeVec(*GetHits, strToFloatArr(fmt.Sprint(samples["get_hits"])), bucket.Name, node, clusterName)
						setGaugeVec(*GetMisses, strToFloatArr(fmt.Sprint(samples["get_misses"])), bucket.Name, node, clusterName)
						setGaugeVec(*IncrHits, strToFloatArr(fmt.Sprint(samples["incr_hits"])), bucket.Name, node, clusterName)
						setGaugeVec(*IncrMisses, strToFloatArr(fmt.Sprint(samples["incr_misses"])), bucket.Name, node, clusterName)
						setGaugeVec(*Misses, strToFloatArr(fmt.Sprint(samples["misses"])), bucket.Name, node, clusterName)
						setGaugeVec(*Ops, strToFloatArr(fmt.Sprint(samples["ops"])), bucket.Name, node, clusterName)

						setGaugeVec(*MemActualFree, strToFloatArr(fmt.Sprint(samples["mem_actual_free"])), bucket.Name, node, clusterName)
						setGaugeVec(*MemActualUsed, strToFloatArr(fmt.Sprint(samples["mem_actual_used"])), bucket.Name, node, clusterName)
						setGaugeVec(*MemFree, strToFloatArr(fmt.Sprint(samples["mem_free"])), bucket.Name, node, clusterName)
						setGaugeVec(*MemUsed, strToFloatArr(fmt.Sprint(samples["mem_used"])), bucket.Name, node, clusterName)
						setGaugeVec(*MemTotal, strToFloatArr(fmt.Sprint(samples["mem_total"])), bucket.Name, node, clusterName)
						setGaugeVec(*MemUsedSys, strToFloatArr(fmt.Sprint(samples["mem_used_sys"])), bucket.Name, node, clusterName)
						setGaugeVec(*RestRequests, strToFloatArr(fmt.Sprint(samples["rest_requests"])), bucket.Name, node, clusterName)
						setGaugeVec(*SwapTotal, strToFloatArr(fmt.Sprint(samples["swap_total"])), bucket.Name, node, clusterName)
						setGaugeVec(*SwapUsed, strToFloatArr(fmt.Sprint(samples["swap_used"])), bucket.Name, node, clusterName)

					}
					time.Sleep(time.Second * time.Duration(refreshTime))
				}
			}()
			log.Info("Per Node Bucket Stats Go Thread executed successfully")
			return true, nil
		}
	})
	if outerErr != nil {
		log.Error("Getting Per Node Bucket Stats failed %s", outerErr)
	}
}

func RunPerNodeBucketStatsCollection(client util.Client, refreshTime int) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	outerErr := util.Retry(ctx, 20*time.Second, 8, func() (bool, error) {
		if currNode, err := getCurrentNode(client); err != nil {
			log.Error("could not get current node, will retry. %s", err)
			return false, err
		} else {
			collectPerNodeBucketMetrics(client, currNode, refreshTime)
		}
		return true, nil
	})
	if outerErr != nil {
		fmt.Println("getting default stats failed")
		fmt.Println(outerErr)
	}
}
