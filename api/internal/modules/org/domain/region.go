package domain

type Province struct {
	Code string
	Name string
}

type City struct {
	Code         string
	Name         string
	ProvinceName string
	Centroid     Coordinates
}
