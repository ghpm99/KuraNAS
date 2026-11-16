package vacuum

import "math"

func integerDataPoint(dataPoints map[string]any, dataPointID string) (int, bool) {
	switch reported := dataPoints[dataPointID].(type) {
	case float64:
		return int(math.Round(reported)), true
	case int:
		return reported, true
	default:
		return 0, false
	}
}

func integerDataPointOrZero(dataPoints map[string]any, dataPointID string) int {
	reported, _ := integerDataPoint(dataPoints, dataPointID)
	return reported
}

func optionalIntegerDataPoint(dataPoints map[string]any, dataPointID string) *int {
	reported, isPresent := integerDataPoint(dataPoints, dataPointID)
	if !isPresent {
		return nil
	}
	return &reported
}

func optionalTextDataPoint(dataPoints map[string]any, dataPointID string) *string {
	reported, isText := dataPoints[dataPointID].(string)
	if !isText {
		return nil
	}
	return &reported
}

func optionalBooleanDataPoint(dataPoints map[string]any, dataPointID string) *bool {
	reported, isBoolean := dataPoints[dataPointID].(bool)
	if !isBoolean {
		return nil
	}
	return &reported
}
