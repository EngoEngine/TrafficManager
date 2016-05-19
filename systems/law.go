package systems

import (
	"fmt"
	"log"
	"time"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
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
	// RoadID is the ID() of the BasicEntity of the Road
	RoadID uint64
	// Position is the amount of units measured from the `From` City.
	Position float32
}

type lawEntityRoad struct {
	*ecs.BasicEntity
	*RoadComponent
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
	laws       map[LawType]float32

	// lastFine is the moment in game-time where the commuter was last fined for a given law
	lastFine map[LawType]map[uint64]time.Time

	roads map[uint64]lawEntityRoad
	cp    []lawEntityCheckpoint
}

func (l *LawSystem) New(w *ecs.World) {
	l.world = w
	l.laws = make(map[LawType]float32)
	l.roads = make(map[uint64]lawEntityRoad)
	l.lastFine = make(map[LawType]map[uint64]time.Time)

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

	// TODO; default laws?
	l.laws[LawMaxSpeed] = 120

	// Listening for LawMessages allows us to update the law
	engo.Mailbox.Listen("LawMessage", func(m engo.Message) {
		lm, ok := m.(LawMessage)
		if !ok {
			return
		}

		l.laws[lm.Law.Type] = lm.Law.Value
	})
}

func (l *LawSystem) AddCheckpoint(basic *ecs.BasicEntity, law *LawComponent, roadloc *RoadLocationComponent) {
	l.cp = append(l.cp, lawEntityCheckpoint{basic, law, roadloc})
}

func (l *LawSystem) AddRoad(basic *ecs.BasicEntity, road *RoadComponent) {
	l.roads[basic.ID()] = lawEntityRoad{basic, road}
}

func (l *LawSystem) Remove(basic ecs.BasicEntity) {
	del := -1
	for index, e := range l.cp {
		if e.BasicEntity.ID() == basic.ID() {
			del = index
			break
		}
	}
	if del >= 0 {
		l.cp = append(l.cp[:del], l.cp[del+1:]...)
	}

	delete(l.roads, basic.ID())
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
	road := l.roads[loc.RoadID]
	for _, lane := range road.Lanes {
		for _, comm := range lane.Commuters {
			if comm.Speed > l.laws[LawMaxSpeed] {
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
	// TODO: l.cmtrSystem.clock.TimeComponent?! It should also use the clock component!
	if !ok || l.cmtrSystem.clock.Time.Sub(cmtr).Hours() > 1 {
		cmtrs[comm.ID()] = l.cmtrSystem.clock.Time
		fmt.Println("Commuter", comm.ID(), "has been fined")
	}
}

type SpeedCheckpoint struct {
	ecs.BasicEntity
	LawComponent
	RoadLocationComponent
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	common.AudioComponent

	icon IconComponent
}

type IconComponent struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}
