package system

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/influxdata/telegraf/testutil"
)

var event1 = &docker.APIEvents{
	Action: "pull",
	Type:   "image",
	Actor: docker.APIActor{
		ID: "redis:latest",
		Attributes: map[string]string{
			"name": "redis",
		},
	},
	Status:   "pull",
	ID:       "redis:latest",
	From:     "",
	Time:     1491941160,
	TimeNano: 1491941160813293396,
}

var event2 = &docker.APIEvents{
	Action: "create",
	Type:   "container",
	Actor: docker.APIActor{
		ID: "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
		Attributes: map[string]string{
			"image": "redis",
			"name":  "awesome_shockley",
		},
	},
	Status:   "create",
	ID:       "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
	From:     "redis",
	Time:     1491941161,
	TimeNano: 1491941161642448153,
}

var event3 = &docker.APIEvents{
	Action: "attach",
	Type:   "container",
	Actor: docker.APIActor{
		ID: "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
		Attributes: map[string]string{
			"image": "redis",
			"name":  "awesome_shockley",
		},
	},
	Status:   "attach",
	ID:       "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
	From:     "redis",
	Time:     1491941161,
	TimeNano: 1491941161644831384,
}

var event4 = &docker.APIEvents{
	Action: "connect",
	Type:   "network",
	Actor: docker.APIActor{
		ID: "f3e873d1bf0aff5e84ce078f052471fc4ee48e561cac3652dfa9729bf1eaf56b",
		Attributes: map[string]string{
			"container": "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
			"name":      "bridge",
			"type":      "bridge",
		},
	},
	Status:   "network:connect",
	ID:       "f3e873d1bf0aff5e84ce078f052471fc4ee48e561cac3652dfa9729bf1eaf56b",
	From:     "",
	Time:     1491941161,
	TimeNano: 1491941161663042459,
}

var event5 = &docker.APIEvents{
	Action: "mount",
	Type:   "volume",
	Actor: docker.APIActor{
		ID: "14b901013ece150d80146074c80991633a6eea391d11d2259a9e32fac8a38c1d",
		Attributes: map[string]string{
			"container":   "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
			"destination": "/data",
			"driver":      "local",
			"propagation": "",
			"read/write":  "true",
		},
	},
	Status:   "volume:mount",
	ID:       "14b901013ece150d80146074c80991633a6eea391d11d2259a9e32fac8a38c1d",
	From:     "",
	Time:     1491941161,
	TimeNano: 1491941161665132043,
}

var event6 = &docker.APIEvents{
	Action: "start",
	Type:   "container",
	Actor: docker.APIActor{
		ID: "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
		Attributes: map[string]string{
			"image": "redis",
			"name":  "awesome_shockley",
		},
	},
	Status:   "start",
	ID:       "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
	From:     "redis",
	Time:     1491941161,
	TimeNano: 1491941161974897108,
}

var event7 = &docker.APIEvents{
	Action: "resize",
	Type:   "container",
	Actor: docker.APIActor{
		ID: "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
		Attributes: map[string]string{
			"height": "74",
			"image":  "redis",
			"name":   "awesome_shockley",
			"width":  "253",
		},
	},
	Status:   "resize",
	ID:       "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
	From:     "redis",
	Time:     1491941161,
	TimeNano: 1491941161977609343,
}

type FakeDockerClient struct {
	events []*docker.APIEvents
}

func NewFakeDocker() DockerClient {
	d := FakeDockerClient{
		events: []*docker.APIEvents{event1, event2, event3, event4, event5, event6, event7},
	}
	return &d
}

func (d *FakeDockerClient) AddEventListener(listener chan<- *docker.APIEvents) error {
	go func() {
		for _, event := range d.events {
			listener <- event
		}
		close(listener)
	}()
	return nil
}

func TestGather(t *testing.T) {
	dc := NewFakeDocker()
	var acc testutil.Accumulator
	dockerEvents := DockerEvents{
		client: dc,
		EventStore: EventStore{
			List: []*docker.APIEvents{event1, event2, event3, event4, event5, event6, event7},
		},
	}
	if err := dockerEvents.Gather(&acc); err != nil {
		t.Error(err.Error())
	}
	acc.AssertContainsTaggedFields(t,
		"docker_event",
		map[string]interface{}{
			"status": "create",
			"id":     "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
			"from":   "redis",
			"type":   "container",
			"action": "create",
		},
		map[string]string{
			"image": "redis",
			"name":  "awesome_shockley",
			"id":    "1299aba1255b7e621a156fa61af5332c4567731ac9d6a6376e41b48d233ccd46",
		},
	)
}
