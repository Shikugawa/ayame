package network

type Attacheable interface {
	Attach(*Veth) error
}

type Link interface {
	Destroy() error
	CreateLink(left Attacheable, right Attacheable) error
}
