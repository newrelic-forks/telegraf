package rate

import (
	"fmt"
	"reflect"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type cache struct {
	store map[uint64][]telegraf.Metric
}

func newCache() *cache {
	var c cache
	c.clear()
	return &c
}

func (c *cache) keys() []uint64 {
	var keys []uint64
	for key := range c.store {
		keys = append(keys, key)
	}
	return keys
}

func (c *cache) add(m telegraf.Metric) {
	id := m.HashID()
	c.store[id] = append(c.store[id], m)
}

func (c *cache) first(key uint64) telegraf.Metric {
	return c.store[key][0]
}

func (c *cache) last(key uint64) telegraf.Metric {
	return c.store[key][len(c.store[key])-1]
}

func (c *cache) clear() {
	c.store = make(map[uint64][]telegraf.Metric)
}

// Rate calculates a rate of change based on values seen since last call to Reset
type Rate struct {
	cache      *cache
	EventNames map[string][]string
}

// Reset clears the cached metric store
func (m *Rate) Reset() {
	m.cache.clear()
}

// NewRate return a properly configured Rate struct
func NewRate() telegraf.Aggregator {
	mm := &Rate{
		cache:      newCache(),
		EventNames: map[string][]string{},
	}
	return mm
}

var sampleConfig = `
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
`

// SampleConfig returns a sane configuration that can be used to configure this plugin
func (m *Rate) SampleConfig() string {
	return sampleConfig
}

// Description describes this plugin
func (m *Rate) Description() string {
	return "Calculate the rate of counters of each Docker network stats object"
}

// Add caches a metric for later calculation
func (m *Rate) Add(in telegraf.Metric) {
	m.cache.add(in)
}

// Push does the work of calculating the rate of the cached metrics
func (m *Rate) Push(acc telegraf.Accumulator) {
	for _, key := range m.cache.keys() {
		firstMetric := m.cache.first(key)
		lastMetric := m.cache.last(key)
		newFields := lastMetric.Copy().Fields()
		for _, field := range m.EventNames[firstMetric.Name()] {
			firstValue := firstMetric.Fields()[field]
			lastValue := lastMetric.Fields()[field]
			delete(newFields, field)
			switch firstValue.(type) {
			case int64:
				deltaT := lastMetric.Time().Sub(firstMetric.Time())
				deltaV := lastValue.(int64) - firstValue.(int64)
				if deltaT <= 0 {
					newFields[field+"_rate"] = 0
					continue
				}
				newValue := float64(deltaV) / float64(deltaT.Seconds())
				newFields[field+"_per_second"] = newValue
			case float64:
				deltaT := lastMetric.Time().Sub(firstMetric.Time())
				deltaV := lastValue.(float64) - firstValue.(float64)
				if deltaT <= 0 {
					newFields[field+"_rate"] = 0
					continue
				}
				newValue := deltaV / float64(deltaT.Seconds())
				newFields[field+"_per_second"] = newValue
			default:
				fmt.Printf("Got unhandled metric type: %s\n", reflect.TypeOf(firstValue).Kind())
			}
		}
		acc.AddFields(firstMetric.Name(), newFields, firstMetric.Tags())
	}
}

func init() {
	aggregators.Add("rate", func() telegraf.Aggregator {
		return NewRate()
	})
}
