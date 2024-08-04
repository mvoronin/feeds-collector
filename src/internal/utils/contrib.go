package utils

import (
	"errors"
	"feedscollector/internal/models"
	"fmt"
	"reflect"
)

func CompareFeedChannels(channel1, channel2 *models.FeedChannel, variables []string) (bool, string, error) {
	if channel1 == nil || channel2 == nil {
		return false, "", errors.New("one or both channels are nil")
	}

	v1 := reflect.ValueOf(channel1).Elem()
	v2 := reflect.ValueOf(channel2).Elem()

	for _, variable := range variables {
		field1 := v1.FieldByName(variable)
		field2 := v2.FieldByName(variable)

		if !field1.IsValid() || !field2.IsValid() {
			return false, variable, fmt.Errorf("variable %s not found in FeedChannel", variable)
		}

		if !reflect.DeepEqual(field1.Interface(), field2.Interface()) {
			return false, variable, nil
		}
	}

	return true, "", nil
}
