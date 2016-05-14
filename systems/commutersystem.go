package systems

import (
	"image/color"
	"log"
	"math/rand"
	"time"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/luxengine/math"
)

const (
	SpeedOne          = 1
	SpeedTwo          = 6
	SpeedThree        = 15
	MinTravelDistance = float32(24)
	MinCarDistance    = float32(12)
)

type commuterEntityCity struct {
	*ecs.BasicEntity
	*CityComponent
}

type commuterEntityRoad struct {
	*ecs.BasicEntity
	*RoadComponent
}

type commuterEntityCommuter struct {
	*ecs.BasicEntity
	*CommuterComponent
	*common.SpaceComponent
}

type CommuterSystem struct {
	world *ecs.World

	gameSpeed      float32
	gameTime       time.Time
	clock          HUDText
	clockFont      common.Font
	previousSecond int

	cities    map[uint64]commuterEntityCity
	roads     map[uint64]commuterEntityRoad
	commuters []commuterEntityCommuter
}

func (c *CommuterSystem) New(w *ecs.World) {
	c.world = w
	c.gameSpeed = SpeedOne
	c.cities = make(map[uint64]commuterEntityCity)
	c.roads = make(map[uint64]commuterEntityRoad)
	c.gameTime = time.Now()
	c.addClock()

	engo.Input.RegisterButton("speed1", engo.NumOne, engo.One)
	engo.Input.RegisterButton("speed2", engo.NumTwo, engo.Two)
	engo.Input.RegisterButton("speed3", engo.NumThree, engo.Three)
}

func (c *CommuterSystem) addClock() {
	var (
		height  float32 = 24
		width   float32 = height * 2.5
		padding float32 = 4
	)

	c.clockFont = common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.Black,
		Size: 24,
	}
	err := c.clockFont.CreatePreloaded()
	if err != nil {
		log.Println(err)
		return
	}

	c.clock.BasicEntity = ecs.NewBasic()
	c.clock.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{engo.CanvasWidth() - width - padding, padding},
		Width:    width + 2*padding,
		Height:   height + 2*padding,
	}
	c.clock.RenderComponent = common.RenderComponent{
		Drawable: c.clockFont.Render(c.gameTime.Format("15:04")),
		Color:    color.Black,
	}
	c.clock.SetZIndex(1000)
	c.clock.SetShader(common.HUDShader)

	for _, system := range c.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&c.clock.BasicEntity, &c.clock.RenderComponent, &c.clock.SpaceComponent)
		}
	}
}

func (c *CommuterSystem) Remove(basic ecs.BasicEntity) {
	delete(c.cities, basic.ID())
	delete(c.roads, basic.ID())

	delete := -1
	for index, e := range c.commuters {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		c.commuters = append(c.commuters[:delete], c.commuters[delete+1:]...)
	}
}

func (c *CommuterSystem) AddCity(basic *ecs.BasicEntity, city *CityComponent) {
	c.cities[basic.ID()] = commuterEntityCity{basic, city}
}

func (c *CommuterSystem) AddRoad(basic *ecs.BasicEntity, road *RoadComponent) {
	c.roads[basic.ID()] = commuterEntityRoad{basic, road}
}

func (c *CommuterSystem) AddCommuter(basic *ecs.BasicEntity, comm *CommuterComponent, space *common.SpaceComponent) {
	c.commuters = append(c.commuters, commuterEntityCommuter{basic, comm, space})
}

func (c *CommuterSystem) Update(dt float32) {
	// Update clock
	c.gameTime = c.gameTime.Add(time.Duration(float32(time.Minute) * dt * c.gameSpeed))
	c.clock.Drawable.Close()
	c.clock.Drawable = c.clockFont.Render(c.gameTime.Format("15:04"))
	c.clock.Position.X = engo.CanvasWidth() - c.clock.Width

	// Watch for speed changes
	if engo.Input.Button("speed1").Down() {
		c.gameSpeed = SpeedOne
	} else if engo.Input.Button("speed2").Down() {
		c.gameSpeed = SpeedTwo
	} else if engo.Input.Button("speed3").Down() {
		c.gameSpeed = SpeedThree
	}

	// Send commuters
	estimates := c.commuterEstimates()
	for uid, estimate := range estimates {
		city := c.cities[uid]
		for _, road := range city.Roads {
			// perRoad commuters want to leave this "minute"
			perRoad := estimate / len(city.Roads)

			curCmtrs := len(road.Commuters)
			if curCmtrs == 0 || (road.Commuters[curCmtrs-1].DistanceTravelled > MinTravelDistance && perRoad > 0) {
				c.newCommuter(road)
			}
		}
	}

	// Move commuters
	for _, road := range c.roads {
		for index, comm := range road.Commuters {
			// This bit computes how far we travel, and if we can drive through other cars in front of us (hint: of course not!)
			newDistance := comm.PreferredSpeed * dt * c.gameSpeed

			if index > 0 { // so there is a car in front of us
				distanceToNext := road.Commuters[index-1].DistanceTravelled - comm.DistanceTravelled
				if distanceToNext-newDistance < MinTravelDistance {
					newDistance = distanceToNext - MinTravelDistance
				}
			}

			comm.DistanceTravelled += newDistance

			// Using the Law of Sines, we now compute the dx (c) and dy (a)
			b_length := newDistance
			alpha := (comm.Rotation / 180) * math.Pi
			beta := float32(0.5) * math.Pi
			gamma := 0.5*math.Pi - alpha

			b_part := b_length / math.Sin(beta)
			a_length := math.Sin(alpha) * b_part
			c_length := math.Sin(gamma) * b_part

			comm.Position.Y += a_length
			comm.Position.X += c_length
		}
	}

	// Remove commuters
	for _, comm := range c.commuters {
		if comm.DistanceTravelled > comm.Road.SpaceComponent.Width-15 {
			// Done!
			log.Println(comm.ID(), "reached his/her destination")

			if len(comm.Road.Commuters) > 0 {
				comm.Road.Commuters = comm.Road.Commuters[1:]
			} else {
				comm.Road.Commuters = []*Commuter{}
			}

			city := c.cities[comm.Road.To.ID()]
			city.Population++

			c.world.RemoveEntity(*comm.BasicEntity)
		}
	}
}

func (c *CommuterSystem) commuterEstimates() map[uint64]int {
	rushHours := []float32{8.50, 17.50} // so that's 8:30 am and 5:30 pm

	hr := float32(c.gameTime.Hour()) + float32(c.gameTime.Second())/60

	diff := float32(math.MaxFloat32)
	for _, rushHr := range rushHours {
		d := math.Pow(math.Abs(hr-rushHr), 6)
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

func (c *CommuterSystem) newCommuter(road *Road) {
	cmtr := &Commuter{BasicEntity: ecs.NewBasic()}
	cmtr.PreferredSpeed = float32(rand.Intn(60) + 80) // 80 being minimum speed, 40 being the variation
	cmtr.Road = road
	cmtr.SpaceComponent = common.SpaceComponent{
		Position: road.SpaceComponent.Position,
		Width:    12,
		Height:   6,
		Rotation: road.Rotation,
	}
	cmtr.RenderComponent = common.RenderComponent{
		Drawable: common.Rectangle{},
		Color:    color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255},
	}
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

	road.Commuters = append(road.Commuters, cmtr)
}
