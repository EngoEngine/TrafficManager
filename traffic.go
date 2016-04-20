package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
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
	engo.Files.Add("assets/textures/city.png")
}

// Setup is called before the main loop starts. It allows you to add entities and systems to your Scene.
func (*myScene) Setup(world *ecs.World) {
	engo.SetBackground(color.White)

	world.AddSystem(&engo.MouseSystem{})
	world.AddSystem(&engo.RenderSystem{})

	kbs := engo.NewKeyboardScroller(KeyboardScrollSpeed, engo.W, engo.D, engo.S, engo.A)
	kbs.BindKeyboard(engo.ArrowUp, engo.ArrowRight, engo.ArrowDown, engo.ArrowLeft)
	world.AddSystem(kbs)

	world.AddSystem(&engo.EdgeScroller{EdgeScrollSpeed, EdgeWidth})
	world.AddSystem(&engo.MouseZoomer{ZoomSpeed})

	world.AddSystem(&systems.CityBuildingSystem{})
}

func main() {
	opts := engo.RunOptions{
		Title:  "Hello World",
		Width:  400,
		Height: 400,
	}
	engo.Run(opts, &myScene{})
}
