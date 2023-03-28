package datastores

import "fmt"

type EnumDatastores int

const (
	CANDIDATE EnumDatastores = iota
	RUNNING
	STATE
	TOOLS
)

type Datastore struct {
	Datastore string `json:"datastore,omitempty"`
}

func (d *Datastore) GetDatastore() (string, error) {
	switch d.Datastore {
	case "candidate":
		break
	case "running":
		break
	case "state":
		break
	case "tools":
		break
	default:
		return "", fmt.Errorf("datastore isn't set properly, while should be CANDIDATE / RUNNING / STATE / TOOLS")
	}
	return d.Datastore, nil
}

func (d *Datastore) SetDatastore(rd EnumDatastores) error {
	switch rd {
	case CANDIDATE:
		d.Datastore = "candidate"
		break
	case RUNNING:
		d.Datastore = "running"
		break
	case STATE:
		d.Datastore = "state"
		break
	case TOOLS:
		d.Datastore = "tools"
		break
	default:
		return fmt.Errorf("datastore provided isn't correct, while should be CANDIDATE / RUNNING / STATE / TOOLS")
	}
	return nil
}
