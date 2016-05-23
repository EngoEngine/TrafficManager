package systems

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/luxengine/math"
)

const (
	MinTravelDistance = float32(24)
	commuterZIndex    = 50
)

type commuterEntityCity struct {
	*ecs.BasicEntity
	*CityComponent
}

type commuterEntityRoad struct {
	*ecs.BasicEntity
	*RoadComponent
	*common.SpaceComponent
}

type commuterEntityCommuter struct {
	*ecs.BasicEntity
	*CommuterComponent
	*common.SpaceComponent
}

type commuterEntityClock struct {
	*ecs.BasicEntity
	*TimeComponent
	*common.SpaceComponent
}

type CommuterSystem struct {
	world *ecs.World

	Vehicles map[string]Vehicle

	clock     commuterEntityClock
	cities    map[uint64]commuterEntityCity
	roads     map[uint64]commuterEntityRoad
	commuters []*Commuter

	// TODO: this may be another system?
	lastHour int
}

type Crash struct {
	ecs.BasicEntity
	common.AudioComponent
	common.SpaceComponent
}

func (c *CommuterSystem) New(w *ecs.World) {
	c.world = w
	c.cities = make(map[uint64]commuterEntityCity)
	c.roads = make(map[uint64]commuterEntityRoad)

}

func (c *CommuterSystem) Remove(basic ecs.BasicEntity) {
	delete(c.cities, basic.ID())
	delete(c.roads, basic.ID())

	var commuter *CommuterComponent
	for index, cmtr := range c.commuters {
		if cmtr.ID() == basic.ID() {
			commuter = &c.commuters[index].CommuterComponent
			c.commuters = append(c.commuters[:index], c.commuters[index+1:]...)
			break
		}
	}

	if commuter != nil {
		if commuter.Lane != nil {
			for index, cmtr := range commuter.Lane.Commuters {
				if cmtr.ID() == basic.ID() {
					commuter.Lane.Commuters = append(commuter.Lane.Commuters[:index], commuter.Lane.Commuters[index+1:]...)
					break
				}
			}
		}

		if commuter.NewLane != nil {
			for index, cmtr := range commuter.NewLane.Commuters {
				if cmtr.ID() == basic.ID() {
					commuter.NewLane.Commuters = append(commuter.NewLane.Commuters[:index], commuter.NewLane.Commuters[index+1:]...)
					break
				}
			}
		}
	}

	if basic.ID() == c.clock.ID() {
		c.clock = commuterEntityClock{}
	}
}

func (c *CommuterSystem) AddCity(basic *ecs.BasicEntity, city *CityComponent) {
	cec := commuterEntityCity{basic, city}
	c.cities[basic.ID()] = cec
}

func (c *CommuterSystem) AddRoad(basic *ecs.BasicEntity, road *RoadComponent, space *common.SpaceComponent) {
	c.roads[basic.ID()] = commuterEntityRoad{basic, road, space}
}

// SetClock sets the current clock to the given one
func (c *CommuterSystem) SetClock(basic *ecs.BasicEntity, clock *TimeComponent, space *common.SpaceComponent) {
	c.clock = commuterEntityClock{basic, clock, space}
}

func (c *CommuterSystem) Update(dt float32) {
	// Do all of these things once per gameSpeed level
	for i := float32(0); i < c.clock.Speed; i++ {
		c.commuterDispatch()
		c.commuterSpeed(dt)
		c.commuterLaneSwitching()
		c.commuterMove(dt)
		c.commuterArrival()
	}

	engo.SetTitle(fmt.Sprintf("%f", engo.Time.FPS()))
}

func (c *CommuterSystem) commuterSpeed(dt float32) {
	for _, road := range c.roads {
		for _, lane := range road.Lanes {
			for commIndex, comm := range lane.Commuters {
				if commIndex > 0 {
					// Someone is in front of us
					distance := lane.Commuters[commIndex-1].DistanceTravelled - comm.DistanceTravelled
					minCarDistance := ((comm.Speed/comm.BrakeSpeed)*(comm.Speed/2) + lane.Commuters[commIndex-1].Width)

					switch {
					case distance < minCarDistance:
						// Can we switch lanes?
						if !comm.SwitchingLane {
							leftBefore, rightBefore := c.canSwitch(road, lane, comm)

							if rightBefore >= -1 {
								comm.NewLane = road.Lanes[lane.Index+1]
								comm.SwitchingLane = true

								if rightBefore == -1 {
									comm.NewLane.Commuters = append(comm.NewLane.Commuters, comm)
								} else {
									comm.NewLane.Commuters = append(comm.NewLane.Commuters[:rightBefore], append([]*Commuter{comm}, comm.NewLane.Commuters[rightBefore:]...)...)
								}
							}

							if comm.NewLane == nil && leftBefore >= -1 {
								comm.NewLane = road.Lanes[lane.Index-1]
								comm.SwitchingLane = true

								if leftBefore == -1 {
									comm.NewLane.Commuters = append(comm.NewLane.Commuters, comm)
								} else {
									comm.NewLane.Commuters = append(comm.NewLane.Commuters[:leftBefore], append([]*Commuter{comm}, comm.NewLane.Commuters[leftBefore:]...)...)
								}
							}
						}

						// Hit the brakes! (at least until we're done moving? )
						comm.Speed -= comm.BrakeSpeed * dt

					case distance > minCarDistance:
						// Speed up if we want to
						if comm.Speed < comm.PreferredSpeed && !comm.SwitchingLane {
							comm.Speed += comm.AccelerationSpeed * dt
						}
					}
				} else {
					// We're all alone
					switch {
					case comm.Speed < comm.PreferredSpeed:
						if !comm.SwitchingLane {
							comm.Speed += comm.AccelerationSpeed * dt // TODO: not failsafe
						}
					case comm.Speed > comm.PreferredSpeed:
						comm.Speed -= comm.BrakeSpeed * dt
					}
				}
			}
		}
	}
}

// canSwitch indicates if you can move to the lane on the left (first one), and if you can to the left (second one).
// Values -2 mean you cannot move, values -1 mean you can move because there's no-one else, and other values indicate
// you have to move in front of that car
func (c *CommuterSystem) canSwitch(road commuterEntityRoad, lane *Lane, comm *Commuter) (int, int) {
	lr := [2]int{-1, 1}
	result := [2]int{-1, -1}

	for index, dir := range lr {
		if newIndex := lane.Index + 1*dir; newIndex == len(road.Lanes) || newIndex < 0 {
			result[index] = -2
			continue // lane does not exist
		}

		for rightIndex, rightCommuter := range road.Lanes[lane.Index+1*dir].Commuters {
			if rightCommuter.DistanceTravelled > comm.DistanceTravelled {
				// In front of us
				if rightCommuter.DistanceTravelled-rightCommuter.Width < comm.DistanceTravelled {
					// But only the front of it is, it's partly next to us: don't move!
					result[index] = -2
					break
				}
				if rightCommuter.DistanceTravelled-rightCommuter.Width-(comm.Speed/comm.BrakeSpeed)*(comm.Speed/2) < comm.DistanceTravelled {
					// We might bump into it, even though it's far away: don't move!
					result[index] = -2
					break
				}
			} else if rightCommuter.DistanceTravelled < comm.DistanceTravelled {
				// Behind us
				if comm.DistanceTravelled-comm.Width < rightCommuter.DistanceTravelled+(rightCommuter.Speed/rightCommuter.BrakeSpeed)*(rightCommuter.Speed/2) {
					// But only part of it is, it's partly next to us: don't move!
					result[index] = -2
					break
				}
				if result[index] == -1 {
					// First one that's completely behind us, so let's move in front of that one!
					result[index] = rightIndex
					break // since we can move
				}
			}
		}
	}

	return result[0], result[1]
}

func (c *CommuterSystem) commuterDispatch() {
	for _, city := range c.cities {
		for i, q := range city.Queue {
			if q.Amount == 0 {
				continue
			}

			// TODO: process q.Name

			var dispatched bool
		Outer:
			for _, road := range c.roads {
				if c.cities[road.To.ID()].Category == q.To && !dispatched {
					// We can dispatch! :)
					for _, lane := range road.Lanes {
						if lane.CanDispatch() {
							dispatched = true
							c.dispatch(c.newCommuter(city, c.Vehicles[q.Name]), road, lane)
							break Outer
						}
					}
				}
			}

			if dispatched {
				city.Queue[i].Amount--
			}
		}
	}
}

func (l *Lane) CanDispatch() bool {
	if len(l.Commuters) == 0 {
		return true
	}

	return l.Commuters[len(l.Commuters)-1].DistanceTravelled-l.Commuters[len(l.Commuters)-1].Width >= MinTravelDistance
}

func (c *CommuterSystem) commuterLaneSwitching() {
	for _, comm := range c.commuters {
		if !comm.SwitchingLane {
			continue // with other commuters
		}

		amount := float32(1)
		if comm.Lane.Index > comm.NewLane.Index {
			amount = -1
		}
		comm.SwitchingProgress += amount // todo: update speed

		// Move it graphically
		angle := (comm.Rotation / 180) * math.Pi
		dx := math.Sin(angle) * amount
		dy := math.Cos(angle) * amount
		comm.SpaceComponent.Position.X -= dx
		comm.SpaceComponent.Position.Y += dy

		if math.Abs(comm.SwitchingProgress) >= laneWidth {
			// Done switching, remove it from lane
			comm.Lane.Remove(comm.BasicEntity)
			comm.Lane = comm.NewLane
			comm.NewLane = nil
			comm.SwitchingLane = false
		}
	}
}

func (c *CommuterSystem) commuterMove(dt float32) {
	for _, road := range c.roads {
		var (
			alpha = (road.Rotation / 180) * math.Pi
			beta  = float32(0.5) * math.Pi
			gamma = 0.5*math.Pi - alpha

			sinAlpha = math.Sin(alpha)
			sinBeta  = math.Sin(beta)
			sinGamma = math.Sin(gamma)
		)

		var crashed []ecs.BasicEntity
		for _, lane := range road.Lanes {
			for commIndex, comm := range lane.Commuters {
				// Move the current speed
				newDistance := comm.Speed * dt
				comm.DistanceTravelled += newDistance

				if commIndex > 0 {
					if comm.DistanceTravelled > (lane.Commuters[commIndex-1].DistanceTravelled - lane.Commuters[commIndex-1].Width) {
						// Crash!

						crash := Crash{
							BasicEntity: ecs.NewBasic(),
							AudioComponent: common.AudioComponent{
								File:       "sfx/crash.wav",
								Repeat:     false,
								Background: false,
							},
							SpaceComponent: lane.Commuters[commIndex-1].SpaceComponent,
						}

						for _, system := range c.world.Systems() {
							switch sys := system.(type) {
							case *common.AudioSystem:
								sys.Add(&crash.BasicEntity, &crash.AudioComponent, &crash.SpaceComponent)
							}
						}

						fmt.Println("Crash", comm.ID(), lane.Commuters[commIndex-1].BasicEntity.ID())

						// Actual removal happens outside this loop
						crashed = append(crashed, comm.BasicEntity)
						crashed = append(crashed, lane.Commuters[commIndex-1].BasicEntity)

						// TODO; apply vehicle "Cost"
					}
				}

				// Using the Law of Sines, we now compute the dx (c) and dy (a)
				b_length := newDistance

				b_part := b_length / sinBeta
				a_length := sinAlpha * b_part
				c_length := sinGamma * b_part

				comm.Position.Y += a_length
				comm.Position.X += c_length
			}
		}

		// Remove crashed commuters
		for _, crash := range crashed {
			c.world.RemoveEntity(crash)
		}
	}
}

func (c *CommuterSystem) commuterArrival() {
	for _, road := range c.roads {
		for _, lane := range road.Lanes {
			for i := len(lane.Commuters) - 1; i >= 0; i-- {
				comm := lane.Commuters[i]

				if comm.DistanceTravelled > comm.Road.SpaceComponent.Width-15 {
					engo.Mailbox.Dispatch(CommuterArrivedMessage{&comm.CommuterComponent})

					comm.Lane.Remove(comm.BasicEntity)
					if comm.NewLane != nil {
						comm.NewLane.Remove(comm.BasicEntity)
					}

					comm.Road = commuterEntityRoad{}
					comm.Lane = nil
					comm.NewLane = nil
					comm.SwitchingLane = false
					comm.DistanceTravelled = 0

					for _, system := range c.world.Systems() {
						switch sys := system.(type) {
						case *common.RenderSystem:
							sys.Remove(comm.BasicEntity)
						}
					}
				}
			}
		}
	}
}

func (c *CommuterSystem) newCommuter(city commuterEntityCity, v Vehicle) *Commuter {
	cmtr := &Commuter{BasicEntity: ecs.NewBasic()}
	cmtr.CommuterComponent = CommuterComponent{
		PreferredSpeed: v.Minspeed + rand.Float32()*(v.Maxspeed-v.Minspeed), // 60 being minimum speed, 80 being the variation,
		DepartureTimes: []time.Duration{
			time.Hour*6 + time.Duration(rand.Intn(180))*time.Minute,
			time.Hour*16 + time.Duration(rand.Intn(240))*time.Minute,
		},
		Speed:             50, // coming from city
		AccelerationSpeed: v.Acceleration,
		BrakeSpeed:        v.Brakes,
	}
	cmtr.SpaceComponent = common.SpaceComponent{
		Width:  v.Length,
		Height: 6,
	}
	cmtr.RenderComponent = common.RenderComponent{
		Drawable: common.Rectangle{BorderWidth: 0.5, BorderColor: color.RGBA{128, 128, 128, 128}},
		Color:    color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255},
	}
	cmtr.Vehicle = v

	cmtr.SetZIndex(commuterZIndex)
	cmtr.SetShader(common.LegacyShader)
	c.commuters = append(c.commuters, cmtr)

	return cmtr
}

func (c *CommuterSystem) dispatch(cmtr *Commuter, road commuterEntityRoad, lane *Lane) {
	cmtr.Road = road
	cmtr.Lane = lane
	cmtr.Position = road.SpaceComponent.Position
	cmtr.Rotation = road.Rotation

	// Translate the commuter for the given lane (hopefully!) - TODO: this can be a Shader
	angle := (cmtr.Rotation / 180) * math.Pi
	lanewidth := float32(lane.Index) * laneWidth
	dx := math.Sin(angle) * (lanewidth + 2) // 2 == (laneHeight - carHeight)/2
	dy := math.Cos(angle) * (lanewidth + 2) // 2 == (laneHeight - carHeight)/2
	cmtr.SpaceComponent.Position.X -= dx
	cmtr.SpaceComponent.Position.Y += dy

	// Add it to the system to make visual
	for _, system := range c.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&cmtr.BasicEntity, &cmtr.RenderComponent, &cmtr.SpaceComponent)
		}
	}

	lane.Commuters = append(lane.Commuters, cmtr)
}

type CommuterArrivedMessage struct {
	Commuter *CommuterComponent
}

func (CommuterArrivedMessage) Type() string { return "CommuterArrivedMessage" }
