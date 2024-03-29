/*
Copyright 2022-present The Ztalab Authors.
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

package models

// Statistic is the representation of a statistic used by the monitoring service.
type Statistic struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Values map[string]interface{} `json:"values"`
}

// NewStatistic returns an initialized Statistic.
func NewStatistic(name string) Statistic {
	return Statistic{
		Name:   name,
		Tags:   make(map[string]string),
		Values: make(map[string]interface{}),
	}
}

// StatisticTags is a map that can be merged with others without causing
// mutations to either map.
type StatisticTags map[string]string

// Merge creates a new map containing the merged contents of tags and t.
// If both tags and the receiver map contain the same key, the value in tags
// is used in the resulting map.
//
// Merge always returns a usable map.
func (t StatisticTags) Merge(tags map[string]string) map[string]string {
	// Add everything in tags to the result.
	out := make(map[string]string, len(tags))
	for k, v := range tags {
		out[k] = v
	}

	// Only add values from t that don't appear in tags.
	for k, v := range t {
		if _, ok := tags[k]; !ok {
			out[k] = v
		}
	}
	return out
}
