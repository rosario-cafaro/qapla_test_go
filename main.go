// Qapla test script

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// SearchLocalisedStringIds
// Recursively search the interface for the string IDs of the localization map
func SearchLocalisedStringIds(obj map[string]interface{}, w http.ResponseWriter) ([]string, error) {
	var ids []string
	var err error

	for k, v := range obj {

		if k == "localisedStringId" {
			ids = append(ids, v.(string))
		}

		switch kind := reflect.TypeOf(v).Kind(); kind {
		case reflect.Slice:
			for _, innerV := range v.([]interface{}) {
				res, _ := SearchLocalisedStringIds(innerV.(map[string]interface{}), w)
				ids = append(ids, res...)
			}
		case reflect.Map:
			res, _ := SearchLocalisedStringIds(v.(map[string]interface{}), w)
			ids = append(ids, res...)
		}

	}

	return ids, err
}

func handler(w http.ResponseWriter, r *http.Request) {

	/**
	Constants
	*/
	const ApiUrl = "https://track.amazon.it/api/tracker/"
	const LocalizationUrl = "https://track.amazon.it/getLocalizedStrings"

	/**
	Localization map (Static fallback if the POST fails)
	*/
	localizationMap := map[string]string{
		"swa_rex_delivering_no_updated_eddday": "Consegnato",
		"swa_rex_detail_pickedUp":              "Pacco ritirato",
		"swa_rex_arrived_at_sort_center":       "Il pacco è arrivato presso la sede del corriere",
		"swa_rex_ofd":                          "In consegna",
		"swa_rex_detail_creation_confirmed":    "Etichetta creata",
		"swa_rex_shipping_label_created":       "Etichetta creata",
		"swa_rex_detail_departed":              "Il pacco ha lasciato la sede del corriere",
	}

	/**
	Get the parameters
	*/
	var trackingNumber string
	var jsonFlag int64
	trackingNumber = r.URL.Query().Get("tracking")
	jsonFlag, _ = strconv.ParseInt(r.URL.Query().Get("json"), 10, 64)

	/**
	Check mandatory parameter/s
	*/
	if len(trackingNumber) <= 0 {
		if jsonFlag == 1 {
			//w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Set("Content-Type", "application/json")
			mapResponse := map[string]interface{}{
				"response": 422,
				"msg":      "Missing Tracking number!",
			}
			jsonResponse, _ := json.Marshal(mapResponse)
			_, err := fmt.Fprintf(w, string(jsonResponse))
			if err != nil {
				_, err = fmt.Fprintf(w, "000 %v", err)
				return
			}

		} else {
			_, err := fmt.Fprintf(w, "Missing Tracking number!")
			if err != nil {
				_, err = fmt.Fprintf(w, "001 %v", err)
				return
			}
		}
		return
	}

	/**
	 * Get the page/JSON content
	 */
	res, err := http.Get(ApiUrl + trackingNumber)
	if err != nil {
		_, err = fmt.Fprintf(w, "003 %v", err)
		return
	}
	content, err := io.ReadAll(res.Body)
	if err != nil {
		_, err = fmt.Fprintf(w, "004 %v", err)
		return
	}
	err = res.Body.Close()
	if err != nil {
		_, err = fmt.Fprintf(w, "005 %v", err)
		return
	}
	amazonTrackerApiResponse := content
	var payload interface{}
	err = json.Unmarshal(amazonTrackerApiResponse, &payload)
	if err != nil {
		_, err = fmt.Fprintf(w, "006 %v", err)
		return
	}

	apiResponseMap := payload.(map[string]interface{})

	/**
	Init the localization keys array
	*/
	var localizationKeys []string

	/**
	Decode inner elements
	*/
	var progressTrackerPayload interface{}
	err = json.Unmarshal([]byte(apiResponseMap["progressTracker"].(string)), &progressTrackerPayload)
	if err != nil {
		_, err = fmt.Fprintf(w, "008 %v", err)
		return
	}
	progressTracker := progressTrackerPayload.(map[string]interface{})

	var eventHistoryPayload interface{}
	err = json.Unmarshal([]byte(apiResponseMap["eventHistory"].(string)), &eventHistoryPayload)
	if err != nil {
		_, err = fmt.Fprintf(w, "009 %v", err)
		return
	}
	eventHistory := eventHistoryPayload.(map[string]interface{})

	/**
	Add the keys that needs to be translated
	*/
	ids, err := SearchLocalisedStringIds(progressTracker["progressMeter"].(map[string]interface{}), w)
	if err != nil {
		_, err = fmt.Fprintf(w, "010 %v", err)
		return
	}
	localizationKeys = append(localizationKeys, ids...)

	ids, err = SearchLocalisedStringIds(eventHistory, w)
	if err != nil {
		_, err = fmt.Fprintf(w, "011 %v", err)
		return
	}
	localizationKeys = append(localizationKeys, ids...)

	/**
	Get the localizations' values
	*/
	postPayloadData := map[string]interface{}{
		"localizationKeys": localizationKeys,
	}
	postPayload, err := json.Marshal(postPayloadData)

	if err != nil {
		_, err = fmt.Fprintf(w, "012 %v", err)
		return
	}

	req, err := http.NewRequest("POST", LocalizationUrl, bytes.NewBuffer(postPayload))
	if err != nil {
		_, err = fmt.Fprintf(w, "013 %v", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anti-csrftoken-a2z", "hMOny/PlioLqIfCNSnGcF2OEGuNO1XVbQmal8o6eICdsAAAAAGV4w9kAAAAB")
	req.Header.Set("Origin", "https://track.amazon.it")
	req.Header.Set("Connection", "keep-alive")
	referer := strings.Join([]string{LocalizationUrl, trackingNumber}, "")
	req.Header.Set("Referer", referer)
	req.Header.Set("Cookie", "session-id=257-5784047-3884732; session-id-time=2082787201l; csm-hit=tb:s-VYH1DKCJS293CC44TVHZ|1702413273725&t:1702413274591&adb:adblk_no; ubid-acbit=258-7911589-2663967; session-token=9/TU7XVBwIubpQKjYyLs4bsqMYKoO1cs30OnB8f4aZAnRxEd9nSI+E7DmR+62uZnUNtPP/pghsTqeIWUuCKCqpGpAJfDnycClA6/DFDMvf62rsur5ayeC0YbvhHLRXq+ac1wkulN1oitnWp8xEGJuwOhnH78MNhvNBqiaFzm1ukmuBnZhE6ft40A5DbXR86M3h3wvIEF/qHdjJCg6mN+kSEocqhiCKwicTE508pkO90wRCQUp4AmCuDVNE1yE9r5pZFb1LvM1I9cFMv5LR5fD/UZ3MLCtrxkLxUYiRb+cTCKSeZlvB5UVuenTGfgi49gwEixWxU/16DhqLQXRWlQtuc+Y1yNc8dZ")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		_, err = fmt.Fprintf(w, "014 %v", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_, err = fmt.Fprintf(w, "015 %v", err)
			return
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	if err != nil {
		_, err = fmt.Fprintf(w, "016 %v", err)
		return
	}

	/**
	 * If the dynamic slice is empty, use the fallback static array
	 */
	postResponse := map[string]string{}
	err = json.Unmarshal(body, &postResponse)
	//if err != nil {
	//	_, err = fmt.Fprintf(w, "017 %v", err)
	//	return
	//}

	localizations := map[string]string{}
	if len(postResponse) >= 1 {
		localizations = postResponse
	} else {
		localizations = localizationMap
	}

	/**
	Build the solution output
	*/
	response := make(map[string]interface{})
	response["shipper"] = map[string]interface{}{
		"label": "Ordine effettuato presso",
		"value": apiResponseMap["shipperDetails"].(map[string]interface{})["shipperName"].(string),
	}

	response["expectedDeliveryDate"] = map[string]interface{}{
		"label": "Data di consegna prevista",
		"value": progressTracker["expectedDeliveryDate"].(string),
	}

	statusString := strings.Join([]string{progressTracker["summary"].(map[string]interface{})["metadata"].(map[string]interface{})["trackingStatus"].(map[string]interface{})["stringValue"].(string), " (", progressTracker["summary"].(map[string]interface{})["status"].(string), ")"}, "")
	response["status"] = map[string]interface{}{
		"label": "Stato",
		"value": statusString,
	}

	var historyValue []map[string]interface{}
	response["history"] = map[string]interface{}{
		"label": "Storico",
		"value": historyValue,
	}

	for _, event := range eventHistory["eventHistory"].([]interface{}) {
		statusString = localizations[event.(map[string]interface{})["statusSummary"].(map[string]interface{})["localisedStringId"].(string)]
		if len(statusString) <= 0 {
			statusString = event.(map[string]interface{})["eventCode"].(string)
		}
		tmp := map[string]interface{}{
			"Stato spedizione": statusString,
			"Data":             event.(map[string]interface{})["eventTime"].(string),
		}

		city := event.(map[string]interface{})["location"].(map[string]interface{})["city"]
		if city != nil {
			stateProvince := event.(map[string]interface{})["location"].(map[string]interface{})["stateProvince"]
			countryCode := event.(map[string]interface{})["location"].(map[string]interface{})["countryCode"]
			postalCode := event.(map[string]interface{})["location"].(map[string]interface{})["postalCode"]
			address := strings.Join([]string{city.(string), stateProvince.(string), countryCode.(string), postalCode.(string)}, ", ")
			tmp["Luogo"] = address
		}

		response["history"].(map[string]interface{})["value"] = append(response["history"].(map[string]interface{})["value"].([]map[string]interface{}), tmp)
	}

	/**
	Output the response based on the input parameters
	*/
	if jsonFlag == 1 {
		//w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(map[string]interface{}{
			"response": 200,
			"data":     response,
		})
		_, err := fmt.Fprintf(w, string(jsonResponse))
		if err != nil {
			_, err = fmt.Fprintf(w, "018 %v", err)
			return
		}

	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		for _, responseV := range response {
			_, err = fmt.Fprintf(w, "<strong>%v: </strong>", responseV.(map[string]interface{})["label"])
			if reflect.TypeOf(responseV.(map[string]interface{})["value"]).Kind() == reflect.Slice {
				_, err = fmt.Fprintf(w, "<br/>")
				for _, el := range responseV.(map[string]interface{})["value"].([]map[string]interface{}) {
					for elK, elV := range el {
						_, err = fmt.Fprintf(w, "&emsp;<strong>%v: </strong> %v<br/>", elK, elV)
					}
					_, err = fmt.Fprintf(w, "<br/>")
				}
			} else {
				_, err = fmt.Fprintf(w, "%v<br/>", responseV.(map[string]interface{})["value"])
			}
		}

	}

}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
