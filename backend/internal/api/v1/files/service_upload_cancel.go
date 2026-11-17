package files

// CancelUpload discards an open upload session and everything it received.
func (s *Service) CancelUpload(uploadID string) error {
	if err := requireValidUploadID(uploadID); err != nil {
		return err
	}
	unlock := lockUploadSession(uploadID)
	defer unlock()

	if _, err := loadUploadSession(uploadID); err != nil {
		return err
	}
	return deleteUploadSession(uploadID)
}
