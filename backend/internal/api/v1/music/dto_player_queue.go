package music

const maxPlayerQueueEntries = maxQueueEntries

type ReplacePlayerQueueRequest struct {
	FileIDs      []int `json:"file_ids"`
	CurrentIndex int   `json:"current_index"`
}

type PlayerQueueDto struct {
	Items        []MusicQueueEntryDto `json:"items"`
	CurrentIndex int                  `json:"current_index"`
}

func (request ReplacePlayerQueueRequest) isValid() bool {
	if len(request.FileIDs) > maxPlayerQueueEntries {
		return false
	}
	if request.CurrentIndex < 0 {
		return false
	}
	if len(request.FileIDs) == 0 {
		return request.CurrentIndex == 0
	}
	return request.CurrentIndex < len(request.FileIDs)
}
