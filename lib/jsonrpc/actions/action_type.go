package actions

import (
	"fmt"
)

type EnumActions int

const (
	_                   = iota
	REPLACE EnumActions = iota + 1
	UPDATE
	DELETE
)

type Action struct {
	Action string `json:"action"`
}

func (a *Action) GetAction() (string, error) {
	switch a.Action {
	case "replace":
		break
	case "update":
		break
	case "delete":
		break
	default:
		return "", fmt.Errorf("action isn't set properly, while should be REPLACE / UPDATE / DELETE")
	}
	return a.Action, nil
}

func (a *Action) SetAction(ra EnumActions) error {
	switch ra {
	case DELETE:
		a.Action = "delete"
		break
	case REPLACE:
		a.Action = "replace"
		break
	case UPDATE:
		a.Action = "update"
		break
	default:
		return fmt.Errorf("action provided isn't correct, while should be REPLACE / UPDATE / DELETE")
	}
	return nil
}
