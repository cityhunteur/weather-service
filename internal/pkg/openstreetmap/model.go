package openstreetmap

// Place represents a geographical location.
type Place struct {
	ID          int64  `json:"place_id"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}
