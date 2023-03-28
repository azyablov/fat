package methods

import "fmt"

type EnumMethods int

const (
	_               = iota
	GET EnumMethods = iota + 1
	SET
	CLI
	VALIDATE
)

type Method struct {
	Method string `json:"method"`
}

func (m *Method) GetMethod() (string, error) {
	switch m.Method {
	case "get":
		break
	case "set":
		break
	case "cli":
		break
	case "validate":
		break
	default:
		return "", fmt.Errorf("method isn't set properly, while should be GET / SET / CLI / VALIDATE")
	}
	return m.Method, nil
}

func (m *Method) SetMethod(rm EnumMethods) error {
	switch rm {
	case GET:
		m.Method = "get"
		break
	case SET:
		m.Method = "set"
		break
	case CLI:
		m.Method = "cli"
		break
	case VALIDATE:
		m.Method = "validate"
		break
	default:
		return fmt.Errorf("method provided isn't correct, while should be GET / SET / CLI / VALIDATE")
	}
	return nil
}
