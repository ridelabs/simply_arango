package encoding

import (
	"github.com/mitchellh/mapstructure"
)

func ObjectToMap(obj interface{}) (map[string]interface{}, error) {
	r := make(map[string]interface{})
	config := mapstructure.DecoderConfig{
		TagName: "json",
		Squash:  true,
		Result:  &r,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(obj); err != nil {
		return nil, err
	}

	return r, nil
}

func MapToObject(r map[string]interface{}, obj interface{}) error {
	config := mapstructure.DecoderConfig{
		TagName: "json",
		Squash:  true,
		Result:  obj,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(r); err != nil {
		return err
	}

	return nil
}
