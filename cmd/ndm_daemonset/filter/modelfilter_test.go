/*
Copyright 2023 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filter

import (
	"strings"
	"sync"
	"testing"

	. "github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"

	"github.com/stretchr/testify/assert"
)

func TestModelFilterRegister(t *testing.T) {
	expectedFilterList := make([]*controller.Filter, 0)
	fakeController := &controller.Controller{
		Filters: make([]*controller.Filter, 0),
		Mutex:   &sync.Mutex{},
	}
	go func() {
		controller.ControllerBroadcastChannel <- fakeController
	}()
	modelFilterRegister()
	var fi controller.FilterInterface = &modelFilter{
		controller:    fakeController,
		includeModels: make([]string, 0),
		excludeModels: []string{modelValueMayastor},
	}
	filter := &controller.Filter{
		Name:      modelFilterName,
		State:     modelFilterState,
		Interface: fi,
	}
	expectedFilterList = append(expectedFilterList, filter)
	tests := map[string]struct {
		actualFilterList   []*controller.Filter
		expectedFilterList []*controller.Filter
	}{
		"add one filter and check if it is present or not": {actualFilterList: fakeController.Filters, expectedFilterList: expectedFilterList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedFilterList, test.actualFilterList)
		})
	}
}

func TestModelStart(t *testing.T) {
	fakeModelFilter1 := modelFilter{}
	fakeModelFilter2 := modelFilter{}
	tests := map[string]struct {
		filter       modelFilter
		includeModel string
		excludeModel string
	}{
		"includeModel is empty":          {filter: fakeModelFilter1, includeModel: "", excludeModel: ""},
		"includeModel and model is same": {filter: fakeModelFilter2, includeModel: "Virtual_Disk", excludeModel: "Virtual_Disk"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			includeModels = test.includeModel
			excludeModels = test.excludeModel
			test.filter.Start()

			// even if no models are specified in the filter config
			// by default the registered filter will have mayastor model for excluding
			excludedModels := []string{modelValueMayastor}
			if test.excludeModel != "" {
				excludedModels = append(excludedModels, strings.Split(test.excludeModel, ",")...)
			}

			assert.Equal(t, excludedModels, test.filter.excludeModels)

			if test.includeModel != "" {
				assert.Equal(t, strings.Split(test.excludeModel, ","), test.filter.includeModels)
			} else {
				assert.Equal(t, make([]string, 0), test.filter.includeModels)
			}
		})
	}
}

func TestModelFilterExclude(t *testing.T) {
	fakeModelFilter1 := modelFilter{}
	fakeModelFilter2 := modelFilter{}
	fakeModelFilter3 := modelFilter{}
	tests := map[string]struct {
		filter       modelFilter
		excludeModel string
		model        string
		expected     bool
	}{
		"excludeModel is empty":              {filter: fakeModelFilter1, excludeModel: "", model: "Virtual_Disk", expected: true},
		"excludeModel and model is same":     {filter: fakeModelFilter2, excludeModel: "ST1200MM0007", model: "ST1200MM0007", expected: false},
		"excludeModel and model is not same": {filter: fakeModelFilter3, excludeModel: "ST1200MM0007", model: "Virtual_Disk", expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bd := &BlockDevice{}
			bd.DeviceAttributes.Model = test.model
			if test.excludeModel != "" {
				test.filter.excludeModels = strings.Split(test.excludeModel, ",")
			}
			assert.Equal(t, test.expected, test.filter.Exclude(bd))
		})
	}
}

func TestModelFilterInclude(t *testing.T) {
	fakeModelFilter1 := modelFilter{}
	fakeModelFilter2 := modelFilter{}
	fakeModelFilter3 := modelFilter{}
	tests := map[string]struct {
		filter       modelFilter
		includeModel string
		model        string
		expected     bool
	}{
		"includeModel is empty":              {filter: fakeModelFilter1, includeModel: "", model: "Virtual_Disk", expected: true},
		"includeModel and model is same":     {filter: fakeModelFilter2, includeModel: "ST1200MM0007", model: "ST1200MM0007", expected: true},
		"includeModel and model is not same": {filter: fakeModelFilter3, includeModel: "ST1200MM0007", model: "Virtual_Disk", expected: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bd := &BlockDevice{}
			bd.DeviceAttributes.Model = test.model
			if test.includeModel != "" {
				test.filter.includeModels = strings.Split(test.includeModel, ",")
			}
			assert.Equal(t, test.expected, test.filter.Include(bd))
		})
	}
}
