// Copyright 2026 PlatFormer Authors
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

package controller

import platformerv1 "github.com/platformer-io/platformer/api/v1"

type tierConfig struct {
	MemoryMB    int32
	TimeoutSecs int32
}

// tierDefaults maps named tiers to their Lambda resource allocations.
// "custom" is intentionally absent — spec values are used as-is.
var tierDefaults = map[string]tierConfig{
	"small":  {MemoryMB: 128, TimeoutSecs: 10},
	"medium": {MemoryMB: 512, TimeoutSecs: 30},
	"large":  {MemoryMB: 1024, TimeoutSecs: 60},
}

// applyTierDefaults overwrites MemoryMB and TimeoutSecs from the tier profile.
// Tier "custom" (or unset) is a no-op — spec values are used as-is.
func applyTierDefaults(spec *platformerv1.ServerlessAppSpec) {
	cfg, ok := tierDefaults[spec.Tier]
	if !ok {
		return
	}
	spec.MemoryMB = cfg.MemoryMB
	spec.TimeoutSecs = cfg.TimeoutSecs
}
