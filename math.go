package bkit

import "math"

var Math = &MathUtil{}

type MathUtil struct{}

const (
	earthRadiusMi = 3958 // 地球半径（单位：英里）。
	earthRaidusKm = 6371 // 地球半径（单位：千米）。
)

// Coord 表示地理坐标。
type Coord struct {
	Lat float64 // 纬度
	Lon float64 // 经度
}

// degreesToRadians 将角度转换为弧度。
func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

// Distance 计算地球表面上两个坐标之间的最短路径距离。
// 使用Haversine公式计算两个坐标之间的距离。
// 此函数返回两个单位的测量结果，第一个是以英里为单位的距离，第二个是以千米为单位的距离。
func (mu *MathUtil) Distance(p, q Coord) (mi, km float64) {
	lat1 := degreesToRadians(p.Lat)
	lon1 := degreesToRadians(p.Lon)
	lat2 := degreesToRadians(q.Lat)
	lon2 := degreesToRadians(q.Lon)

	diffLat := lat2 - lat1
	diffLon := lon2 - lon1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	mi = c * earthRadiusMi
	km = c * earthRaidusKm

	return mi, km
}

// DistanceMeter 计算地球表面上两个坐标之间的最短路径距离（单位：米）。
func (mu *MathUtil) DistanceMeter(p, q Coord) float64 {
	_, km := mu.Distance(p, q)
	return km * 1000
}
