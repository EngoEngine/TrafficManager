package main

import (
	"image/color"
	"log"

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
)

type myScene struct{}

// Type uniquely defines your game type
func (*myScene) Type() string { return "myGame" }

// Preload is called before loading any assets from the disk, to allow you to register / queue them
func (*myScene) Preload() {
	err := engo.Files.LoadMany("textures/city.png", "fonts/Roboto-Regular.ttf")
	if err != nil {
		log.Println("[FATAL]", err)
	}
}

// Setup is called before the main loop starts. It allows you to add entities and systems to your Scene.
func (*myScene) Setup(world *ecs.World) {
	common.SetBackground(color.RGBA{0xf0, 0xf0, 0xf0, 0xff})

	world.AddSystem(&common.RenderSystem{})
	world.AddSystem(&common.MouseSystem{})
	world.AddSystem(common.NewKeyboardScroller(KeyboardScrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis))
	world.AddSystem(&common.EdgeScroller{EdgeScrollSpeed, EdgeWidth})
	world.AddSystem(&common.MouseZoomer{ZoomSpeed})

	world.AddSystem(&systems.CityBuildingSystem{})
	world.AddSystem(&systems.RoadBuildingSystem{})
	world.AddSystem(&systems.HUDSystem{})
	world.AddSystem(&systems.CommuterSystem{})

	fnt := common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.Black,
		Size: 24,
	}
	err := fnt.CreatePreloaded()
	if err != nil {
		log.Println(err)
		return
	}

	welcome := systems.HUDText{}
	welcome.SpaceComponent.Width = engo.CanvasWidth()
	welcome.SpaceComponent.Position = engo.Point{4, 4}
	welcome.RenderComponent.Drawable = fnt.Render("Welcome! Press <B> to spawn cities. ")

	welcome.RenderComponent.SetShader(common.HUDShader)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&welcome.BasicEntity, &welcome.RenderComponent, &welcome.SpaceComponent)
		}
	}
}

func main() {
	opts := engo.RunOptions{
		Title:          "TrafficManager",
		Width:          800,
		Height:         800,
		StandardInputs: true,
	}
	engo.Run(opts, &myScene{})
}
