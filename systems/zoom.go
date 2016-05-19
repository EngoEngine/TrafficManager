package systems

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

// KeyboardZoomSystem allows you to zoom in/out using keyboard keys
type KeyboardZoomSystem struct{}

func (*KeyboardZoomSystem) Remove(ecs.BasicEntity) {}

func (*KeyboardZoomSystem) New(*ecs.World) {
	engo.Input.RegisterAxis("zoom", engo.AxisKeyPair{engo.NumSubtract, engo.NumAdd}, engo.AxisKeyPair{engo.Dash, engo.Equals})
}

func (*KeyboardZoomSystem) Update(dt float32) {
	switch engo.Input.Axis("zoom").Value() {
	case engo.AxisMin:
		engo.Mailbox.Dispatch(common.CameraMessage{
			Value:       dt,
			Incremental: true,
			Axis:        common.ZAxis,
		})
	case engo.AxisMax:
		engo.Mailbox.Dispatch(common.CameraMessage{
			Value:       -dt,
			Incremental: true,
			Axis:        common.ZAxis,
		})
	}
}
