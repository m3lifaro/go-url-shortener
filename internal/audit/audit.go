package audit

type Event struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id"`
	URL       string `json:"url"`
}

type Subscriber interface {
	Notify(event Event)
}

type Manager struct {
	subs []Subscriber
}

func (m *Manager) Register(sub Subscriber) {
	m.subs = append(m.subs, sub)
}

func (m *Manager) NotifyAll(event Event) {
	for _, sub := range m.subs {
		sub.Notify(event)
	}
}
