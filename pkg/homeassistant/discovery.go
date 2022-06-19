package homeassistant

type HaDeviceDiscoveryInfo struct {
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
	UnitOfMeasurement   string `json:"unit_of_measurement"`
}
