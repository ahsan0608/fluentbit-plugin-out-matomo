package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"C"

	"github.com/fluent/fluent-bit-go/output"
)

import "unsafe"

var (
	idsite string
	url    string
)

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	return output.FLBPluginRegister(def, "matomo", "Send logs to Matomo analytics")
}

//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	// Retrieve configuration parameters
	idsite = output.FLBPluginConfigKey(plugin, "idsite")
	if idsite == "" {
		fmt.Println("Error: missing 'idsite' parameter")
		return output.FLB_ERROR
	}

	url = output.FLBPluginConfigKey(plugin, "url")
	if url == "" {
		fmt.Println("Error: missing 'url' parameter")
		return output.FLB_ERROR
	}

	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(plugin unsafe.Pointer, data unsafe.Pointer, length C.int, tag *C.char) int {
	// Convert C data to Go slice
	cdata := C.GoBytes(data, length)

	// Parse Fluent Bit data into JSON
	var parsed interface{}
	err := json.Unmarshal(cdata, &parsed)
	if err != nil {
		fmt.Println(err)
		return output.FLB_ERROR
	}

	// Convert JSON to Matomo request format
	req := make(map[string]interface{})
	req["idsite"] = idsite
	req["rec"] = 1
	req["apiv"] = 1
	req["send_image"] = 0

	events := make([]map[string]interface{}, 0)
	for _, item := range parsed.([]interface{}) {
		event := make(map[string]interface{})
		event["action_name"] = item.(map[string]interface{})["log"].(string)
		events = append(events, event)
	}
	req["requests"] = []interface{}{events}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err)
		return output.FLB_ERROR
	}

	// Send request to Matomo
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		fmt.Println(err)
		return output.FLB_RETRY
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Println("Error: ", resp.StatusCode)
		return output.FLB_ERROR
	}

	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() {
}

func main() {
}
