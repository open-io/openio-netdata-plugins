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
	"oionetdata/util"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("argument required")
	}
	var conf string
	var full bool
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&conf, "conf", "/etc/netdata/oiofs.conf", "Path to endpoint config file")
	fs.BoolVar(&full, "full", false, "Gather all metrics")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatalln("ERROR: Command plugin: Could not parse args", err)
	}
	intervalSeconds := collector.ParseIntervalSeconds(os.Args[1])

	var endpoints []oiofs.Endpoint

	out, err := util.OiofsEndpoints(conf)
	if err != nil {
		log.Fatalln("ERROR: Oiofs plugin: Could not load oiofs endpoints", err)
	}

	for name, url := range out {
		endpoints = append(endpoints, oiofs.Endpoint{Path: name, URL: url})
	}

	writer := netdata.NewDefaultWriter()
	worker := netdata.NewWorker(time.Duration(intervalSeconds)*time.Second, writer)

	for _, endpoint := range endpoints {
		collector := oiofs.NewCollector(endpoint, full)
		worker.AddCollector(collector)
		family := endpoint.Path
		fsType := fmt.Sprintf("oiofs.%s", endpoint.Path)

		if full {
			// Metadata counters
			metaCount := netdata.NewChart(fsType, "meta_count", "", "Metadata Counters", "ops", family, "oiofs.meta.")
			for _, op := range oiofs.Ops["metaDebug"] {
				metaCount.AddDimension(
					fmt.Sprintf("Meta_%s_count", strings.Title(op)),
					strings.TrimSpace(op),
					netdata.IncrementalAlgorithm,
				)
			}
			worker.AddChart(metaCount, collector)

			// Metadata latency
			metaLatency := netdata.NewChart(fsType, "meta_latency", "", "Metadata latency", "us", family, "oiofs.meta_latency")
			for _, op := range oiofs.Ops["metaDebug"] {
				metaLatency.AddDimension(
					fmt.Sprintf("Meta_%s_total_us", strings.Title(op)),
					strings.TrimSpace(op),
					netdata.AbsoluteAlgorithm,
				)
			}
			worker.AddChart(metaLatency, collector)
		}

		// Cache size
		cacheSizeStats := netdata.NewChart(fsType, "cache_size", "", "Cache Size", "bytes", family, "oiofs.cache_size")
		cacheSizeStats.AddDimension("cache_chunk_used_byte", "used", netdata.AbsoluteAlgorithm)
		cacheSizeStats.AddDimension("cache_chunk_total_byte", "total", netdata.AbsoluteAlgorithm)
		cacheSizeStats.AddDimension("cache_read_total_byte", "read", netdata.AbsoluteAlgorithm)
		worker.AddChart(cacheSizeStats, collector)

		// Cache age
		cacheAgeStats := netdata.NewChart(fsType, "cache_age", "", "Cache age", "us", family, "oiofs.cache_age")
		cacheAgeStats.AddDimension("cache_chunk_avg_age_microseconds", "age", netdata.AbsoluteAlgorithm)
		worker.AddChart(cacheAgeStats, collector)

		// Cache latency
		cacheLatency := netdata.NewChart(fsType, "cache_latency", "", "Cache latency", "us", family, "oiofs.cache_latency")
		cacheLatency.AddDimension("cache_read_total_us", "read", netdata.AbsoluteAlgorithm)
		worker.AddChart(cacheLatency, collector)

		// Cache read
		cacheReadStats := netdata.NewChart(fsType, "cache_read", "", "Cache read", "ops", family, "oiofs.cache")
		cacheReadStats.AddDimension("cache_read_count", "total", netdata.IncrementalAlgorithm)
		cacheReadStats.AddDimension("cache_read_hit", "hit", netdata.IncrementalAlgorithm)
		cacheReadStats.AddDimension("cache_read_miss", "miss", netdata.IncrementalAlgorithm)
		worker.AddChart(cacheReadStats, collector)

		// Cache chunks
		cacheChunks := netdata.NewChart(fsType, "cache_chunks", "", "Cache chunks", "chunks", family, "oiofs.cache")
		cacheChunks.AddDimension("cache_chunk_count", "chunks", netdata.AbsoluteAlgorithm)
		worker.AddChart(cacheChunks, collector)

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
			worker.AddChart(fuseCount, collector)

			// Fuse latency
			fuseLatency := netdata.NewChart(fsType, "fuse_latency", "", "Fuse latency", "us", family, "oiofs.fuse_latency")
			for _, op := range oiofs.Ops["fuseDebug"] {
				fuseLatency.AddDimension(
					fmt.Sprintf("fuse_%s_total_us", op),
					strings.TrimSpace(op),
					netdata.AbsoluteAlgorithm,
				)
			}
			worker.AddChart(fuseLatency, collector)
		}

		// Fuse I/O
		fuseIO := netdata.NewChart(fsType, "fuse_io", "", "Fuse I/O", "bytes", family, "oiofs.fuse")
		fuseIO.AddDimension("fuse_read_total_byte", "read", netdata.IncrementalAlgorithm)
		fuseIO.AddDimension("fuse_write_total_byte", "write", netdata.IncrementalAlgorithm)
		worker.AddChart(fuseIO, collector)

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
			worker.AddChart(sdsCount, collector)

			// SDS latency
			sdsLatency := netdata.NewChart(fsType, "sds_latency", "", "SDS latency", "us", family, "oiofs.sds_latency")
			for _, op := range oiofs.Ops["sdsDebug"] {
				sdsLatency.AddDimension(
					fmt.Sprintf("fuse_%s_count", strings.Title(op)),
					strings.TrimSpace(op),
					netdata.IncrementalAlgorithm,
				)
			}
			worker.AddChart(sdsLatency, collector)
		}

		sdsUpload := netdata.NewChart(fsType, "sds_upload", "", "SDS uploads", "ops", family, "oiofs.sds_ul")
		sdsUpload.AddDimension("sds_upload_failed", "failed", netdata.IncrementalAlgorithm)
		sdsUpload.AddDimension("sds_upload_succeeded", "succeeded", netdata.IncrementalAlgorithm)
		worker.AddChart(sdsUpload, collector)

		sdsDownload := netdata.NewChart(fsType, "sds_download", "", "SDS downloads", "ops", family, "oiofs.sds_dl")
		sdsDownload.AddDimension("sds_download_failed", "failed", netdata.IncrementalAlgorithm)
		sdsDownload.AddDimension("sds_download_succeeded", "succeeded", netdata.IncrementalAlgorithm)
		worker.AddChart(sdsDownload, collector)

		sdsData := netdata.NewChart(fsType, "sds_data", "", "SDS data", "bytes", family, "oiofs.sds")
		sdsData.AddDimension("sds_download_total_byte", "download", netdata.IncrementalAlgorithm)
		sdsData.AddDimension("sds_upload_total_byte", "upload", netdata.IncrementalAlgorithm)
		worker.AddChart(sdsData, collector)
	}

	worker.Run()
}
