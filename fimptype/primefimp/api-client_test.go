package primefimp

import (
	"strings"
	"testing"

	"github.com/futurehomeno/fimpgo"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var brokerUrl = "tcp://dev-sdu-sm-beta.local:1884"
var brokerUser = "simsek"
var brokerPass = "SivErAmEtOnyRIDIci"

func TestPrimeFimp_ClientApi_Update(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	uuid := uuid.New().String()
	validClientID := strings.ReplaceAll(uuid, "-", "")[0:22]

	mqtt := fimpgo.NewMqttTransport(brokerUrl, validClientID, brokerUser, brokerPass, true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	client := NewApiClient("test-1", mqtt, false)
	site, err := client.GetSite(false)
	if err != nil {
		t.Error("Error", err)
		t.Fail()
	}

	if len(site.Devices) == 0 {
		t.Error("Site should have more then 0 devices ")
	}
	log.Infof("Site contains %d devices", len(site.Devices))
	client.Stop()
}

func TestPrimeFimp_ClientApi_Notify(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	validClientID := strings.ReplaceAll(uuid.New().String(), "-", "")[0:22]

	mqtt := fimpgo.NewMqttTransport(brokerUrl, validClientID, brokerUser, brokerPass, true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	// Actual test
	notifyCh := make(chan Notify, 10)
	apiclientid := uuid.New().String()[0:12]
	client := NewApiClient(apiclientid, mqtt, true) // (clientID string, mqttTransport *fimpgo.MqttTransport, isCacheEnabled bool)
	channelID := uuid.New().String()[0:12]
	// Using "RegisterChannel" will send a message to our notify channel for all messages
	// If you want to use filters, check "RegisterChannelWithFilter"
	client.RegisterChannel(channelID, notifyCh) // (channelId string, ch chan Notify)
	client.StartNotifyRouter()
	// Notify router is started. Now please, make 3 "add", "edit" or "delete" actions to finalize the test.
	i := 0
	limit := 3
	for {
		select {
		case msg := <-notifyCh:
			log.Infof("Check %d/%d: New notify message of cmd = %s,comp = %s", i, limit, msg.Cmd, msg.Component)
			i++
			if i > limit {
				client.Stop()
				break
			}
		}
	}
}

func TestPrimeFimp_ClientApi_Notify_With_Filter(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	validClientID := strings.ReplaceAll(uuid.New().String(), "-", "")[0:22]

	mqtt := fimpgo.NewMqttTransport(brokerUrl, validClientID, brokerUser, brokerPass, true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	// Actual test
	channelIDAdd := uuid.New().String()[0:12]
	notifyAreaAdd := make(chan Notify, 10)
	notifyfilterAreaAdd := NotifyFilter{Cmd: CmdAdd, Component: ComponentArea}

	channelIDDelete := uuid.New().String()[0:12]
	notifyAreaDelete := make(chan Notify, 10)
	notifyfilterAreaDelete := NotifyFilter{Cmd: CmdDelete, Component: ComponentArea}

	channelIDEdit := uuid.New().String()[0:12]
	notifyAreaEdit := make(chan Notify, 10)
	notifyfilterAreaEdit := NotifyFilter{Cmd: CmdEdit, Component: ComponentArea}

	apiclientid := uuid.New().String()[0:12]
	client := NewApiClient(apiclientid, mqtt, true)                                             // (clientID string, mqttTransport *fimpgo.MqttTransport, isCacheEnabled bool)
	client.RegisterChannelWithFilter(channelIDAdd, notifyAreaAdd, notifyfilterAreaAdd)          // (channelId string, ch chan Notify, filter NotifyFilter)
	client.RegisterChannelWithFilter(channelIDDelete, notifyAreaDelete, notifyfilterAreaDelete) // (channelId string, ch chan Notify, filter NotifyFilter)
	client.RegisterChannelWithFilter(channelIDEdit, notifyAreaEdit, notifyfilterAreaEdit)       // (channelId string, ch chan Notify, filter NotifyFilter)
	client.StartNotifyRouter()
	// We started the channel with filter now let's add an area, edit the name and then delete it to finalize the test.
	addarea := 0
	deletearea := 0
	editarea := 0

	closeChan := make(chan string)
	go func() {
		for {
			select {
			case msg := <-notifyAreaAdd:
				addarea++
				log.Infof("Check %s: New notify message of cmd = %s,comp = %s", msg.Cmd, msg.Cmd, msg.Component)
				if addarea > 0 && deletearea > 0 && editarea > 0 {
					client.Stop()
					closeChan <- "shit"
					break
				}
			case msg := <-notifyAreaDelete:
				deletearea++
				log.Infof("Check %s: New notify message of cmd = %s,comp = %s", msg.Cmd, msg.Cmd, msg.Component)
				if addarea > 0 && deletearea > 0 && editarea > 0 {
					client.Stop()
					closeChan <- "shit"
					break
				}
			case msg := <-notifyAreaEdit:
				editarea++
				log.Infof("Check %s: New notify message of cmd = %s,comp = %s", msg.Cmd, msg.Cmd, msg.Component)
				if addarea > 0 && deletearea > 0 && editarea > 0 {
					client.Stop()
					closeChan <- "shit"
					break
				}
			}
		}
	}()

	<-closeChan
	t.Log("Tadaaa")
}
