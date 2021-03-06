//
// Copyright 2016 Authors of Cilium
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
//
package policy

import (
	"strconv"

	"github.com/cilium/cilium/bpf/policymap"
	"github.com/cilium/cilium/pkg/labels"
)

type Consumer struct {
	ID           uint32
	Reverse      *Consumer
	DeletionMark bool
	Decision     ConsumableDecision
}

func (c *Consumer) DeepCopy() *Consumer {
	cpy := &Consumer{ID: c.ID,
		DeletionMark: c.DeletionMark,
		Decision:     c.Decision,
	}
	if c.Reverse != nil {
		cpy.Reverse = c.Reverse.DeepCopy()
	}
	return cpy
}

func NewConsumer(id uint32) *Consumer {
	return &Consumer{ID: id, Decision: ACCEPT}
}

type Consumable struct {
	ID           uint32                       `json:"id"`
	Iteration    int                          `json:"-"`
	Labels       *labels.SecCtxLabel          `json:"labels"`
	LabelList    []labels.Label               `json:"-"`
	Maps         map[int]*policymap.PolicyMap `json:"-"`
	Consumers    map[string]*Consumer         `json:"consumers"`
	ReverseRules map[uint32]*Consumer         `json:"-"`
}

func (c *Consumable) DeepCopy() *Consumable {
	cpy := &Consumable{
		ID:           c.ID,
		Iteration:    c.Iteration,
		LabelList:    make([]labels.Label, len(c.LabelList)),
		Maps:         make(map[int]*policymap.PolicyMap, len(c.Maps)),
		Consumers:    make(map[string]*Consumer, len(c.Consumers)),
		ReverseRules: make(map[uint32]*Consumer, len(c.ReverseRules)),
	}
	copy(cpy.LabelList, c.LabelList)
	if c.Labels != nil {
		cpy.Labels = c.Labels.DeepCopy()
	}
	for k, v := range c.Maps {
		cpy.Maps[k] = v.DeepCopy()
	}
	for k, v := range c.Consumers {
		cpy.Consumers[k] = v.DeepCopy()
	}
	for k, v := range c.ReverseRules {
		cpy.ReverseRules[k] = v.DeepCopy()
	}
	return cpy
}

func newConsumable(id uint32, lbls *labels.SecCtxLabel) *Consumable {
	consumable := &Consumable{
		ID:           id,
		Iteration:    0,
		Labels:       lbls,
		Maps:         map[int]*policymap.PolicyMap{},
		Consumers:    map[string]*Consumer{},
		ReverseRules: map[uint32]*Consumer{},
	}

	if lbls != nil {
		consumable.LabelList = make([]labels.Label, len(lbls.Labels))
		idx := 0
		for k, v := range lbls.Labels {
			consumable.LabelList[idx] = labels.Label{
				Key:    k,
				Value:  v.Value,
				Source: v.Source,
			}
			idx++
		}
	}

	return consumable
}

var consumableCache = map[uint32]*Consumable{}

func GetConsumable(id uint32, lbls *labels.SecCtxLabel) *Consumable {
	if v, ok := consumableCache[id]; ok {
		return v
	}

	consumableCache[id] = newConsumable(id, lbls)

	return consumableCache[id]
}

func LookupConsumable(id uint32) *Consumable {
	v, _ := consumableCache[id]
	return v
}

func (c *Consumable) AddMap(m *policymap.PolicyMap) {
	if c.Maps == nil {
		c.Maps = make(map[int]*policymap.PolicyMap)
	}

	// Check if map is already associated with this consumable
	if _, ok := c.Maps[m.Fd]; ok {
		return
	}

	log.Debugf("Adding map %v to consumable %v", m, c)
	c.Maps[m.Fd] = m

	// Populate the new map with the already established consumers of
	// this consumable
	for _, c := range c.Consumers {
		if err := m.AllowConsumer(c.ID); err != nil {
			log.Warningf("Update of policy map failed: %s\n", err)
		}
	}
}

func deleteReverseRule(consumable, consumer uint32) {
	if reverse := LookupConsumable(consumable); reverse != nil {
		delete(reverse.ReverseRules, consumer)
		if reverse.wasLastRule(consumer) {
			reverse.removeFromMaps(consumer)
		}
	}
}

func (c *Consumable) Delete() {
	for _, consumer := range c.Consumers {
		// FIXME: This explicit removal could be removed eventually to
		// speed things up as the policy map should get deleted anyway
		if c.wasLastRule(consumer.ID) {
			c.removeFromMaps(consumer.ID)
		}

		deleteReverseRule(consumer.ID, c.ID)
	}

	delete(consumableCache, c.ID)
}

func (c *Consumable) RemoveMap(m *policymap.PolicyMap) {
	if m != nil {
		delete(c.Maps, m.Fd)
		log.Debugf("Removing map %v from consumable %v, new len %d", m, c, len(c.Maps))

		// If the last map of the consumable is gone the consumable is no longer
		// needed and should be removed from the cache and all cross references
		// must be undone.
		if len(c.Maps) == 0 {
			c.Delete()
		}
	}

}

func (c *Consumable) Consumer(id uint32) *Consumer {
	val, _ := c.Consumers[strconv.FormatUint(uint64(id), 10)]
	return val
}

func (c *Consumable) isNewRule(id uint32) bool {
	r1 := c.ReverseRules[id] != nil
	r2 := c.Consumers[strconv.FormatUint(uint64(id), 10)] != nil

	// golang has no XOR ... whaaa?
	return (r1 || r2) && !(r1 && r2)
}

func (c *Consumable) addToMaps(id uint32) {
	for _, m := range c.Maps {
		log.Debugf("Updating policy BPF map %s: allowing %d\n", m.String(), id)
		if err := m.AllowConsumer(id); err != nil {
			log.Warningf("Update of policy map failed: %s\n", err)
		}
	}
}

func (c *Consumable) wasLastRule(id uint32) bool {
	return c.ReverseRules[id] == nil && c.Consumers[strconv.FormatUint(uint64(id), 10)] == nil
}

func (c *Consumable) removeFromMaps(id uint32) {
	for _, m := range c.Maps {
		log.Debugf("Updating policy BPF map %s: denying %d\n", m.String(), id)
		if err := m.DeleteConsumer(id); err != nil {
			log.Warningf("Update of policy map failed: %s\n", err)
		}
	}
}

func (c *Consumable) AllowConsumer(id uint32) *Consumer {
	var consumer *Consumer

	if consumer = c.Consumer(id); consumer == nil {
		log.Debugf("New consumer %d for consumable %v", id, c)
		consumer = NewConsumer(id)
		c.Consumers[strconv.FormatUint(uint64(id), 10)] = consumer

		if c.isNewRule(id) {
			c.addToMaps(id)
		}
	} else {
		consumer.DeletionMark = false
	}

	return consumer
}

func (c *Consumable) AllowConsumerAndReverse(id uint32) {
	log.Debugf("Allowing direction %d -> %d\n", id, c.ID)
	fwd := c.AllowConsumer(id)

	if reverse := LookupConsumable(id); reverse != nil {
		log.Debugf("Allowing reverse direction %d -> %d\n", c.ID, id)
		if _, ok := reverse.ReverseRules[c.ID]; !ok {
			fwd.Reverse = NewConsumer(c.ID)
			reverse.ReverseRules[c.ID] = fwd.Reverse
			if reverse.isNewRule(c.ID) {
				reverse.addToMaps(c.ID)
			}
		}
	} else {
		log.Warningf("Allowed a consumer %d->%d which can't be found in the reverse direction", c.ID, id)
	}
}

func (c *Consumable) BanConsumer(id uint32) {
	n := strconv.FormatUint(uint64(id), 10)

	if consumer, ok := c.Consumers[n]; ok {
		log.Debugf("Removing consumer %v\n", consumer)
		delete(c.Consumers, n)
		if c.wasLastRule(id) {
			c.removeFromMaps(id)
		}

		if consumer.Reverse != nil {
			deleteReverseRule(id, c.ID)
		}
	}
}

func (c *Consumable) Allows(id uint32) bool {
	if consumer := c.Consumer(id); consumer != nil {
		if consumer.Decision == ACCEPT {
			return true
		}
	}

	return false
}
