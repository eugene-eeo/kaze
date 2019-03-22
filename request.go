package main

type Reason uint32
type RequestType byte
type ExpiryType byte
type ActionType byte

const (
	RequestNewNotification = RequestType(iota)
	RequestCloseNotification
	RequestAction

	ActionCloseOne = ActionType(iota)
	ActionCloseTop
	ActionShowAll
	ActionContextMenu
	ActionContextMenuDone

	ExpiryTimeout = ExpiryType(iota)
	ExpiryPopupClose

	ReasonExpired = Reason(1 + iota)
	ReasonUserDismissed
	ReasonCloseNotification
	ReasonUndefined
)

type Request struct {
	Type RequestType
	Body interface{}
}

type ActionRequest struct {
	Type ActionType
	Nid  uint32
	Uid  UID
}

type Expiry struct {
	Type ExpiryType
	Uid  UID
}
