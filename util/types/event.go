package types

type Events []*Event

type Event struct {
	Type string
	Obj  interface{}
}