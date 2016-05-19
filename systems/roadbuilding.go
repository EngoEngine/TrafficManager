package systems

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"fmt"
	"github.com/luxengine/math"
)

var (
	colorDefault         = color.Black
	colorHover           = color.RGBA{100, 100, 255, 255}
	colorSelectedBorder  = colorHover
	colorRoadAvailable   = color.RGBA{0, 255, 0, 255}
	colorRoadUnavailable = color.RGBA{255, 0, 0, 255}
	colorRoadDefault     = color.RGBA{128, 128, 128, 255}

	costPerUnit float32 = 100
	laneWidth   float32 = 10
)

type Road struct {
	ecs.BasicEntity
	RoadComponent
	common.SpaceComponent
	common.RenderComponent
	common.MouseComponent
}

type RoadComponent struct {
	Type     RoadType
	From, To ecs.BasicEntity
	Lanes    []*Lane

	isHovered bool
}

type Lane struct {
	ecs.BasicEntity
	LaneComponent
}

type LaneComponent struct {
	Type      RoadType
	Commuters []*Commuter
	Index     int
}

type Commuter struct {
	ecs.BasicEntity
	common.SpaceComponent
	common.RenderComponent
	CommuterComponent
}

type CommuterComponent struct {
	DistanceTravelled float32

	Speed             float32
	PreferredSpeed    float32
	BrakeSpeed        float32
	AccelerationSpeed float32

	SwitchingLane     bool
	SwitchingProgress float32
	Road              *Road
	Lane              *Lane
	NewLane           *Lane

	// TODO: stuff like reaction time, amount of people,
}

type RoadType uint8

const (
	RoadNone RoadType = iota
	RoadBasic
)

const (
	roadZIndex = -1
)

type roadBuildingEntity struct {
	*ecs.BasicEntity
	*CityComponent
	*common.SpaceComponent
	*common.RenderComponent
	*common.MouseComponent
}

type RoadBuildingSystem struct {
	world  *ecs.World
	cities []roadBuildingEntity

	roadHint       Road
	roadCostHint   HUDText
	selectedEntity roadBuildingEntity
	hovering       bool
	mouseTracker   MouseTracker
}

func (r *RoadBuildingSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range r.cities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		r.cities = append(r.cities[:delete], r.cities[delete+1:]...)
	}
}

func (r *RoadBuildingSystem) AddCity(basic *ecs.BasicEntity, city *CityComponent, space *common.SpaceComponent, render *common.RenderComponent, mouse *common.MouseComponent) {
	r.cities = append(r.cities, roadBuildingEntity{basic, city, space, render, mouse})
}

func (r *RoadBuildingSystem) New(w *ecs.World) {
	r.world = w

	r.mouseTracker.BasicEntity = ecs.NewBasic()
	r.mouseTracker.MouseComponent = common.MouseComponent{Track: true}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.MouseSystem:
			sys.Add(&r.mouseTracker.BasicEntity, &r.mouseTracker.MouseComponent, nil, nil)
		}
	}
}

func (r *RoadBuildingSystem) Update(dt float32) {
	var hovered bool
	var hoveredId int = -1

	for index, e := range r.cities {
		// The entity we've clicked
		if e.MouseComponent.Clicked {
			if r.selectedEntity.BasicEntity == nil {
				// Select the City
				r.selectedEntity = e
				r.selectedEntity.Color = colorDefault
				e.RenderComponent.Drawable = common.Rectangle{BorderColor: colorSelectedBorder, BorderWidth: 5}
			} else {
				if r.selectedEntity.BasicEntity.ID() != e.BasicEntity.ID() {
					// Check if we can afford it (and if so, pay for it)
					if cashAmount < r.roadHint.Width*costPerUnit {
						break // can't afford it
					}
					cashAmount -= r.roadHint.Width * costPerUnit

					// Check if one exists already
					var currentRoad *Road
					for _, road := range r.selectedEntity.Roads {
						if road.To.ID() == e.BasicEntity.ID() {
							currentRoad = road
							break
						}
					}

					if currentRoad == nil {
						// Build a Road
						actualRoad := Road{BasicEntity: ecs.NewBasic()}
						actualRoad.SpaceComponent = r.roadHint.SpaceComponent
						actualRoad.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{BorderWidth: 0.5, BorderColor: color.Black}, Color: colorRoadDefault}
						actualRoad.RenderComponent.SetZIndex(roadZIndex)
						actualRoad.RenderComponent.SetShader(common.LegacyShader)
						actualRoad.RoadComponent = RoadComponent{
							Type: RoadBasic,
							From: *r.selectedEntity.BasicEntity,
							To:   *e.BasicEntity,
						}
						actualRoad.Lanes = []*Lane{&Lane{BasicEntity: ecs.NewBasic(), LaneComponent: LaneComponent{Index: 0}}}

						for _, system := range r.world.Systems() {
							switch sys := system.(type) {
							case *common.RenderSystem:
								sys.Add(&actualRoad.BasicEntity, &actualRoad.RenderComponent, &actualRoad.SpaceComponent)
							case *CommuterSystem:
								sys.AddRoad(&actualRoad.BasicEntity, &actualRoad.RoadComponent, &actualRoad.SpaceComponent)
							case *SpeedCameraBuildingSystem:
								sys.AddRoad(&actualRoad.BasicEntity, &actualRoad.RoadComponent, &actualRoad.SpaceComponent, &actualRoad.MouseComponent)
							case *common.MouseSystem:
								sys.Add(&actualRoad.BasicEntity, &actualRoad.MouseComponent, &actualRoad.SpaceComponent, &actualRoad.RenderComponent)
							case *LawSystem:
								sys.AddRoad(&actualRoad.BasicEntity, &actualRoad.RoadComponent)
							}
						}

						r.selectedEntity.Roads = append(r.selectedEntity.Roads, &actualRoad)
					} else {
						// Add a Lane to it
						currentRoad.Lanes = append(currentRoad.Lanes, &Lane{
							BasicEntity:   ecs.NewBasic(),
							LaneComponent: LaneComponent{Index: len(currentRoad.Lanes)},
						})
						currentRoad.SpaceComponent.Height = laneWidth * float32(len(currentRoad.Lanes))
					}
				}

				// Cleanup the roadHint
				r.world.RemoveEntity(r.roadHint.BasicEntity)
				r.roadHint = Road{}

				r.world.RemoveEntity(r.roadCostHint.BasicEntity)
				r.roadCostHint = HUDText{}

				// Deselect the City
				e.RenderComponent.Color = colorDefault
				e.RenderComponent.Drawable = common.Rectangle{}
				r.selectedEntity.RenderComponent.Drawable = common.Rectangle{} // so no border
				r.selectedEntity = roadBuildingEntity{}
			}
		}

		// The entity we're hovering (or not)
		if r.selectedEntity.BasicEntity == nil || r.selectedEntity.BasicEntity.ID() != e.BasicEntity.ID() {
			if e.MouseComponent.Hovered {
				// If it's hovered, we should make it visual
				e.RenderComponent.Color = colorHover
				e.isHovered = true
				hovered = true

				hoveredId = index
			} else if e.isHovered {
				// Then reset to base values
				e.RenderComponent.Color = colorDefault
			}
		}
	}

	// The (possibly non-existent) roadHint
	if r.selectedEntity.BasicEntity != nil {
		// We should make it extra visual if a road can be built
		var roadHintNew bool
		var target common.SpaceComponent

		if hoveredId >= 0 {
			target = *r.cities[hoveredId].SpaceComponent
		} else {
			target = common.SpaceComponent{Position: engo.Point{r.mouseTracker.MouseX, r.mouseTracker.MouseY}}
		}

		if r.roadHint.BasicEntity.ID() == 0 {
			r.roadHint = Road{BasicEntity: ecs.NewBasic()}
			r.roadHint.RenderComponent = common.RenderComponent{
				Drawable: common.Rectangle{},
			}
			r.roadHint.RenderComponent.SetZIndex(1)
			r.roadHint.RenderComponent.SetShader(common.LegacyShader)

			roadHintNew = true
		}

		ab1 := target.AABB()
		ab2 := r.selectedEntity.SpaceComponent.AABB()
		centerA := engo.Point{(ab1.Max.X-ab1.Min.X)/2 + ab1.Min.X, (ab1.Max.Y-ab1.Min.Y)/2 + ab1.Min.Y}
		centerB := engo.Point{(ab2.Max.X-ab2.Min.X)/2 + ab2.Min.X, (ab2.Max.Y-ab2.Min.Y)/2 + ab2.Min.Y}

		roadWidth := laneWidth

		// Euclidian distance between the two cities
		roadLength := math.Sqrt(
			math.Pow(centerA.X-centerB.X, 2) +
				math.Pow(centerA.Y-centerB.Y, 2),
		)

		if hoveredId >= 0 && cashAmount >= roadLength*costPerUnit {
			r.roadHint.RenderComponent.Color = colorRoadAvailable
		} else {
			r.roadHint.RenderComponent.Color = colorRoadUnavailable
		}

		// Using the Law of Cosines
		// Solve for "alpha": (a2 means a squared)
		// a2 = b2 + c2 - 2bc * cos alpha
		// a2 - b2 - c2 = - 2bc * cos alpha
		// -a2 + b2 + c2 = 2bc * cos alpha
		// (-a2 + b2 + c2)/(2bc) = cos alpha
		// arccos ((-a2 + b2 + c2)/(2bc)) = alpha
		a := centerA.Y - centerB.Y // dy
		b := roadLength
		c := centerA.X - centerB.X // dx
		rotation_rad := math.Acos((-math.Pow(a, 2) + math.Pow(b, 2) + math.Pow(c, 2)) / (2 * b * c))
		rotation := 180 * (rotation_rad / math.Pi)

		dirY := float32(1)
		if centerA.Y < centerB.Y {
			dirY = -1
		}

		laneCount := float32(1)
		if hoveredId >= 0 {
			for _, road := range r.selectedEntity.Roads {
				if road.To.ID() == r.cities[hoveredId].ID() {
					laneCount += float32(len(road.Lanes))
					break
				}
			}
		}

		r.roadHint.SpaceComponent = common.SpaceComponent{
			Position: engo.Point{
				centerB.X - roadWidth/2,
				centerB.Y - roadWidth/2,
			},
			Width:    roadLength,
			Height:   roadWidth * laneCount,
			Rotation: rotation * dirY,
		}

		if roadHintNew {
			for _, system := range r.world.Systems() {
				switch sys := system.(type) {
				case *common.RenderSystem:
					sys.Add(&r.roadHint.BasicEntity, &r.roadHint.RenderComponent, &r.roadHint.SpaceComponent)
				}
			}
		}

		// Add money hint as well
		var action string
		if laneCount == 1 {
			action = "Build"
		} else {
			action = "Expand"
		}

		if hoveredId < 0 || cashAmount < roadLength*costPerUnit {
			action += " unavailable"
		}

		cost := costPerUnit * roadLength

		var roadCostHintNew bool
		if r.roadCostHint.ID() == 0 {
			r.roadCostHint.BasicEntity = ecs.NewBasic()
			roadCostHintNew = true
		}

		r.roadCostHint.SpaceComponent = common.SpaceComponent{
			Position: engo.Point{engo.Input.Mouse.X, engo.Input.Mouse.Y + 20},
			Width:    200, // TODO: set values?
			Height:   16,
		}

		fnt := common.Font{
			URL:  "fonts/Roboto-Regular.ttf",
			FG:   color.RGBA{100, 0, 0, 200}, // dark red, but somewhat transparent
			Size: 16,
		}
		fnt.CreatePreloaded()

		r.roadCostHint.RenderComponent = common.RenderComponent{
			Drawable: fnt.Render(fmt.Sprintf("%s ($ %.2f)", action, cost)),
		}
		r.roadCostHint.SetShader(common.HUDShader)

		if roadCostHintNew {
			for _, system := range r.world.Systems() {
				switch sys := system.(type) {
				case *common.RenderSystem:
					sys.Add(&r.roadCostHint.BasicEntity, &r.roadCostHint.RenderComponent, &r.roadCostHint.SpaceComponent)
				}
			}
		}
	}

	// Set the cursor so we know what we're hovering
	if hovered && !r.hovering {
		engo.SetCursor(engo.CursorHand)
		r.hovering = true
	} else if !hovered && r.hovering {
		engo.SetCursor(engo.CursorNone)
		r.hovering = false
	}

}