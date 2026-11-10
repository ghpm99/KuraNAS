package trash

// MaxBatchItems caps how many trash items one batch request may address.
const MaxBatchItems = 500

// ItemFailure pairs a trash item id with the reason its operation failed.
type ItemFailure struct {
	ID  int
	Err error
}

// RestoreItems restores each item independently: one failure does not stop the
// others. Returns the restored ids and the failures, both in request order.
func (s *Service) RestoreItems(ids []int) ([]int, []ItemFailure) {
	return applyToEachItem(ids, func(id int) error {
		_, err := s.RestoreItem(id)
		return err
	})
}

// DeleteItemsPermanently purges each item independently: one failure does not
// stop the others. Returns the deleted ids and the failures, in request order.
func (s *Service) DeleteItemsPermanently(ids []int) ([]int, []ItemFailure) {
	return applyToEachItem(ids, s.DeleteItemPermanently)
}

func applyToEachItem(ids []int, operation func(id int) error) ([]int, []ItemFailure) {
	succeededIDs := make([]int, 0, len(ids))
	failures := []ItemFailure{}
	for _, id := range uniqueIDs(ids) {
		if err := operation(id); err != nil {
			failures = append(failures, ItemFailure{ID: id, Err: err})
			continue
		}
		succeededIDs = append(succeededIDs, id)
	}
	return succeededIDs, failures
}

func uniqueIDs(ids []int) []int {
	seenIDs := make(map[int]struct{}, len(ids))
	distinctIDs := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, isDuplicate := seenIDs[id]; isDuplicate {
			continue
		}
		seenIDs[id] = struct{}{}
		distinctIDs = append(distinctIDs, id)
	}
	return distinctIDs
}
