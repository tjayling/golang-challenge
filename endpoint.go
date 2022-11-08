package main

import (
	"database/sql"
	"fmt"
	"math"
	"strings"

	_ "github.com/lib/pq"
)

type Spot struct {
	ID          []uint8
	name        string
	website     sql.NullString
	x           float64
	y           float64
	description sql.NullString
	rating      float64
	distance    float64
}

type Spots []Spot

func main() {
	results := endpoint(51.4935, 0.1178, 7000, "square")
	for i, spot := range results {
		fmt.Printf("%d: Name:%s Website: %s Location: %f, %f Description: %s, rating %f", i+1, spot.name, spot.website.String, spot.x, spot.y, spot.description.String, spot.rating)
		fmt.Println()
	}
}

func endpoint(lat float64, lon float64, rad float64, areaType string) Spots {
	sql := generateSql(lat, lon, rad, areaType)
	if sql == "" {
		return nil
	}
	spots := getSpots(lat, lon, sql)
	sortedSpots := quickSortStart(spots)
	result := finalSort(sortedSpots)
	return result
}

func finalSort(sortedSpots Spots) Spots {
	for i := 1; i < len(sortedSpots); i++ {
		if sortedSpots[i].distance-sortedSpots[i-1].distance < 0.0004498870783 && sortedSpots[i].rating > sortedSpots[i-1].rating {
			sortedSpots[i], sortedSpots[i-1] = sortedSpots[i-1], sortedSpots[i]
		}
	}
	return sortedSpots
}

func generateSql(lat float64, lon float64, rad float64, areaType string) string {
	switch strings.ToLower(areaType) {
	case "square":
		rad /= 111139
		p1x := lon - float64(rad)
		p1y := lat - float64(rad)
		p2x := lon + float64(rad)
		p2y := lat - float64(rad)
		p3x := lon + float64(rad)
		p3y := lat + float64(rad)
		p4x := lon - float64(rad)
		p4y := lat + float64(rad)
		rad *= 111139
		return fmt.Sprintf(`SELECT id, name, website, ST_X(coordinates::geometry), ST_Y(coordinates::geometry), description, rating FROM "MY_TABLE" WHERE ST_Intersects(ST_GEOMFROMTEXT('POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))'), "MY_TABLE".coordinates);`, p1x, p1y, p2x, p2y, p3x, p3y, p4x, p4y, p1x, p1y)
	case "circle":
		return fmt.Sprintf(`SELECT id, name, website, ST_X(coordinates::geometry), ST_Y(coordinates::geometry), description, rating FROM "MY_TABLE" WHERE ST_DistanceSphere("MY_TABLE".coordinates::geometry, ST_MakePoint(%f, %f)) <= %f;`, lon, lat, rad)
	default:
		fmt.Println("Area type not valid")
		return ""
	}
}

func getSpots(lat float64, lon float64, sql string) Spots {
	db := connect()
	data, err := db.Query(sql)

	if err != nil {
		panic(err)
	}
	var spots Spots
	count := 0
	for data.Next() {
		var eachSpot Spot
		err = data.Scan(&eachSpot.ID, &eachSpot.name, &eachSpot.website, &eachSpot.x, &eachSpot.y, &eachSpot.description, &eachSpot.rating)
		if err != nil {
			panic(err)
		}
		spots = append(spots, eachSpot)
		count++
	}

	for index, spot := range spots {
		spots[index].distance = getDist(spot.x, spot.y, lon, lat)
	}
	return spots
}

func getDist(ax float64, ay float64, bx float64, by float64) float64 {
	xDist := ax - bx
	yDist := ay - by
	return math.Sqrt((xDist * xDist) + (yDist * yDist))
}

func partition(arr Spots, low, high int) (Spots, int) {
	pivot := arr[high]
	i := low
	for j := low; j < high; j++ {
		if arr[j].distance < pivot.distance {
			arr[i], arr[j] = arr[j], arr[i]
			i++
		}
	}
	arr[i], arr[high] = arr[high], arr[i]
	return arr, i
}

func quickSort(arr Spots, low, high int) Spots {
	if low < high {
		var p int
		arr, p = partition(arr, low, high)
		arr = quickSort(arr, low, p-1)
		arr = quickSort(arr, p+1, high)
	}
	return arr
}

func quickSortStart(arr Spots) Spots {
	low, high := 0, len(arr)-1
	return quickSort(arr, low, high)
}

func connect() *sql.DB {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "pass"
		dbname   = "postgis_31_sample"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}
