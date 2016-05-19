package systems

import (
	"fmt"
	"image/color"
	"math/rand"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/luxengine/math"
)

var (
	cashPerUnit float32 = 0.05
)

const (
	MinTravelDistance = float32(24)
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

	clock     commuterEntityClock
	cities    map[uint64]commuterEntityCity
	roads     map[uint64]commuterEntityRoad
	commuters []commuterEntityCommuter
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

	delete := -1
	var commuter *CommuterComponent
	for index, e := range c.commuters {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			commuter = c.commuters[index].CommuterComponent
			break
		}
	}

	if commuter != nil {
		for index, cmtr := range commuter.Lane.Commuters {
			if cmtr.ID() == basic.ID() {
				commuter.Lane.Commuters = append(commuter.Lane.Commuters[:index], commuter.Lane.Commuters[index+1:]...)
				break
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

	if delete >= 0 {
		c.commuters = append(c.commuters[:delete], c.commuters[delete+1:]...)
	}

	if basic.ID() == c.clock.ID() {
		c.clock = commuterEntityClock{}
	}
}

func (c *CommuterSystem) AddCity(basic *ecs.BasicEntity, city *CityComponent) {
	c.cities[basic.ID()] = commuterEntityCity{basic, city}
}

func (c *CommuterSystem) AddRoad(basic *ecs.BasicEntity, road *RoadComponent, space *common.SpaceComponent) {
	c.roads[basic.ID()] = commuterEntityRoad{basic, road, space}
}

func (c *CommuterSystem) AddCommuter(basic *ecs.BasicEntity, comm *CommuterComponent, space *common.SpaceComponent) {
	c.commuters = append(c.commuters, commuterEntityCommuter{basic, comm, space})
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
							hasRight := lane.Index != len(road.Lanes)-1
							hasLeft := lane.Index != 0

							if hasRight {
								canMove := true // by default
								before := -1
								for rightIndex, rightCommuter := range road.Lanes[lane.Index+1].Commuters {
									if rightCommuter.DistanceTravelled > comm.DistanceTravelled {
										// In front of us
										if rightCommuter.DistanceTravelled-rightCommuter.Width < comm.DistanceTravelled {
											// But only the front of it is, it's partly next to us: don't move!
											canMove = false
											break // since we dont move
										}
										if rightCommuter.DistanceTravelled-rightCommuter.Width-(comm.Speed/comm.BrakeSpeed)*(comm.Speed/2) < comm.DistanceTravelled {
											// We might bump into it, even though it's far away
											canMove = false
											break // since we dont move
										}
									} else if rightCommuter.DistanceTravelled < comm.DistanceTravelled {
										// Behind us
										if comm.DistanceTravelled-comm.Width < rightCommuter.DistanceTravelled+(rightCommuter.Speed/rightCommuter.BrakeSpeed)*(rightCommuter.Speed/2) {
											// But only part of it is, it's partly next to us: don't move!
											canMove = false
											break // since we dont move
										}
										if before == -1 {
											// First one that's completely behind us, so let's move in front of that one!
											before = rightIndex
											break // since we can move
										}
									}
								}

								if canMove {
									// Move!
									comm.SwitchingLane = true
									comm.NewLane = road.Lanes[lane.Index+1]
									// Add this commuter to the commuters in that lane
									if before == -1 {
										comm.NewLane.Commuters = append(comm.NewLane.Commuters, comm)
									} else {
										comm.NewLane.Commuters = append(comm.NewLane.Commuters[:before], append([]*Commuter{comm}, comm.NewLane.Commuters[before:]...)...)
									}
								}
							}
							if !comm.SwitchingLane && hasLeft {
								canMove := true // by default
								before := -1
								for rightIndex, rightCommuter := range road.Lanes[lane.Index-1].Commuters {
									if rightCommuter.DistanceTravelled > comm.DistanceTravelled {
										// In front of us
										if rightCommuter.DistanceTravelled-rightCommuter.Width < comm.DistanceTravelled {
											// But only the front of it is, it's partly next to us: don't move!
											canMove = false
											break // since we dont move
										}
										if rightCommuter.DistanceTravelled-rightCommuter.Width-(comm.Speed/comm.BrakeSpeed)*(comm.Speed/2) < comm.DistanceTravelled {
											// We might bump into it, even though it's far away
											canMove = false
											break // since we dont move
										}
									} else if rightCommuter.DistanceTravelled < comm.DistanceTravelled {
										// Behind us
										if comm.DistanceTravelled-comm.Width < rightCommuter.DistanceTravelled+(rightCommuter.Speed/rightCommuter.BrakeSpeed)*(rightCommuter.Speed/2) {
											// But only part of it is, it's partly next to us: don't move!
											canMove = false
											break // since we dont move
										}
										if before == -1 {
											// First one that's completely behind us, so let's move in front of that one!
											before = rightIndex
											break // since we can move
										}
									}
								}

								if canMove {
									// Move!
									comm.SwitchingLane = true
									comm.NewLane = road.Lanes[lane.Index-1]
									// Add this commuter to the commuters in that lane
									if before == -1 {
										comm.NewLane.Commuters = append(comm.NewLane.Commuters, comm)
									} else {
										comm.NewLane.Commuters = append(comm.NewLane.Commuters[:before], append([]*Commuter{comm}, comm.NewLane.Commuters[before:]...)...)
									}
								}
							}
						}

						// Hit the brakes! (at least until we're done moving? )
						comm.Speed -= comm.BrakeSpeed * dt

					case distance > minCarDistance:
						// Speed up if we want to
						if comm.Speed < comm.PreferredSpeed {
							comm.Speed += comm.AccelerationSpeed * dt
						}
					}
				} else {
					// We're all alone
					switch {
					case comm.Speed < comm.PreferredSpeed:
						comm.Speed += comm.AccelerationSpeed * dt
					case comm.Speed > comm.PreferredSpeed:
						comm.Speed -= comm.BrakeSpeed * dt
					}
				}
			}
		}
	}
}

func (c *CommuterSystem) commuterDispatch() {
	estimates := c.commuterEstimates()
	for uid, estimate := range estimates {
		city := c.cities[uid]
		for _, road := range city.Roads {
			if city.Population == 0 {
				continue // with other Cities
			}

			// perRoad commuters want to leave this "minute"
			perRoad := estimate / len(city.Roads)

			for _, lane := range road.Lanes {
				curCmtrs := len(lane.Commuters)
				if perRoad > 0 && (curCmtrs == 0 || lane.Commuters[curCmtrs-1].DistanceTravelled > MinTravelDistance) {
					c.newCommuter(road, lane)
					perRoad--
				}
			}
		}
	}
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

		if math.Abs(comm.SwitchingProgress) == laneWidth {
			// Done switching

			// Remove it from lane
			remove := -1
			for index, laneCommuter := range comm.Lane.Commuters {
				if laneCommuter.ID() == comm.ID() {
					remove = index
					break
				} else if laneCommuter.DistanceTravelled < comm.DistanceTravelled {
					break // pruning
				}
			}
			if remove >= 0 {
				comm.Lane.Commuters = append(comm.Lane.Commuters[:remove], comm.Lane.Commuters[remove+1:]...)
			}

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

						crashed = append(crashed, comm.BasicEntity)
						crashed = append(crashed, lane.Commuters[commIndex-1].BasicEntity)
						// Now let's remove the two cars

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
	for _, comm := range c.commuters {
		if comm.DistanceTravelled > comm.Road.SpaceComponent.Width-15 {
			engo.Mailbox.Dispatch(CommuterArrivedMessage{comm.CommuterComponent})

			// Done!
			c.cities[comm.Road.To.ID()].Population++
			c.world.RemoveEntity(*comm.BasicEntity)
		}
	}
}

func (c *CommuterSystem) commuterEstimates() map[uint64]int {
	rushHours := []float32{8.50, 17.50} // so that's 8:30 am and 5:30 pm

	hr := float32(c.clock.Time.Hour()) + float32(c.clock.Time.Second())/60

	diff := float32(math.MaxFloat32)
	for _, rushHr := range rushHours {
		d := math.Pow(math.Abs(hr-rushHr), 3)
		if d < diff {
			diff = d
		}
	}
	if diff == 0 {
		diff = 0.0001
	}

	estimates := make(map[uint64]int)
	counter := 0
	for uid, city := range c.cities {
		estimates[uid] = int(float32(city.Population) / (0.5 * diff))
		counter++
	}

	return estimates
}

func (c *CommuterSystem) newCommuter(road *Road, lane *Lane) {
	cmtr := &Commuter{BasicEntity: ecs.NewBasic()}
	cmtr.CommuterComponent = CommuterComponent{
		PreferredSpeed:    float32(rand.Intn(60) + 80), // 80 being minimum speed, 40 being the variation,
		Speed:             50,                          // coming from city
		AccelerationSpeed: 80,                          // for this car specifically
		BrakeSpeed:        240,                         // for this car specifically

		Lane: lane,
		Road: road,
	}
	cmtr.SpaceComponent = common.SpaceComponent{
		Position: road.SpaceComponent.Position,
		Width:    12,
		Height:   6,
		Rotation: road.Rotation,
	}
	cmtr.RenderComponent = common.RenderComponent{
		Drawable: common.Rectangle{BorderWidth: 0.5, BorderColor: color.RGBA{128, 128, 128, 128}},
		Color:    color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255},
	}

	// Translate the commuter for the given lane (hopefully!) - TODO: this can be a Shader
	angle := (cmtr.Rotation / 180) * math.Pi
	lanewidth := float32(lane.Index) * laneWidth
	dx := math.Sin(angle) * (lanewidth + 2) // 2 == (laneHeight - carHeight)/2
	dy := math.Cos(angle) * (lanewidth + 2) // 2 == (laneHeight - carHeight)/2
	cmtr.SpaceComponent.Position.X -= dx
	cmtr.SpaceComponent.Position.Y += dy

	cmtr.SetZIndex(50)
	cmtr.SetShader(common.LegacyShader)

	city := c.cities[road.From.ID()]
	city.Population--

	for _, system := range c.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&cmtr.BasicEntity, &cmtr.RenderComponent, &cmtr.SpaceComponent)
		case *CommuterSystem:
			sys.AddCommuter(&cmtr.BasicEntity, &cmtr.CommuterComponent, &cmtr.SpaceComponent)
		}
	}

	lane.Commuters = append(lane.Commuters, cmtr)
}

type CommuterArrivedMessage struct {
	Commuter *CommuterComponent
}

func (CommuterArrivedMessage) Type() string { return "CommuterArrivedMessage" }
