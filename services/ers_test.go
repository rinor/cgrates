/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package services

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/ers"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestEventReaderSCoverage for cover testing
func TestEventReaderSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	server := cores.NewServer(nil)
	srv := NewEventReaderService(cfg, nil, filterSChan, shdChan, nil, server, nil, nil, srvDep)

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	srv2 := EventReaderService{
		RWMutex:     sync.RWMutex{},
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		ers:         &ers.ERService{},
		rldChan:     make(chan struct{}, 1),
		stopChan:    make(chan struct{}, 1),
		connMgr:     nil,
		srvDep:      srvDep,
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.ERs) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ERs, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	srv2.Shutdown()
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
