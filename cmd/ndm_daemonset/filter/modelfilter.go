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

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	modelFilterKey     = "model-filter"
	modelValueMayastor = "Mayastor NVMe controller"
)

var (
	modelFilterName       = "model filter"
	modelFilterState      = defaultEnabled
	includeModels         = ""
	excludeModels         = ""
	defaultExcludedModels = []string{modelValueMayastor}
)

var modelFilterRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		return
	}
	if ctrl.NDMConfig != nil {
		for _, filterConfig := range ctrl.NDMConfig.FilterConfigs {
			if filterConfig.Key == modelFilterKey {
				modelFilterName = filterConfig.Name
				modelFilterState = util.CheckTruthy(filterConfig.State)
				includeModels = filterConfig.Include
				excludeModels = filterConfig.Exclude
				break
			}
		}
	}
	var fi controller.FilterInterface = newModelFilter(ctrl)
	newRegisterFilter := &registerFilter{
		name:       modelFilterName,
		state:      modelFilterState,
		fi:         fi,
		controller: ctrl,
	}
	newRegisterFilter.register()
}

type modelFilter struct {
	controller    *controller.Controller
	excludeModels []string
	includeModels []string
}

func newModelFilter(ctrl *controller.Controller) *modelFilter {
	return &modelFilter{
		controller: ctrl,
	}
}

func (vf *modelFilter) Start() {
	vf.includeModels = make([]string, 0)
	vf.excludeModels = make([]string, 0)

	vf.excludeModels = append(vf.excludeModels, defaultExcludedModels...)

	if includeModels != "" {
		vf.includeModels = strings.Split(includeModels, ",")
	}
	if excludeModels != "" {
		vf.excludeModels = append(vf.excludeModels, strings.Split(excludeModels, ",")...)
	}
}

func (vf *modelFilter) Include(blockDevice *blockdevice.BlockDevice) bool {
	if len(vf.includeModels) == 0 {
		return true
	}
	return util.ContainsIgnoredCase(vf.includeModels, blockDevice.DeviceAttributes.Model)
}

func (vf *modelFilter) Exclude(blockDevice *blockdevice.BlockDevice) bool {
	if len(vf.excludeModels) == 0 {
		return true
	}
	return !util.ContainsIgnoredCase(vf.excludeModels, blockDevice.DeviceAttributes.Model)
}
