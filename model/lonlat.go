package model

import "github.com/asmarques/geodist"

// LonLat is a tuple of longitude,latitude in GeoJSON format.
// https://datatracker.ietf.org/doc/html/rfc7946#section-3.1.1
type LonLat [2]float64

func (pos LonLat) point() geodist.Point {
	return geodist.Point{
		Lat:  pos[1],
		Long: pos[0],
	}
}

// Distance returns geographical distance between two coordinates in kilometers.
func Distance(a, b LonLat) float64 {
	return geodist.HaversineDistance(a.point(), b.point())
}
