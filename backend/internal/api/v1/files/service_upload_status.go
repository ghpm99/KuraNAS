package files

type UploadSessionStatus struct {
	UploadID     string
	Size         int64
	ReceivedSize int64
}

// GetUploadSessionStatus reports how many bytes of an open upload the server
// already holds, so a client can resume after a restart.
func (s *Service) GetUploadSessionStatus(uploadID string) (UploadSessionStatus, error) {
	session, err := loadUploadSession(uploadID)
	if err != nil {
		return UploadSessionStatus{}, err
	}
	return UploadSessionStatus{UploadID: session.UploadID, Size: session.Size, ReceivedSize: session.ReceivedSize}, nil
}
