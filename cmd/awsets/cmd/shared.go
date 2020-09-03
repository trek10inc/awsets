package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/trek10inc/awsets/resource"
)

type IdLite struct {
	Region string
	Id     string
	//Version string
	Type resource.ResourceType
}

func loadData(fname string) (map[IdLite]resource.Resource, error) {
	var resources []resource.Resource
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &resources)
	if err != nil {
		return nil, err
	}

	res := make(map[IdLite]resource.Resource)

	for i := range resources {
		r := resources[i]
		id := genId(r.Identifier)
		if _, exists := res[id]; exists {
			fmt.Printf("Hm... already exists - %v\n", id)
		}
		res[id] = r
	}

	return res, nil
}

func genId(identifier resource.Identifier) IdLite {
	return IdLite{
		Region: identifier.Region,
		Id:     identifier.Id,
		Type:   identifier.Type,
		//Version: identifier.Version,
	}
}
