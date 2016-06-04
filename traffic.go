package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/EngoEngine/TrafficManager/systems"
)

const (
	KeyboardScrollSpeed = 400
	EdgeScrollSpeed     = KeyboardScrollSpeed
	EdgeWidth           = 20
	ZoomSpeed           = -0.125

	WorldPadding = 50
)

type myScene struct{}

// Type uniquely defines your game type
func (*myScene) Type() string { return "myGame" }

// Preload is called before loading any assets from the disk, to allow you to register / queue them
func (*myScene) Preload() {
	common.AudioSystemPreload()
	err := engo.Files.Load(
		"textures/city.png",
		"fonts/Roboto-Regular.ttf",
		"fonts/fontello.ttf",
		"logic/1.level.yaml",
		"logic/vehicles.yaml",
	)
	if err != nil {
		panic(err)
	}

	// These are allowed to fail
	engo.Files.Load("sfx/crash.wav")
}

// Setup is called before the main loop starts. It allows you to add entities and systems to your Scene.
func (*myScene) Setup(world *ecs.World) {
	common.SetBackground(color.RGBA{0xf0, 0xf0, 0xf0, 0xff})

	world.AddSystem(&common.RenderSystem{})
	world.AddSystem(&common.MouseSystem{})
	world.AddSystem(&common.AudioSystem{})
	world.AddSystem(common.NewKeyboardScroller(KeyboardScrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis))
	world.AddSystem(&common.EdgeScroller{EdgeScrollSpeed, EdgeWidth})
	world.AddSystem(&common.MouseZoomer{ZoomSpeed})

	world.AddSystem(&systems.RoadBuildingSystem{})
	world.AddSystem(&systems.HUDSystem{})
	world.AddSystem(&systems.CommuterSystem{})
	world.AddSystem(&systems.LawSystem{})
	world.AddSystem(&systems.SpeedCameraBuildingSystem{})
	world.AddSystem(&systems.KeyboardZoomSystem{})
	world.AddSystem(&systems.MoneySystem{})
	world.AddSystem(&systems.TimeSystem{})
	world.AddSystem(&systems.WaveSystem{})

	// Load this specific level
	lvlRes, err := engo.Files.Resource("logic/1.level.yaml")
	if err != nil {
		panic(err)
	}

	lvl := lvlRes.(systems.LevelResource)
	var min, max engo.Point
	for _, city := range lvl.Level.Cities {
		if min.X == 0 || city.X < min.X {
			min.X = city.X
		}
		if min.Y == 0 || city.Y < min.Y {
			min.Y = city.Y
		}
		if city.X > max.X {
			max.X = city.X
		}
		if city.Y > max.Y {
			max.Y = city.Y
		}
	}

	common.CameraBounds = engo.AABB{min, max}
	cities := make([]*systems.City, len(lvl.Level.Cities))

	for i, city := range lvl.Level.Cities {
		cities[i] = systems.BuildCity(city.X, city.Y, city.Category, world)
	}

	// Load vehicles
	vehRes, err := engo.Files.Resource("logic/vehicles.yaml")
	if err != nil {
		panic(err)
	}
	veh := vehRes.(systems.VehicleResource)

	vehMap := make(map[string]systems.Vehicle)
	for _, vehicle := range veh.Vehicles.Vehicles {
		vehMap[vehicle.Name] = vehicle
	}

	bg := Background{BasicEntity: ecs.BasicEntity{}}
	bg.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{min.X - WorldPadding, min.Y - WorldPadding},
		Width:    max.X - min.X + systems.CityWidth + 2*WorldPadding,
		Height:   max.Y - min.Y + systems.CityHeight + 2*WorldPadding,
	}
	bg.RenderComponent = common.RenderComponent{
		Drawable: common.Rectangle{},
		Color:    color.RGBA{200, 200, 200, 255},
	}
	bg.SetZIndex(-10000)
	bg.SetShader(common.LegacyShader)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&bg.BasicEntity, &bg.RenderComponent, &bg.SpaceComponent)
		case *systems.CommuterSystem:
			sys.Vehicles = vehMap
		case *systems.WaveSystem:
			sys.SetWaves(lvl.Level.Waves)
		}
	}
}

type Background struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

func main() {
	opts := engo.RunOptions{
		Title:          "TrafficDefense",
		Width:          800,
		Height:         800,
		StandardInputs: true,
		MSAA:           4,
	}
	engo.Run(opts, &myScene{})
}
