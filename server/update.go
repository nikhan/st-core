package server

type Update struct {
	Action string      `json:"action"`
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
}

type BroadcastId struct {
	Id int `json:"id"`
}

type BroadcastLabel struct {
	BroadcastId
	Label string `json:"label"`
}

type BroadcastPosition struct {
	BroadcastId
	Position Position `json:"position"`
}

type BroadcastBlockCreate struct {
	Block BlockLedger `json:"block"`
}

type BroadcastBlockLabel struct {
	Block BroadcastLabel `json:"block"`
}

type BroadcastBlockPosition struct {
	Block BroadcastPosition `json:"block"`
}

type BroadcastBlockDelete struct {
	Block BroadcastId `json:"block"`
}

type BroadcastGroupCreate struct {
	Group Group `json:"group"`
}

type BroadcastGroupLabel struct {
	Group BroadcastLabel `json:"group"`
}

type BroadcastGroupPosition struct {
	Group BroadcastPosition `json:"group"`
}

type BroadcastGroupDelete struct {
	Group BroadcastId `json:"group"`
}

type BroadcastGroupChild struct {
	Group BroadcastId `json:"group"`
	Child BroadcastId `json:"child"`
}

type BroadcastSourceCreate struct {
	Source SourceLedger `json:"source"`
}

type BroadcastSourceLabel struct {
	Source BroadcastLabel `json:"source"`
}

type BroadcastSourcePosition struct {
	Source BroadcastPosition `json:"source"`
}

type BroadcastSourceModify struct {
	Source struct {
		BroadcastId
		Param string `json:"param"`
		Value string `json:"value"`
	} `json:"source"`
}

type BroadcastSourceDelete struct {
	Source BroadcastId `json:"source"`
}

type BroadcastLinkCreate struct {
	Link struct {
		BroadcastId
		Source BroadcastId `json:"source"`
		Block  BroadcastId `json:"block"`
	} `json:"link"`
}

type BroadcastLinkDelete struct {
	Link BroadcastId `json:"link"`
}

type BroadcastConnectionCreate struct {
	Connection ConnectionLedger `json:"connection"`
}

type BroadcastConnectionDelete struct {
	Connection BroadcastId `json:"connection"`
}
