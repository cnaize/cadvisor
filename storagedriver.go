// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cnaize/cadvisor/manager"
	"github.com/cnaize/cadvisor/storage"
	"github.com/cnaize/cadvisor/storage/bigquery"
	"github.com/cnaize/cadvisor/storage/influxdb"
	"github.com/cnaize/cadvisor/storage/memory"
	"github.com/golang/glog"
)

var argDbUsername = flag.String("storage_driver_user", "root", "database username")
var argDbPassword = flag.String("storage_driver_password", "root", "database password")
var argDbHost = flag.String("storage_driver_host", "localhost:8086", "database host:port")
var argDbName = flag.String("storage_driver_db", "cadvisor", "database name")
var argDbTable = flag.String("storage_driver_table", "stats", "table name")
var argDbIsSecure = flag.Bool("storage_driver_secure", false, "use secure connection with database")
var argDbBufferDuration = flag.Duration("storage_driver_buffer_duration", 60*time.Second, "Writes in the storage driver will be buffered for this duration, and committed to the non memory backends as a single transaction")

const statsRequestedByUI = 60

// Creates a memory storage with an optional backend storage option.
func NewMemoryStorage(backendStorageName string) (*memory.InMemoryStorage, error) {
	var storageDriver *memory.InMemoryStorage
	var backendStorage storage.StorageDriver
	var err error
	// TODO(vmarmol): We shouldn't need the housekeeping interval here and it shouldn't be public.
	statsToCache := int(*argDbBufferDuration / *manager.HousekeepingInterval)
	if statsToCache < statsRequestedByUI {
		// The UI requests the most recent 60 stats by default.
		statsToCache = statsRequestedByUI
	}
	switch backendStorageName {
	case "":
		backendStorage = nil
	case "influxdb":
		var hostname string
		hostname, err = os.Hostname()
		if err != nil {
			return nil, err
		}

		backendStorage, err = influxdb.New(
			hostname,
			*argDbTable,
			*argDbName,
			*argDbUsername,
			*argDbPassword,
			*argDbHost,
			*argDbIsSecure,
			*argDbBufferDuration,
		)
	case "bigquery":
		var hostname string
		hostname, err = os.Hostname()
		if err != nil {
			return nil, err
		}
		backendStorage, err = bigquery.New(
			hostname,
			*argDbTable,
			*argDbName,
		)
	default:
		err = fmt.Errorf("unknown backend storage driver: %v", *argDbDriver)
	}
	if err != nil {
		return nil, err
	}
	if backendStorageName != "" {
		glog.Infof("Using backend storage type %q", backendStorageName)
	} else {
		glog.Infof("No backend storage selected")
	}
	glog.Info("Caching %d stats in memory", statsToCache)
	storageDriver = memory.New(statsToCache, backendStorage)
	return storageDriver, nil
}
