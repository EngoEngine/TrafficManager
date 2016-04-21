package main

import (
	"image"
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

type HUD struct {
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

	world.AddSystem(&engo.MouseSystem{})
	world.AddSystem(&engo.RenderSystem{})

	kbs := engo.NewKeyboardScroller(KeyboardScrollSpeed, engo.W, engo.D, engo.S, engo.A)
	kbs.BindKeyboard(engo.ArrowUp, engo.ArrowRight, engo.ArrowDown, engo.ArrowLeft)
	world.AddSystem(kbs)

	world.AddSystem(&engo.EdgeScroller{EdgeScrollSpeed, EdgeWidth})
	world.AddSystem(&engo.MouseZoomer{ZoomSpeed})

	world.AddSystem(&systems.CityBuildingSystem{})

	hud := HUD{BasicEntity: ecs.NewBasic()}
	hud.SpaceComponent = engo.SpaceComponent{
		Position: engo.Point{0, engo.WindowHeight() - 200},
		Width:    200,
		Height:   200,
	}

	hudImage := image.NewUniform(color.RGBA{205, 205, 205, 255})
	hudNRGBA := engo.ImageToNRGBA(hudImage, 200, 200)
	hudImageObj := engo.NewImageObject(hudNRGBA)
	hudTexture := engo.NewTexture(hudImageObj)

	hud.RenderComponent = engo.NewRenderComponent(
		hudTexture,
		engo.Point{1, 1},
		"hud",
	)
	hud.RenderComponent.SetShader(engo.HUDShader)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *engo.RenderSystem:
			sys.Add(&hud.BasicEntity, &hud.RenderComponent, &hud.SpaceComponent)
		}
	}
}

func main() {
	opts := engo.RunOptions{
		Title:  "TrafficManager",
		Width:  800,
		Height: 800,
	}
	engo.Run(opts, &myScene{})
}
