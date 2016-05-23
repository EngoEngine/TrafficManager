package systems

import (
	"fmt"
	"image/color"
	"log"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/EngoEngine/TrafficDefense/systems/ui"
	"github.com/luxengine/math"
)

var (
	speedCamBackground = color.RGBA{150, 150, 200, 255}
)

type speedCameraBuildingEntityRoad struct {
	*ecs.BasicEntity
	*RoadComponent
	*common.SpaceComponent
	*common.MouseComponent
}

type SpeedCameraBuildingSystem struct {
	world *ecs.World

	mouseTracker MouseTracker

	building    bool
	camHint     SpeedCheckpoint
	camHintIcon common.Texture
	camHintText ui.Graphic

	roads []speedCameraBuildingEntityRoad
}

// Remove is called whenever an Entity is removed from the scene, and thus from this system
func (builder *SpeedCameraBuildingSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range builder.roads {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		builder.roads = append(builder.roads[:delete], builder.roads[delete+1:]...)
	}
}

// New is the initialisation of the System
func (builder *SpeedCameraBuildingSystem) New(w *ecs.World) {
	builder.world = w

	engo.Input.RegisterButton("build-speedcam", engo.F2, engo.N)

	builder.mouseTracker.BasicEntity = ecs.NewBasic()
	builder.mouseTracker.MouseComponent = common.MouseComponent{Track: true}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.MouseSystem:
			sys.Add(&builder.mouseTracker.BasicEntity, &builder.mouseTracker.MouseComponent, nil, nil)
		}
	}

	iconFont := common.Font{
		URL:  "fonts/fontello.ttf",
		Size: 26,
		FG:   color.Black,
	}

	err := iconFont.CreatePreloaded()
	if err != nil {
		log.Println("Could not load preloaded font:", err)
		return
	}

	builder.camHintIcon = iconFont.Render("\uE80A")
}

func (builder *SpeedCameraBuildingSystem) AddRoad(basic *ecs.BasicEntity, road *RoadComponent, space *common.SpaceComponent, mouse *common.MouseComponent) {
	builder.roads = append(builder.roads, speedCameraBuildingEntityRoad{basic, road, space, mouse})
}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (builder *SpeedCameraBuildingSystem) Update(dt float32) {
	if engo.Input.Button("build-speedcam").JustPressed() {
		if !builder.building {
			builder.building = true

			builder.camHint = SpeedCheckpoint{BasicEntity: ecs.NewBasic()}
			builder.camHint.SpaceComponent = common.SpaceComponent{
				Position: engo.Point{builder.mouseTracker.MouseComponent.MouseX, builder.mouseTracker.MouseComponent.MouseY},
				Width:    30,
				Height:   8,
			}
			builder.camHint.RenderComponent = common.RenderComponent{
				Drawable: common.Rectangle{},
			}
			builder.camHint.RenderComponent.SetZIndex(cityZIndex + 1)
			builder.camHint.RenderComponent.SetShader(common.LegacyShader)

			builder.camHint.icon = IconComponent{BasicEntity: ecs.NewBasic()}
			builder.camHint.icon.RenderComponent = common.RenderComponent{
				Drawable: builder.camHintIcon,
				Scale:    engo.Point{0.25, 0.25},
			}
			builder.camHint.icon.SpaceComponent = common.SpaceComponent{
				Position: engo.Point{
					builder.camHint.SpaceComponent.Position.X + builder.camHint.Width - builder.camHint.icon.Drawable.Width(),
					builder.camHint.SpaceComponent.Position.Y + builder.camHint.Height - builder.camHint.icon.Drawable.Height(),
				},
				Width:  builder.camHint.icon.Drawable.Width(),
				Height: builder.camHint.icon.Drawable.Height(),
			}
			builder.camHint.icon.RenderComponent.SetZIndex(cityZIndex + 2)

			for _, system := range builder.world.Systems() {
				switch sys := system.(type) {
				case *common.RenderSystem:
					sys.Add(&builder.camHint.BasicEntity, &builder.camHint.RenderComponent, &builder.camHint.SpaceComponent)
					sys.Add(&builder.camHint.icon.BasicEntity, &builder.camHint.icon.RenderComponent, &builder.camHint.icon.SpaceComponent)
				}
			}

		} else {
			builder.world.RemoveEntity(builder.camHint.icon.BasicEntity)
			builder.camHint.icon = IconComponent{}

			builder.world.RemoveEntity(builder.camHint.BasicEntity)
			builder.camHint = SpeedCheckpoint{}

			builder.building = false
		}
	}

	if builder.building {
		builder.camHint.SpaceComponent.Position = engo.Point{builder.mouseTracker.MouseX, builder.mouseTracker.MouseY}
		builder.camHint.icon.SpaceComponent.Position = engo.Point{
			builder.camHint.SpaceComponent.Position.X + builder.camHint.Width - builder.camHint.icon.Drawable.Width()*builder.camHint.icon.Scale.X,
			builder.camHint.SpaceComponent.Position.Y + builder.camHint.Height - builder.camHint.icon.Drawable.Height()*builder.camHint.icon.Scale.Y,
		}

		roadIndex := -1
		for index, road := range builder.roads {
			if road.Hovered {
				roadIndex = index
				break
			}
		}

		if roadIndex >= 0 {
			road := builder.roads[roadIndex]

			builder.camHint.RoadLocationComponent.RoadID = road.ID()

			padding := builder.camHint.Height * 1.25

			builder.camHint.Color = color.RGBA{0, 255, 0, 128}
			builder.camHint.Rotation = (road.Rotation + 90)
			builder.camHint.icon.Rotation = (road.Rotation + 90)
			builder.camHint.Width = road.Height + padding

			// This code translates the camHint to "snap" onto the road (DistanceTravelled idea)
			var (
				x_length = road.Position.X - builder.camHint.SpaceComponent.Position.X
				y_length = road.Position.Y - builder.camHint.SpaceComponent.Position.Y

				sin, cos = math.Sincos(-road.Rotation * math.Pi / 180)

				width = x_length*cos - y_length*sin

				revSin, revCos = math.Sincos(road.Rotation * math.Pi / 180)
			)

			builder.camHint.RoadLocationComponent.Position = width
			builder.camHint.SpaceComponent.Position.X = road.Position.X - width*revCos
			builder.camHint.SpaceComponent.Position.Y = road.Position.Y - width*revSin

			// This code translates the camHintIcon to "snap" onto the road (DistanceTravelled idea)
			width = x_length*cos - y_length*sin + (builder.camHint.Height-builder.camHint.icon.Drawable.Height()*builder.camHint.icon.Scale.Y)/2
			height := road.Height + (padding-builder.camHint.icon.Drawable.Width()*builder.camHint.icon.Scale.X)/2

			builder.camHint.icon.Position.X = road.Position.X - width*revCos + height*sin
			builder.camHint.icon.Position.Y = road.Position.Y - width*revSin + height*cos

			if road.Clicked {
				builder.buildSpeedCamera()
			}
		} else {
			builder.camHint.RoadLocationComponent.RoadID = math.MaxUint64
			builder.camHint.Color = color.RGBA{255, 0, 0, 128}
			builder.camHint.Rotation = 0

			builder.camHint.icon.Rotation = 0
		}
	}
}

func (builder *SpeedCameraBuildingSystem) buildSpeedCamera() {
	speedCam := SpeedCheckpoint{BasicEntity: ecs.NewBasic()}
	speedCam.SpaceComponent = builder.camHint.SpaceComponent
	speedCam.RenderComponent = common.RenderComponent{
		Drawable: common.Rectangle{},
		Color:    speedCamBackground,
	}
	speedCam.RenderComponent.SetZIndex(roadZIndex - 0.02) // so below the road
	speedCam.RenderComponent.SetShader(common.LegacyShader)
	speedCam.LawComponent = LawComponent{
		Name:      fmt.Sprintf("Speed Camera %d", speedCam.ID()),
		ChecksFor: []LawType{LawMaxSpeed},
	}
	speedCam.RoadLocationComponent.RoadID = builder.camHint.RoadID
	speedCam.RoadLocationComponent.Position = builder.camHint.RoadLocationComponent.Position

	speedCam.icon.BasicEntity = ecs.NewBasic()
	speedCam.icon.SpaceComponent = builder.camHint.icon.SpaceComponent
	speedCam.icon.RenderComponent = common.RenderComponent{
		Drawable: builder.camHintIcon,
		Scale:    engo.Point{0.25, 0.25},
	}
	speedCam.icon.RenderComponent.SetZIndex(roadZIndex - 0.01) // so below the road, above the speedCam background

	for _, system := range builder.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&speedCam.BasicEntity, &speedCam.RenderComponent, &speedCam.SpaceComponent)
			sys.Add(&speedCam.icon.BasicEntity, &speedCam.icon.RenderComponent, &speedCam.icon.SpaceComponent)
		case *LawSystem:
			sys.AddCheckpoint(&speedCam.BasicEntity, &speedCam.LawComponent, &speedCam.RoadLocationComponent)
		}
	}
}
