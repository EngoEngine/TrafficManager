package systems

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"fmt"
	"log"
	"time"
)

type LawType uint8

const (
	LawMinSpeed LawType = iota
	LawMaxSpeed
)

type Law struct {
	Type  LawType
	Value float32
}

func (l Law) Bool() bool {
	return l.Value != 0
}

// A LawMessage can be sent to update the current laws
type LawMessage struct {
	Law Law
}

func (LawMessage) Type() string {
	return "LawMessage"
}

type LawComponent struct {
	Name      string
	ChecksFor []LawType
}

type RoadLocationComponent struct {
	Road *Road
	// Position is the amount of units measuerd from the `From` City.
	Position float32
}

type lawEntityCheckpoint struct {
	*ecs.BasicEntity
	*LawComponent
	*RoadLocationComponent
}

// The LawSsytem handles things such as speed cameras
type LawSystem struct {
	world *ecs.World

	cmtrSystem *CommuterSystem

	laws map[LawType]float32

	// lastFine is the moment in game-time where the commuter was last fined for a given law
	lastFine map[LawType]map[uint64]time.Time

	cp []lawEntityCheckpoint
}

func (l *LawSystem) New(w *ecs.World) {
	l.world = w
	l.laws = make(map[LawType]float32)

	// Find reference for CommuterSystem
	for _, system := range l.world.Systems() {
		switch sys := system.(type) {
		case *CommuterSystem:
			l.cmtrSystem = sys
		}
	}

	if l.cmtrSystem == nil {
		log.Println("[FATAL] [LawSystem] `CommuterSystem` was not found. Did you add it?")
		return
	}

	// Listening for LawMessages allows us to update the law
	engo.Mailbox.Listen("LawMessage", func(m engo.Message) {
		lm, ok := m.(LawMessage)
		if !ok {
			return
		}

		l.laws[lm.Law.Type] = lm.Law.Value
	})
}

func (l *LawSystem) AddCheckpoint(basic *ecs.BasicEntity) {

}

func (l *LawSystem) Remove(ecs.BasicEntity) {

}

func (l *LawSystem) Update(dt float32) {
	for _, cp := range l.cp {
		for _, cf := range cp.ChecksFor {
			switch cf {
			case LawMinSpeed:

			case LawMaxSpeed:
				l.checkMaxSpeed(cp.RoadLocationComponent)
			}
		}
	}
}

func (l *LawSystem) checkMaxSpeed(loc *RoadLocationComponent) {
	for _, lane := range loc.Road.Lanes {
		for _, comm := range lane.Commuters {
			if comm.Speed > l.laws[LawMaxSpeed] {
				fmt.Println(comm.ID(), "is speeding")
				l.fine(LawMaxSpeed, comm)
			}
		}
	}
}

// fine fines the commuter for breaking a given LawType (assumes Commuter has broken law) - iff the Commuter wasn't
// just fined of the same law in the same minute.
func (l *LawSystem) fine(t LawType, comm *Commuter) {
	cmtrs, ok := l.lastFine[t]
	if !ok {
		l.lastFine[t] = make(map[uint64]time.Time)
		cmtrs = l.lastFine[t]
	}

	cmtr, ok := cmtrs[comm.ID()]
	if !ok || l.cmtrSystem.gameTime.Sub(cmtr).Minutes() > 1 {
		cmtrs[comm.ID()] = l.cmtrSystem.gameTime
		fmt.Println(comm.ID(), "has been fined")
	}
}

type SpeedCheckpoint struct {
	ecs.BasicEntity
	LawComponent
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	common.AudioComponent
}
