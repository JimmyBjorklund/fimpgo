package primefimp

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

const (
	ComponentDevice   = "device"
	ComponentThing    = "thing"
	ComponentRoom     = "room"
	ComponentArea     = "area"
	ComponentHouse    = "house"
	ComponentHub      = "hub"
	ComponentShortcut = "shortcut"
	ComponentMode     = "mode"
	ComponentTimer    = "timer"

	CmdGet    = "get"
	CmdSet    = "set"
	CmdEdit   = "edit"
	CmdDelete = "delete"
	CmdAdd    = "add"
)

// Top level element for commands
type Request struct {
	Cmd       string       `json:"cmd"`
	Component interface{}  `json:"component"`
	Param     RequestParam `json:"param"`
	RequestID interface{}  `json:"requestId,omitempty"`
	Id        interface{}  `json:"id,omitempty"`
}

type RequestParam struct {
	Id         int      `json:"id,omitempty"`
	Components []string `json:"components,omitempty"`
}

type Fimp struct {
	Adapter string `json:"adapter"`
	Address string `json:"address"`
	Group   string `json:"group"`
}

type Client struct {
	Name string `json:"name"`
}

type Device struct {
	Fimp          Fimp                   `json:"fimp"`
	Client        Client                 `json:"client"`
	Functionality string                 `json:"functionality"`
	Service       map[string]Service     `json:"services"`
	ID            int                    `json:"id"`
	Lrn           bool                   `json:"lrn"`
	Model         string                 `json:"model"`
	Param         map[string]interface{} `json:"param"`
	Problem       bool                   `json:"problem"`
	Room          int                    `json:"room"`
	Changes       map[string]interface{} `json:"changes"`
	ThingID       int                    `json:"thing"`
}

type Thing struct {
	ID      int               `json:"id"`
	Address string            `json:"addr"`
	Name    string            `json:"name"`
	Devices []int             `json:"devices,omitempty"`
	Props   map[string]string `json:"props,omitempty"`
	RoomID  int               `json:"room"`
}

type House struct {
	Learning interface{} `json:"learning"`
	Mode     string      `json:"mode"`
	Time     time.Time   `json:"time"`
}

type Room struct {
	Alias   string     `json:"alias"`
	ID      int        `json:"id"`
	Param   RoomParams `json:"param"`
	Client  Client     `json:"client"`
	Type    string     `json:"type"`
	Area    int        `json:"area,omitempty"`
	Outside bool       `json:"outside"`
}

type RoomParams struct {
	Heating RoomHeating `json:"heating"`
}

type RoomHeating struct {
	Desired float64 `json:"desired"`
	Target  float64 `json:"target"`
}

type Service struct {
	Addr       string                 `json:"addr,omitempty"`
	Enabled    bool                   `json:"enabled,omitempty"`
	Interfaces []string               `json:"intf"`
	Props      map[string]interface{} `json:"props"`
}

type UserInfo struct {
	UID  string   `json:"uuid,omitempty"`
	Name UserName `json:"name,omitempty"`
}

type UserName struct {
	Fullname string `json:"fullname,omitempty"`
}

type Area struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ActionDevice map[string]interface{}

type ActionRoom map[string]interface{}

type ShortcutAction struct {
	Device map[int]ActionDevice `json:"device"`
	Room   map[int]ActionRoom   `json:"room"`
}

type Shortcut struct {
	ID     int            `json:"id"`
	Client Client         `json:"client"`
	Action ShortcutAction `json:"action"`
}

type HubMode struct {
	Current  string `json:"current"`
	Previous string `json:"prev"`
}

type Hub struct {
	Mode HubMode `json:"mode"`
}

type ModeAction struct {
	Device map[int]ActionDevice `json:"device"`
	Room   map[int]ActionRoom   `json:"room"`
}

type Mode struct {
	Id     string     `json:"id"`
	Action ModeAction `json:"action"`
}

type TimerAction struct {
	Type     string
	Shortcut int
	Mode     string
	Action   ShortcutAction
}

type Timer struct {
	Action  TimerAction
	Client  Client                 `json:"client"`
	Enabled bool                   `json:"enabled"`
	Time    map[string]interface{} `json:"time"`
	Id      int                    `json:"id"`
}

func (t *Timer) UnmarshalJSON(b []byte) error {
	temp := &struct {
		Action  interface{}
		Client  Client                 `json:"client"`
		Enabled bool                   `json:"enabled"`
		Time    map[string]interface{} `json:"time"`
		Id      int                    `json:"id"`
	}{}

	err := json.Unmarshal(b, temp)
	if err != nil {
		return err
	}
	t.Client = temp.Client
	t.Enabled = temp.Enabled
	t.Time = temp.Time
	t.Id = temp.Id

	switch temp.Action.(type) {
	case float64:
		t.Action.Type = "shortcut"
		t.Action.Shortcut = int(temp.Action.(float64))
	case float32:
		// If we are running on a 32 bit machine
		t.Action.Type = "shortcut"
		t.Action.Shortcut = int(temp.Action.(float32))
	case string:
		t.Action.Type = "mode"
		t.Action.Mode = temp.Action.(string)
	case map[string]interface{}:
		t.Action.Type = "custom"
		act := temp.Action.(map[string]interface{})
		if actRoom, ok := act["room"]; ok {
			t.Action.Action.Room = make(map[int]ActionRoom)
			for idRoom, act := range actRoom.(map[string]interface{}) {
				actTransposed, ok := act.(map[string]interface{})
				if !ok {
					continue
				}
				idRoom, err := strconv.Atoi(idRoom)
				if err != nil {
					return err
				}
				t.Action.Action.Room[idRoom] = actTransposed
			}
		}
		if actDevice, ok := act["device"]; ok {
			t.Action.Action.Device = make(map[int]ActionDevice)
			for idDevice, act := range actDevice.(map[string]interface{}) {
				actTransposed, ok := act.(map[string]interface{})
				if !ok {
					continue
				}
				idDev, err := strconv.Atoi(idDevice)
				if err != nil {
					return err
				}
				t.Action.Action.Device[idDev] = actTransposed
			}
		}
	default:
		return errors.New("invalid timer structure")
	}

	return nil
}
