package ui

import (
	"engo.io/ecs"
	"engo.io/engo/common"
)

type Label struct {
	Font  *common.Font
	cache string

	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

func (l *Label) SetText(s string) bool {
	if l.Font == nil {
		panic("Label.SetText called without setting Label.Font")
	}

	if l.cache == s {
		return false
	}

	if l.RenderComponent.Drawable == nil {
		l.RenderComponent.Drawable = common.Text{Font: l.Font}
	}

	fnt := l.RenderComponent.Drawable.(common.Text)
	fnt.Text = s

	return true
}

type Graphic struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type Button struct {
	Label   Label
	Graphic Graphic
	common.MouseComponent

	OnClick     func(*Button)
	OnMouseOver func(*Button)
	OnMouseOut  func(*Button)
}

func NewButton(f *common.Font, label string) *Button {
	b := new(Button)
	b.Label.BasicEntity = ecs.NewBasic()
	b.Graphic.BasicEntity = ecs.NewBasic()

	if f != nil && len(label) > 0 {
		b.Label.Font = f
		b.Label.SetText(label)
	}

	return b
}
