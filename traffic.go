package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"

	"image/color"
	"log"
)

type City struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type myScene struct{}

// Type uniquely defines your game type
func (*myScene) Type() string { return "myGame" }

// Preload is called before loading any assets from the disk,
// to allow you to register / queue them
func (*myScene) Preload() {
	engo.Files.Load("textures/city.png")
}

// Setup is called before the main loop starts. It allows you
// to add entities and systems to your Scene.

func (*myScene) Setup(u engo.Updater) {
	world, _ := u.(*ecs.World)
	world.AddSystem(new(common.RenderSystem))
	city := City{BasicEntity: ecs.NewBasic()}
	city.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{10, 10},
		Width:    303,
		Height:   641,
	}

	texture, err := common.LoadedSprite("textures/city.png")
	if err != nil {
		log.Println("Unable to load texture: " + err.Error())
	}

	city.RenderComponent = common.RenderComponent{
		Drawable: texture,
		Scale:    engo.Point{1, 1},
	}

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&city.BasicEntity, &city.RenderComponent, &city.SpaceComponent)
		}
	}
	common.SetBackground(color.White)
}

func main() {
	opts := engo.RunOptions{
		Title:  "Hello World",
		Width:  400,
		Height: 400,
	}
	engo.Run(opts, new(myScene))
}
