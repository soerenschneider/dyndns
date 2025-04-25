package conf

type NatsConfig struct {
	Url string `yaml:"url" env:"URL" validate:"required_with=EventsSubject DispatchUpdatesSubject ListenUpdatesSubjects,omitempty,nats_url"`

	EventsSubject          string `yaml:"events_subject" env:"EVENTS_SUBJECT" validate:"omitempty,nats_subject"`
	DispatchUpdatesSubject string `yaml:"dispatch_updates_subject" env:"UPDATE_REQUEST_SUBJECT" validate:"omitempty,nats_subject"`

	StreamName            string   `yaml:"stream_name" env:"STREAM_NAME" validate:"required_with=ConsumerName"`
	ListenUpdatesSubjects []string `yaml:"listen_updates_subjects" envSeparator:"," env:"STREAM_SUBJECTS" validate:"required_with=ConsumerName,omitempty,dive,nats_subject"`
	ConsumerName          string   `yaml:"consumer_name" env:"CONSUMER_NAME" validate:"required_with=StreamName"`
}

func (n *NatsConfig) SupportsCloudeventsDispatch() bool {
	return n.EventsSubject != ""
}

func (n *NatsConfig) IsConfiguredForUpdates() bool {
	return n.DispatchUpdatesSubject != ""
}

func (n *NatsConfig) IsConfiguredAsListener() bool {
	return n.ConsumerName != "" && n.StreamName != "" && len(n.ListenUpdatesSubjects) > 0
}
