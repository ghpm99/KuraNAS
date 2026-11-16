package vacuum

import (
	"context"
	"fmt"
)

func (s *Service) GetConsumables(ctx context.Context) (ConsumablesDto, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return ConsumablesDto{}, err
	}

	reportedParts := []ConsumableDto{}
	for _, part := range consumableParts {
		seconds, isReported := integerDataPoint(snapshot.values, part.dataPointID)
		if isReported {
			reportedParts = append(reportedParts, ConsumableDto{Part: part.name, Seconds: seconds, MaxSeconds: part.maxSeconds})
		}
	}
	return ConsumablesDto{Items: reportedParts}, nil
}

func (s *Service) ResetConsumable(ctx context.Context, partName string) error {
	for _, part := range consumableParts {
		if part.name == partName {
			return s.writeDataPoints(ctx, dataPointWrite{dpConsumableReset, part.resetValue}.asPayload())
		}
	}
	return fmt.Errorf("%w: %q", ErrInvalidPart, partName)
}

func (s *Service) GetTotals(ctx context.Context) (TotalsDto, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return TotalsDto{}, err
	}
	return TotalsDto{
		CleaningSeconds: integerDataPointOrZero(snapshot.values, dpTotalCleaningSeconds),
		CleaningArea:    integerDataPointOrZero(snapshot.values, dpTotalCleaningArea),
		CleaningCount:   integerDataPointOrZero(snapshot.values, dpTotalCleaningCount),
	}, nil
}
