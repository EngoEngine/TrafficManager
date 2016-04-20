package main

import (
	"engo.io/engo"
	"engo.io/ecs"
	"image/color"
)

type myScene struct {}

type City struct {
	ecs.BasicEntity
	engo.RenderComponent
	engo.SpaceComponent
}

// Type uniquely defines your game type
func (*myScene) Type() string { return "myGame" }

// Preload is called before loading any assets from the disk, to allow you to register / queue them
func (*myScene) Preload() {
	engo.Files.Add("assets/textures/city.png")
}

// Setup is called before the main loop starts. It allows you to add entities and systems to your Scene.
func (*myScene) Setup(world *ecs.World) {
	engo.SetBackground(color.White)

	world.AddSystem(&engo.RenderSystem{})

	city := City{BasicEntity: ecs.NewBasic()}

	city.SpaceComponent = engo.SpaceComponent{
		Position: engo.Point{10, 10},
		Width: 303,
		Height: 641,
	}

	texture := engo.Files.Image("city.png")
	city.RenderComponent = engo.NewRenderComponent(
		texture,
		engo.Point{1, 1},
		"city texture",
	)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *engo.RenderSystem:
			sys.Add(&city.BasicEntity, &city.RenderComponent, &city.SpaceComponent)
		}
	}
}

func main() {
	opts := engo.RunOptions{
		Title: "Hello World",
		Width: 400,
		Height: 400,
	}
	engo.Run(opts, &myScene{})
}
