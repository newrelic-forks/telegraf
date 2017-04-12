package nrinsights

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	insights "source.datanerd.us/tools/go-insights-v2/client"
)

type Insights struct {
	LicenseKey string
	AccountId  string
	EventType  string
	client     *insights.InsertClient
}

func NewInsights() *Insights {
	return &Insights{}
}

func (n *Insights) Connect() error {
	n.client = insights.NewInsertClient(n.LicenseKey, n.AccountId)
	if err := n.client.Validate(); err != nil {
		return fmt.Errorf("Error validating InsertClient: %s", err.Error())
	}
	if err := n.client.Start(); err != nil {
		return fmt.Errorf("Error starting InsertClient: %s", err.Error())
	}
	return nil
}

func (n *Insights) Close() error {
	return n.client.Flush()
}

func (n *Insights) Description() string {
	//TODO Write description
	return ""
}

func (n *Insights) SampleConfig() string {
	//TODO sample config
	return ""
}

func (n *Insights) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		nrMetric := map[string]interface{}{}
		for name, value := range metric.Fields() {
			nrMetric[name] = value
		}
		for name, value := range metric.Tags() {
			nrMetric[name] = value
		}
		nrMetric["eventType"] = metric.Name()
		nrMetric["timestamp"] = metric.Time().Unix()
		if err := n.client.EnqueueEvent(nrMetric); err != nil {
			return fmt.Errorf("Error enqueuing insights metric: %s", err.Error())
		}
	}
	return nil
}

func init() {
	outputs.Add("nr_insights", func() telegraf.Output {
		return NewInsights()
	})
}
