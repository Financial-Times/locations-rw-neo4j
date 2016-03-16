package locations

type Location struct {
	UUID          string `json:"uuid"`
	CanonicalName string `json:"canonicalName"`
	TmeIdentifier string `json:"tmeIdentifier,omitempty"`
	Type          string `json:"type,omitempty"`
}

type LocationLink struct {
	ApiUrl string `json:"apiUrl"`
}
