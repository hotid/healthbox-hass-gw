package homeassistant

type HaCommonDiscoveryInfo struct {
	Device struct {
		Name        string `json:"name"`
		Identifiers string `json:"identifiers"`
	} `json:"device"`
	Name                string `json:"name"`
	UniqueId            string `json:"unique_id"`
	ObjectId            string `json:"object_id"`
	AvailabilityTopic   string `json:"availability_topic"`
	PayloadAvailable    string `json:"payload_available"`
	PayloadNotAvailable string `json:"payload_not_available"`
	StateTopic          string `json:"state_topic"`
}
type HaSensorDiscoveryInfo struct {
	HaCommonDiscoveryInfo
	UnitOfMeasurement string `json:"unit_of_measurement"`
	StateClass        string `json:"state_class"`
}

type HaSwitchDiscoveryInfo struct {
	HaCommonDiscoveryInfo
	CommandTopic string `json:"command_topic"`
	PayloadOn    string `json:"payload_on"`
	PayloadOff   string `json:"payload_off"`
	StateOn      string `json:"state_on"`
	StateOff     string `json:"state_off"`
}
