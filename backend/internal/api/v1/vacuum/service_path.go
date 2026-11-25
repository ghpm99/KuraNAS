package vacuum

import "nas-go/api/pkg/tuya"

func (s *Service) GetPath() PathDto {
	return s.pathTracker.snapshot()
}

func (s *Service) acceptPush(dataPoints map[string]any) {
	encodedReport, isText := dataPoints[dpPath].(string)
	if !isText {
		return
	}
	s.pathTracker.applyReport(encodedReport)
}

func (s *Service) deviceToListenTo() (device tuya.Device, canListen bool, err error) {
	settings, err := s.loadSettings()
	if err != nil {
		return tuya.Device{}, false, err
	}
	return settings.toDevice(), settings.isConfigured() && settings.Address != "", nil
}
