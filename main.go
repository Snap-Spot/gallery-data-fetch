package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type JSONBody struct {
	Response struct {
		Header struct {
			ResultCode string
			ResultMsg  string
		}
		Body struct {
			Items struct {
				Item []struct {
					GalContentId         string
					GalContentTypeId     string
					GalTitle             string
					GalWebImageUrl       string
					GalCreatedtime       string
					GalModifiedtime      string
					GalPhotographyMonth  string
					GalPhotographyLocation string
					GalPhotographer      string 
					GalSearchKeyword     string
				}
			}
			NumOfRows int 
			PageNo    int
			TotalCount int
		}
	}
}

func GetConnector() *sql.DB {
	config := mysql.Config{
		User: "efub2023",
		Passwd: "vjqltmsoq1886",
		Net: "tcp",
		Addr: "snapspot-db.cechzp23mskp.ap-northeast-2.rds.amazonaws.com:3306",
		Collation: "utf8mb4_general_ci",
		Loc: time.UTC,
		MaxAllowedPacket: 4 << 20.,
		AllowNativePasswords: true,
		CheckConnLiveness: true,
		DBName: "snapspot",
	}
	connector, err := mysql.NewConnector(&config)
	if err != nil {
		panic(err)
	}
	db := sql.OpenDB(connector)
	return db
}

var client = &http.Client{Timeout: 100 * time.Second}

func getJSON(url string) string {
	r, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
        panic(err)
    }

	return string(body)
}


func main() {
	db := GetConnector()
	err := db.Ping()
	if err != nil {
		panic(err)
	}

	for i := 1; i <=160; i++ {
		var city string
		err = db.QueryRow("SELECT city FROM area WHERE area_id = " + strconv.Itoa(i)).Scan(&city)
		if err != nil {
			log.Fatal(err)
		}
		var arr = strings.Split(city, "/")

		for j := 0; j < len(arr); j++ {
			encodedArea := url.QueryEscape(arr[j])
			fmt.Println(arr[j], encodedArea)

			url := "https://apis.data.go.kr/B551011/PhotoGalleryService1/gallerySearchList1?serviceKey=5bn7JWYhnLorQIQl%2B4uFb88GRUdQuZHXnK2zXxazlDor2CKeQ8qVjPCYwm%2F%2FJh7eCfHfIv1%2FiP6I15AsdVJUHw%3D%3D&_type=json&MobileOS=ETC&MobileApp=Snap&numOfRows=150&keyword=" + encodedArea
			var body JSONBody
			err := json.Unmarshal([]byte(getJSON(url)), &body)
			if err != nil {
				fmt.Println("JSON 파싱 오류:", err)
			}

			fmt.Println("Items:")
			for _, item := range body.Response.Body.Items.Item {
				fmt.Printf("  GalContentId: %s\n", item.GalContentId)
				fmt.Printf("  GalTitle: %s\n", item.GalTitle)
				fmt.Printf("  GalWebImageUrl: %s\n", item.GalWebImageUrl)
				fmt.Printf("  GalPhotographyMonth: %s\n", item.GalPhotographyMonth)
				fmt.Printf("  GalPhotographyLocation: %s\n", item.GalPhotographyLocation)
				fmt.Printf("  GalPhotographer: %s\n", item.GalPhotographer)
				fmt.Printf("  GalSearchKeyword: %s\n", item.GalSearchKeyword)
				fmt.Println("------------------------")
				query := fmt.Sprintf("insert into area_image (area_id, title, location, photographer, url) value (%d, '%s', '%s', '%s', '%s')", i, item.GalTitle, item.GalPhotographyLocation, item.GalPhotographer, item.GalWebImageUrl)
				log.Println("query: ", query)
				result, err := db.Exec(query)
				if err != nil {
					panic(err)
				}
				log.Println(result)
			}
		}
	}
	
}