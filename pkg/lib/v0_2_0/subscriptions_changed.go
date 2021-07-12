package v0_2_0

const SubscriptionsChangedEventName = "sh.keptn.subscriptions.changed"

type SubscriptionsChangedEventData struct {
	IntegrationID string         `json:"integrationid"`
	Subscriptions []Subscription `json:"subscriptions"`
}

type Subscription struct {
	Topic   string `json:"topic"`
	Project string `json:"project"`
	Stage   string `json:"stage"`
	Service string `json:"service"`
}
