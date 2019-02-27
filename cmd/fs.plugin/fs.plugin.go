// OpenIO netdata collectors
// Copyright (C) 2019 OpenIO SAS
//
// This library is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3.0 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"oionetdata/collector"
	"oionetdata/netdata"
	"oionetdata/oiofs"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var addr string
	var id string
	var full bool
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&addr, "addr", "localhost:6999", "IP:PORT of oiofs stats route")
	fs.StringVar(&id, "id", "default", "Connector identifier (alphanumeric)")
	fs.BoolVar(&full, "full", false, "Gather all metrics")
	fs.Parse(os.Args[2:])
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	writer := netdata.NewDefaultWriter()
	collector := oiofs.NewCollector(addr, full)
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer, collector)

	fsType := fmt.Sprintf("oiofs_%s", id)
	family := "oiofs"

	if full {
		// Metadata counters
		metaCount := netdata.NewChart(fsType, "meta_count", "", "Metadata Counters", "ops", family, "oiofs.meta")
		for _, op := range oiofs.Ops["metaDebug"] {
			metaCount.AddDimension(
				fmt.Sprintf("Meta_%s_count", strings.Title(op)),
				strings.TrimSpace(op),
				netdata.IncrementalAlgorithm,
			)
		}
		worker.AddChart(metaCount)

		// Metadata latency
		metaLatency := netdata.NewChart(fsType, "meta_latency", "", "Metadata latency", "us", family, "oiofs.meta_latency")
		for _, op := range oiofs.Ops["metaDebug"] {
			metaLatency.AddDimension(
				fmt.Sprintf("Meta_%s_total_us", strings.Title(op)),
				strings.TrimSpace(op),
				netdata.AbsoluteAlgorithm,
			)
		}
		worker.AddChart(metaLatency)
	}

	// Cache size
	cacheSizeStats := netdata.NewChart(fsType, "cache_size", "", "Cache Size", "bytes", family, "oiofs.cache_size")
	cacheSizeStats.AddDimension("cache_chunk_used_byte", "used", netdata.AbsoluteAlgorithm)
	cacheSizeStats.AddDimension("cache_chunk_total_byte", "total", netdata.AbsoluteAlgorithm)
	cacheSizeStats.AddDimension("cache_read_total_byte", "read", netdata.AbsoluteAlgorithm)
	worker.AddChart(cacheSizeStats)

	// Cache age
	cacheAgeStats := netdata.NewChart(fsType, "cache_age", "", "Cache age", "us", family, "oiofs.cache_age")
	cacheAgeStats.AddDimension("cache_chunk_avg_age_microseconds", "age", netdata.AbsoluteAlgorithm)
	worker.AddChart(cacheAgeStats)

	// Cache latency
	cacheLatency := netdata.NewChart(fsType, "cache_latency", "", "Cache latency", "us", family, "oiofs.cache_latency")
	cacheLatency.AddDimension("cache_read_total_us", "read", netdata.AbsoluteAlgorithm)
	worker.AddChart(cacheLatency)

	// Cache read
	cacheReadStats := netdata.NewChart(fsType, "cache_read", "", "Cache read", "ops", family, "oiofs.cache")
	cacheReadStats.AddDimension("cache_read_count", "total", netdata.IncrementalAlgorithm)
	cacheReadStats.AddDimension("cache_read_hit", "hit", netdata.IncrementalAlgorithm)
	cacheReadStats.AddDimension("cache_read_miss", "miss", netdata.IncrementalAlgorithm)
	worker.AddChart(cacheReadStats)

	// Cache chunks
	cacheChunks := netdata.NewChart(fsType, "cache_chunks", "", "Cache chunks", "chunks", family, "oiofs.cache")
	cacheChunks.AddDimension("cache_chunk_count", "chunks", netdata.AbsoluteAlgorithm)
	worker.AddChart(cacheChunks)

	if full {
		// Fuse counters
		fuseCount := netdata.NewChart(fsType, "fuse_count", "", "Fuse counters", "ops", family, "oiofs.fuse")
		for _, op := range oiofs.Ops["fuseDebug"] {
			fuseCount.AddDimension(
				fmt.Sprintf("fuse_%s_count", op),
				strings.TrimSpace(op),
				netdata.IncrementalAlgorithm,
			)
		}
		worker.AddChart(fuseCount)

		// Fuse latency
		fuseLatency := netdata.NewChart(fsType, "fuse_latency", "", "Fuse latency", "us", family, "oiofs.fuse_latency")
		for _, op := range oiofs.Ops["fuseDebug"] {
			fuseLatency.AddDimension(
				fmt.Sprintf("fuse_%s_total_us", op),
				strings.TrimSpace(op),
				netdata.AbsoluteAlgorithm,
			)
		}
		worker.AddChart(fuseLatency)
	}

	// Fuse I/O
	fuseIO := netdata.NewChart(fsType, "fuse_io", "", "Fuse I/O", "bytes", family, "oiofs.fuse")
	fuseIO.AddDimension("fuse_read_total_byte", "read", netdata.IncrementalAlgorithm)
	fuseIO.AddDimension("fuse_write_total_byte", "write", netdata.IncrementalAlgorithm)
	worker.AddChart(fuseIO)

	if full {
		// SDS counters
		sdsCount := netdata.NewChart(fsType, "sds_count", "", "SDS counters", "ops", family, "oiofs.sds")
		for _, op := range oiofs.Ops["sdsDebug"] {
			sdsCount.AddDimension(
				fmt.Sprintf("sds_%s_count", strings.Title(op)),
				strings.TrimSpace(op),
				netdata.IncrementalAlgorithm,
			)
		}
		worker.AddChart(sdsCount)

		// SDS latency
		sdsLatency := netdata.NewChart(fsType, "sds_latency", "", "SDS latency", "us", family, "oiofs.sds_latency")
		for _, op := range oiofs.Ops["sdsDebug"] {
			sdsLatency.AddDimension(
				fmt.Sprintf("fuse_%s_count", strings.Title(op)),
				strings.TrimSpace(op),
				netdata.IncrementalAlgorithm,
			)
		}
		worker.AddChart(sdsLatency)
	}

	sdsUpload := netdata.NewChart(fsType, "sds_upload", "", "SDS uploads", "ops", family, "oiofs.sds_ul")
	sdsUpload.AddDimension("sds_upload_failed", "failed", netdata.IncrementalAlgorithm)
	sdsUpload.AddDimension("sds_upload_succeeded", "succeeded", netdata.IncrementalAlgorithm)
	worker.AddChart(sdsUpload)

	sdsDownload := netdata.NewChart(fsType, "sds_download", "", "SDS downloads", "ops", family, "oiofs.sds_dl")
	sdsDownload.AddDimension("sds_download_failed", "failed", netdata.IncrementalAlgorithm)
	sdsDownload.AddDimension("sds_download_succeeded", "succeeded", netdata.IncrementalAlgorithm)
	worker.AddChart(sdsDownload)

	sdsData := netdata.NewChart(fsType, "sds_data", "", "SDS data", "bytes", family, "oiofs.sds")
	sdsData.AddDimension("sds_download_total_byte", "download", netdata.IncrementalAlgorithm)
	sdsData.AddDimension("sds_upload_total_byte", "upload", netdata.IncrementalAlgorithm)
	worker.AddChart(sdsData)

	worker.Run()
}
