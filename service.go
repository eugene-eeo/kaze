package main

import "github.com/godbus/dbus"
import "sync"

type Urgency byte

const (
	UrgencyCritical = Urgency(2)
	UrgencyNormal   = Urgency(1)
	UrgencyLow      = Urgency(0)
)

type NotificationAction struct {
	Key   string
	Value string
}

type NotificationHints struct {
	Category string
	Resident bool // Whether to automatically remove the notification when an action has been invoked
	Urgency  Urgency
}

type Notification struct {
	Id            uint32
	AppName       string
	AppIcon       string
	Summary       string
	Body          string
	Hints         NotificationHints
	Actions       []NotificationAction
	ExpireTimeout int32
}

func convertRawActions(actions []string) []NotificationAction {
	rv := make([]NotificationAction, len(actions)/2)
	for i := 0; i < len(actions)/2; i++ {
		rv[i] = NotificationAction{
			Key:   actions[2*i],
			Value: actions[2*i+1],
		}
	}
	return rv
}

func convertRawHints(h map[string]dbus.Variant) NotificationHints {
	hints := NotificationHints{
		Category: "",
		Resident: false,
		Urgency:  UrgencyNormal,
	}
	for key, value := range h {
		switch key {
		case "category":
			category, ok := value.Value().(string)
			if ok {
				hints.Category = category
			}
		case "resident":
			resident, ok := value.Value().(bool)
			if ok {
				hints.Resident = resident
			}
		case "urgency":
			urgency, ok := value.Value().(byte)
			if ok && UrgencyLow <= Urgency(urgency) && Urgency(urgency) <= UrgencyCritical {
				hints.Urgency = Urgency(urgency)
			}
		}
	}
	return hints
}

type Service struct {
	id      uint32
	conn    *dbus.Conn
	lock    sync.Mutex
	handler NotificationHandler
}

func NewService(conn *dbus.Conn, handler NotificationHandler) *Service {
	return &Service{
		conn:    conn,
		handler: handler,
	}
}

func (s *Service) GetServerInformation() (string, string, string, string, *dbus.Error) {
	name := "kaze"
	vendor := "eugene-eeo.github.io"
	version := "0.1"
	spec_version := "1.2"
	return name, vendor, version, spec_version, nil
}

func (s *Service) GetCapabilities() ([]string, *dbus.Error) {
	return s.handler.Capabilities(), nil
}

func (s *Service) Notify(appName string, replacesId uint32, appIcon string, summary string, body string, actions []string, hints map[string]dbus.Variant, expireTimeout int32) (uint32, *dbus.Error) {
	id := replacesId
	if id == 0 {
		s.lock.Lock()
		s.id++
		// need to ensure that s.id > 0 if we get more than 2^32 notifications
		if s.id == 0 {
			s.id++
		}
		id = s.id
		s.lock.Unlock()
	}
	s.handler.HandleNotification(&Notification{
		Id:            id,
		AppName:       appName,
		AppIcon:       appIcon,
		Summary:       summary,
		Body:          body,
		Actions:       convertRawActions(actions),
		Hints:         convertRawHints(hints),
		ExpireTimeout: expireTimeout,
	})
	return id, nil
}

func (s *Service) CloseNotification(id uint32) *dbus.Error {
	return s.handler.HandleClose(id)
}
