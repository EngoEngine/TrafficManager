package systems

import (
	"fmt"
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/luxengine/math"
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
	camHintText HUDText

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

			builder.camHint.LawComponent = LawComponent{
				Name:      fmt.Sprintf("SpeedCamera %d", builder.camHint.BasicEntity.ID()),
				ChecksFor: []LawType{LawMaxSpeed},
			}

			for _, system := range builder.world.Systems() {
				switch sys := system.(type) {
				case *common.RenderSystem:
					sys.Add(&builder.camHint.BasicEntity, &builder.camHint.RenderComponent, &builder.camHint.SpaceComponent)
				}
			}

		} else {
			builder.world.RemoveEntity(builder.camHint.BasicEntity)
			builder.camHint = SpeedCheckpoint{}

			builder.building = false
		}
	}

	// TODO: update location of hint
	if builder.building {
		builder.camHint.SpaceComponent.Position = engo.Point{builder.mouseTracker.MouseX, builder.mouseTracker.MouseY}

		roadIndex := -1
		for index, road := range builder.roads {
			if road.Hovered {
				roadIndex = index
				break
			}
		}

		if roadIndex >= 0 {
			const padding float32 = 5

			road := builder.roads[roadIndex]
			builder.camHint.Color = color.RGBA{0, 255, 0, 128}
			builder.camHint.Rotation = (road.Rotation + 90)
			builder.camHint.Width = road.Height + 2*padding

			// This code translates the camHint to "snap" onto the road (DistanceTravelled idea)
			var (
				x_length = (road.Position.X - builder.camHint.Position.X)
				y_length = (road.Position.Y - builder.camHint.Position.Y)

				sin, cos = math.Sincos(-road.Rotation * math.Pi / 180)

				//non_rotatedX = road.Position.X +
				width  = x_length*cos - y_length*sin
				height = -padding

				revSin, revCos = math.Sincos(road.Rotation * math.Pi / 180)
			)

			builder.camHint.Position.X = road.Position.X - width*revCos + height*sin
			builder.camHint.Position.Y = road.Position.Y - width*revSin + height*cos
		} else {
			builder.camHint.Color = color.RGBA{255, 0, 0, 128}
			builder.camHint.Rotation = 0
		}
	}
}
