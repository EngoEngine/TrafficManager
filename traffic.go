package main

import (
	"engo.io/engo"
	"engo.io/ecs"
	"image/color"
	"github.com/EngoEngine/TrafficManager/systems"
)

type myGame struct {}

// Type uniquely defines your game type
func (*myGame) Type() string { return "myGame" }

// Preload is called before loading any assets from the disk, to allow you to register / queue them
func (*myGame) Preload() {
	engo.Files.Add("assets/textures/city.png")
}

// Setup is called before the main loop starts. It allows you to add entities and systems to your Scene.
func (*myGame) Setup(world *ecs.World) {
	engo.SetBackground(color.White)

	world.AddSystem(&engo.MouseSystem{})
	world.AddSystem(&engo.RenderSystem{})

	world.AddSystem(&systems.CityBuildingSystem{})
}

// Show is called whenever the other Scene becomes inactive, and this one becomes the active one
func (*myGame) Show() {}

// Hide is called when an other Scene becomes active
func (*myGame) Hide() {}

func main() {
	opts := engo.RunOptions{
		Title: "Hello World",
		Width: 400,
		Height: 400,
	}
	engo.Run(opts, &myGame{})
}
