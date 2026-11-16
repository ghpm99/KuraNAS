package vacuum

import (
	"context"

	"nas-go/api/pkg/applog"
)

func (s *Service) GetStatus(ctx context.Context) (StatusDto, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return StatusDto{}, err
	}

	return StatusDto{
		State:           stateFromDataPoints(snapshot.values),
		BatteryPercent:  integerDataPointOrZero(snapshot.values, dpBatteryPercent),
		CleaningSeconds: integerDataPointOrZero(snapshot.values, dpCleaningSeconds),
		CleaningArea:    integerDataPointOrZero(snapshot.values, dpCleaningArea),
		FaultCodes:      faultCodesFromBitmap(integerDataPointOrZero(snapshot.values, dpFaultBitmap)),
		ReadAt:          snapshot.readAt,
	}, nil
}

func stateFromDataPoints(dataPoints map[string]any) string {
	reportedState := optionalTextDataPoint(dataPoints, dpState)
	if reportedState == nil {
		return unknownState
	}

	state, isKnown := stateByDeviceValue[*reportedState]
	if !isKnown {
		applog.Warn("vacuum: device reported an unmapped state", "reported_state", *reportedState)
		return unknownState
	}
	return state
}

func faultCodesFromBitmap(bitmap int) []int {
	faultCodes := []int{}
	for bitIndex := 0; bitIndex < maxFaultCode; bitIndex++ {
		if bitmap&(1<<bitIndex) != 0 {
			faultCodes = append(faultCodes, bitIndex+1)
		}
	}
	return faultCodes
}
