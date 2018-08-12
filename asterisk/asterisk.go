package asterisk

import (
	"html/template"
	"log"
	"fmt"
	"errors"
	"bytes"
	"io/ioutil"
)

type Config struct {
	Outgoing string
	Template string
	Phones struct {
		Pyaterochka string
		Middle string
		Ryabinoviy string
	}
}

type AsteriskCall struct {
	Phone string
}

var cfg *Config
var tpl *template.Template

const GATE_PYATEROCHKA int64 = 1
const GATE_MIDDLE int64 = 2
const GATE_RYBINOVIY int64 = 3

func SetConfig(config *Config) {
	cfg = config
	tpl = template.New("outgoing")

	_, err := tpl.Parse(cfg.Template)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot parse config: %s", err))
	}
}

func Open(gate int64) (bool, error) {
	call := AsteriskCall{}

	switch gate {
	case GATE_PYATEROCHKA:
		call.Phone = cfg.Phones.Pyaterochka
	case GATE_MIDDLE:
		call.Phone = cfg.Phones.Middle
	case GATE_RYBINOVIY:
		call.Phone = cfg.Phones.Ryabinoviy
	default:
		return false, errors.New(fmt.Sprintf("Unknown gate: %d", gate))
	}

	createOutgoingCall(&call)

	return true, nil
}

func createOutgoingCall(call *AsteriskCall) (bool, error) {
	content := bytes.Buffer{}

	err := tpl.Execute(&content, call)
	if err != nil {
		log.Printf("Cannot create asterisk call file: %s", err)
		return false, err
	}

	tmpFile, err := ioutil.TempFile(cfg.Outgoing, "call_")
	tmpFile.Chmod(0666)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpFile.Write(content.Bytes()); err != nil {
		log.Fatal(err)
	}

	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	return true, nil
}