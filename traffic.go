package main

import (
	"image"
	"image/color"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	// EDIT THE FOLLOWING IMPORT TO YOUR systems package
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
	common.RenderComponent
	common.SpaceComponent
}

// Type uniquely defines your game type
func (*myScene) Type() string { return "myGame" }

// Preload is called before loading any assets from the disk, to allow you to register / queue them
func (*myScene) Preload() {
	engo.Files.Load("textures/city.png")
}

// Setup is called before the main loop starts. It allows you to add entities and systems to your Scene.
func (*myScene) Setup(u engo.Updater) {
	world, _ := u.(*ecs.World)
	common.SetBackground(color.White)

	world.AddSystem(&common.RenderSystem{})
	world.AddSystem(&common.MouseSystem{})

	kbs := common.NewKeyboardScroller(KeyboardScrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis)
	world.AddSystem(kbs)
	world.AddSystem(&common.EdgeScroller{EdgeScrollSpeed, EdgeWidth})
	world.AddSystem(&common.MouseZoomer{ZoomSpeed})

	world.AddSystem(&systems.CityBuildingSystem{})

	hud := HUD{BasicEntity: ecs.NewBasic()}
	hud.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{0, engo.WindowHeight() - 200},
		Width:    200,
		Height:   200,
	}

	hudImage := image.NewUniform(color.RGBA{205, 205, 205, 255})
	hudNRGBA := common.ImageToNRGBA(hudImage, 200, 200)
	hudImageObj := common.NewImageObject(hudNRGBA)
	hudTexture := common.NewTextureSingle(hudImageObj)

	hud.RenderComponent = common.RenderComponent{
		Repeat:   common.Repeat,
		Drawable: hudTexture,
		Scale:    engo.Point{1, 1},
	}
	hud.RenderComponent.SetShader(common.HUDShader)
	hud.RenderComponent.SetZIndex(1)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&hud.BasicEntity, &hud.RenderComponent, &hud.SpaceComponent)
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
