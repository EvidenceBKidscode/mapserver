package geometry

type Geometry struct {
	EpsgCarto        string        `json:"epsgCarto"`
	AngleDegres      float32       `json:"angleDegres"`
	CoordinatesCarto [4][2]float32 `json:"coordinatesCarto"`
	CoordinatesGeo   [4][2]float32 `json:"coordinatesGeo"`
	CoordinatesGame  [4][2]float32 `json:"coordinatesGame"`
	Echelle          float32       `json:"echelle"`
	AltitudeZero     int           `json:"altitudeZero"`
}
