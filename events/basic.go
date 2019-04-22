package events

import (
	"encoding/json"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func NewBasicEvent() (data []byte, err error) {
	timestamp := MakeTimestamp()
	deviceName := "RandomDevice-1"
	evt := models.Event{ Created:timestamp, Modified:timestamp, Device:deviceName }
	readings := []models.Reading{}
	readings = append(readings, models.Reading{Created:timestamp, Modified:timestamp, Device:deviceName, Name:"Reading1", Value:"ABC"})
	readings = append(readings, models.Reading{Created:timestamp, Modified:timestamp, Device:deviceName, Name:"Reading2", Value:"123"})
	evt.Readings = readings
	data, err = json.Marshal(evt)
	if err != nil {
		return
	}
	return
}
