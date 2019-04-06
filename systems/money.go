package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
)

// CityType keeps track of the type of city
type CityType int

const (
	// CityTypeNew is a brand new city
	CityTypeNew = iota
	// CityTypeTown is a town, the lowest level
	CityTypeTown
	// CityTypeCity is a city, the moderate city type
	CityTypeCity
	// CityTypeMetro is a metro area, the largest city type
	CityTypeMetro
)

// CityUpdateMessage updates the city types when sent from Old to New
type CityUpdateMessage struct {
	Old, New CityType
}

// CityUpdateMessageType is the type of the CityUpdateMessage
const CityUpdateMessageType string = "CityUpdateMessage"

// Type implements the engo.Message interface
func (CityUpdateMessage) Type() string {
	return CityUpdateMessageType
}

// AddOfficerMessage tells the system to add an officer
type AddOfficerMessage struct{}

// AddOfficerMessageType is the type of an AddOfficerMessage
const AddOfficerMessageType string = "AddOfficerMessage"

// Type implements the engo.Message interface
func (AddOfficerMessage) Type() string {
	return AddOfficerMessageType
}

// MoneySystem keeps track of money available to the player
type MoneySystem struct {
	amount                int
	towns, cities, metros int
	officers              int
	elapsed               float32
}

// New listens to messages to update the number of cities and police in the game.
func (m *MoneySystem) New(w *ecs.World) {
	engo.Mailbox.Listen(CityUpdateMessageType, func(msg engo.Message) {
		upd, ok := msg.(CityUpdateMessage)
		if !ok {
			return
		}
		switch upd.New {
		case CityTypeNew:
			m.towns++
		case CityTypeTown:
			m.towns++
			if upd.Old == CityTypeTown {
				m.towns--
			} else if upd.Old == CityTypeCity {
				m.cities--
			} else if upd.Old == CityTypeMetro {
				m.metros--
			}
		case CityTypeCity:
			m.cities++
			if upd.Old == CityTypeTown {
				m.towns--
			} else if upd.Old == CityTypeCity {
				m.cities--
			} else if upd.Old == CityTypeMetro {
				m.metros--
			}
		case CityTypeMetro:
			m.metros++
			if upd.Old == CityTypeTown {
				m.towns--
			} else if upd.Old == CityTypeCity {
				m.cities--
			} else if upd.Old == CityTypeMetro {
				m.metros--
			}
		}
	})

	engo.Mailbox.Listen(AddOfficerMessageType, func(engo.Message) {
		m.officers++
	})
}

// Update keeps track of how much time has passed since the last addtion of money.
// When enough time passes, it adds money based on the number and type of cities
// and subtracts money based on the size of the police force employed.
func (m *MoneySystem) Update(dt float32) {
	m.elapsed += dt
	if m.elapsed > 10 {
		m.amount += m.towns*100 + m.cities*500 + m.metros*1000
		m.amount -= m.officers * 20
		engo.Mailbox.Dispatch(HUDMoneyMessage{
			Amount: m.amount,
		})
		m.elapsed = 0
	}
}

// Remove doesn't do anything since the system has no entities.
func (m *MoneySystem) Remove(b ecs.BasicEntity) {}
