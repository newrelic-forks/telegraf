package system

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type EventStore struct {
	sync.Mutex
	List []*docker.APIEvents
}

func (es *EventStore) AddEvent(event *docker.APIEvents) {
	es.Lock()
	defer es.Unlock()
	es.List = append(es.List, event)
}

func (es *EventStore) GetAllEvents() []*docker.APIEvents {
	es.Lock()
	defer func() {
		es.List = []*docker.APIEvents{}
		es.Unlock()
	}()
	return es.List
}

type DockerClient interface {
	AddEventListener(listener chan<- *docker.APIEvents) error
}

// Docker object
type DockerEvents struct {
	Endpoint       string
	ContainerNames []string
	Timeout        internal.Duration
	Total          bool
	engine_host    string
	client         DockerClient
	EventStore     EventStore
}

// Description returns input description
func (d *DockerEvents) Description() string {
	return "Read events from Docker event endpoint"
}

// SampleConfig prints sampleConfig
func (d *DockerEvents) SampleConfig() string { return "" }

func (d *DockerEvents) Gather(acc telegraf.Accumulator) error {
	var endpoint = "/var/run/docker.sock"
	var err error
	if d.client == nil {
		if d.Endpoint != "" {
			endpoint = d.Endpoint
		}
		d.client, err = docker.NewClient(endpoint)
		if err != nil {
			return fmt.Errorf("Error initializing Docker client: %s", err.Error())
		}

		eventChan := make(chan *docker.APIEvents)
		if err := d.client.AddEventListener(eventChan); err != nil {
			return fmt.Errorf("Error adding Docker event listener: %s", err.Error())
		}
		go func() {
			for {
				d.EventStore.AddEvent(<-eventChan)
			}
		}()
	}

	for _, event := range d.EventStore.GetAllEvents() {
		tags := map[string]string{}
		tags["id"] = event.Actor.ID
		for k, v := range event.Actor.Attributes {
			tags[k] = v
		}
		eventFields := map[string]interface{}{
			"status": event.Status,
			"id":     event.ID,
			"from":   event.From,
			"type":   event.Type,
			"action": event.Action,
		}
		acc.AddFields("docker_event", eventFields, tags, time.Unix(event.Time, 0))
	}
	return nil
}

func init() {
	inputs.Add("nr_docker_events", func() telegraf.Input {
		return &DockerEvents{
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
