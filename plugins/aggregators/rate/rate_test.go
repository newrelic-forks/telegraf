package rate

import (
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var m1, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a": int64(1),
		"b": int64(1),
		"c": int64(1),
		"d": int64(1),
		"e": int64(1),
		"f": float64(1),
		"g": float64(1),
		"h": float64(2),
		"i": float64(2),
		"j": float64(3),
	},
	time.Unix(0, 0),
)
var m2, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a":        int64(1),
		"b":        int64(3),
		"c":        int64(3),
		"d":        int64(3),
		"e":        int64(3),
		"f":        float64(240),
		"g":        float64(240),
		"h":        float64(1),
		"i":        float64(1),
		"j":        float64(1),
		"k":        float64(200),
		"ignoreme": "string",
		"andme":    true,
	},
	time.Unix(120, 0),
)
var m3, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a":        int64(2),
		"b":        int64(4),
		"c":        int64(5),
		"d":        int64(10),
		"e":        int64(15),
		"f":        float64(212),
		"g":        float64(2645),
		"h":        float64(11),
		"i":        float64(111),
		"j":        float64(1665),
		"k":        float64(211),
		"ignoreme": "string",
		"andme":    true,
	},
	time.Unix(120, 0),
)

func TestCache(t *testing.T) {
	c := newCache()
	c.add(m1)
	c.add(m2)
	c.add(m3)
	keys := c.keys()
	expectedKeys := []uint64{m1.HashID()}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("Expect %v to equal %v", keys, expectedKeys)
	}

	first := c.first(m1.HashID())
	if !reflect.DeepEqual(first, m1) {
		t.Errorf("Expected %v to equal %v", first, m1)
	}

	last := c.last(m1.HashID())
	if !reflect.DeepEqual(last, m3) {
		t.Errorf("Expected %v to equal %v", first, m3)
	}

	c.clear()
	storeLen := len(c.store)
	if storeLen != 0 {
		t.Errorf("expected store to have 0 len, got: %d", storeLen)
	}

}

func TestRateWithPeriod(t *testing.T) {
	expectedFields := map[string]interface{}{
		"a_per_second": float64(0.008333333333333333),
		"andme":        bool(true),
		"b_per_second": float64(0.025),
		"c":            int64(5),
		"d":            int64(10),
		"e":            int64(15),
		"f_per_second": float64(1.7583333333333333),
		"g_per_second": float64(22.033333333333335),
		"h":            float64(11),
		"i":            float64(111),
		"ignoreme":     string("string"),
		"j":            float64(1665),
		"k":            float64(211),
	}

	expectedTags := map[string]string{
		"foo": "bar",
	}

	acc := testutil.Accumulator{}
	rate := NewRate()
	rate.(*Rate).EventNames = map[string][]string{
		"m1": []string{"a", "b", "f", "g"},
	}

	rate.Add(m1)
	rate.Add(m2)
	rate.Add(m3)
	rate.Push(&acc)

	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)

}
