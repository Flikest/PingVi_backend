package geocoder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type GeocoderResponse struct {
	Response struct {
		GeoObjectCollection struct {
			FeatureMember []struct {
				GeoObject struct {
					MetaDataProperty struct {
						GeocoderMetaData struct {
							Address struct {
								Components []struct {
									Kind string `json:"kind"`
									Name string `json:"name"`
								} `json:"Components"`
							} `json:"Address"`
							Text string `json:"text"`
						} `json:"GeocoderMetaData"`
					} `json:"metaDataProperty"`
					Name string `json:"name"`
				} `json:"GeoObject"`
			} `json:"featureMember"`
		} `json:"GeoObjectCollection"`
	} `json:"response"`
}

type LocationInfo struct {
	Country     string
	City        string
	FullAddress string
}

func Geocoder(geocode string, lang string) (LocationInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	apiKey := os.Getenv("GEOCODER_API_KEY")
	if apiKey == "" {
		return LocationInfo{}, fmt.Errorf("GEOCODER_API_KEY environment variable is not set")
	}

	langParam := "en_US"
	if strings.HasPrefix(lang, "ru") {
		langParam = "ru_RU"
	}

	baseUrl := fmt.Sprintf("https://geocode-maps.yandex.ru/v1/?apikey=%s&geocode=%s&format=json&lang=%s&results=1",
		apiKey, geocode, langParam)

	resp, err := client.Get(baseUrl)
	if err != nil {
		return LocationInfo{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return LocationInfo{}, fmt.Errorf("API returned status: %s", resp.Status)
	}

	var geocoderResponse GeocoderResponse

	err = json.NewDecoder(resp.Body).Decode(&geocoderResponse)
	if err != nil {
		return LocationInfo{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geocoderResponse.Response.GeoObjectCollection.FeatureMember) == 0 {
		return LocationInfo{}, fmt.Errorf("no results found for geocode: %s", geocode)
	}

	components := geocoderResponse.Response.GeoObjectCollection.FeatureMember[0].GeoObject.MetaDataProperty.GeocoderMetaData.Address.Components

	var country, city string

	for _, component := range components {
		switch component.Kind {
		case "country":
			country = component.Name
		case "province", "area", "locality":
			if city == "" {
				city = component.Name
			}
		}
	}

	fullAddress := geocoderResponse.Response.GeoObjectCollection.FeatureMember[0].GeoObject.MetaDataProperty.GeocoderMetaData.Text

	return LocationInfo{
		Country:     country,
		City:        city,
		FullAddress: fullAddress,
	}, nil
}
