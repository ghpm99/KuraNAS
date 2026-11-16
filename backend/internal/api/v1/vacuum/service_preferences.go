package vacuum

import "context"

func (s *Service) GetPreferences(ctx context.Context) (PreferencesDto, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return PreferencesDto{}, err
	}
	return preferencesFromDataPoints(snapshot.values), nil
}

func (s *Service) UpdatePreferences(ctx context.Context, patch UpdatePreferencesDto) (PreferencesDto, error) {
	if err := validatePreferencesPatch(patch); err != nil {
		return PreferencesDto{}, err
	}

	changedDataPoints := dataPointsFromPatch(patch)
	if len(changedDataPoints) == 0 {
		return PreferencesDto{}, ErrInvalidRequest
	}

	if err := s.writeDataPoints(ctx, changedDataPoints); err != nil {
		return PreferencesDto{}, err
	}
	return s.GetPreferences(ctx)
}

func preferencesFromDataPoints(dataPoints map[string]any) PreferencesDto {
	return PreferencesDto{
		Suction:             optionalTextDataPoint(dataPoints, dpSuction),
		WaterLevel:          optionalTextDataPoint(dataPoints, dpWaterLevel),
		Volume:              optionalIntegerDataPoint(dataPoints, dpVolume),
		IsDoNotDisturb:      optionalBooleanDataPoint(dataPoints, dpDoNotDisturb),
		IsCarpetBoost:       optionalBooleanDataPoint(dataPoints, dpCarpetBoost),
		IsResumeAfterCharge: optionalBooleanDataPoint(dataPoints, dpResumeAfterCharge),
		IsAutoEmpty:         optionalBooleanDataPoint(dataPoints, dpAutoEmpty),
		AutoEmptyFrequency:  optionalTextDataPoint(dataPoints, dpAutoEmptyFrequency),
	}
}

func dataPointsFromPatch(patch UpdatePreferencesDto) map[string]any {
	changedDataPoints := map[string]any{}
	if patch.Suction != nil {
		changedDataPoints[dpSuction] = *patch.Suction
	}
	if patch.WaterLevel != nil {
		changedDataPoints[dpWaterLevel] = *patch.WaterLevel
	}
	if patch.Volume != nil {
		changedDataPoints[dpVolume] = *patch.Volume
	}
	if patch.IsDoNotDisturb != nil {
		changedDataPoints[dpDoNotDisturb] = *patch.IsDoNotDisturb
	}
	if patch.IsCarpetBoost != nil {
		changedDataPoints[dpCarpetBoost] = *patch.IsCarpetBoost
	}
	if patch.IsResumeAfterCharge != nil {
		changedDataPoints[dpResumeAfterCharge] = *patch.IsResumeAfterCharge
	}
	if patch.IsAutoEmpty != nil {
		changedDataPoints[dpAutoEmpty] = *patch.IsAutoEmpty
	}
	if patch.AutoEmptyFrequency != nil {
		changedDataPoints[dpAutoEmptyFrequency] = *patch.AutoEmptyFrequency
	}
	return changedDataPoints
}
