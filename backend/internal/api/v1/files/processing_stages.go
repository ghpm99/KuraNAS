package files

// ProcessingStages tells which optional upload processing stages are enabled.
type ProcessingStages struct {
	ExtractMetadata  bool
	GeneratePreviews bool
}

// ProcessingStagesReader gives the upload flow the current processing stages.
type ProcessingStagesReader interface {
	ProcessingStages() ProcessingStages
}

func allProcessingStagesEnabled() ProcessingStages {
	return ProcessingStages{ExtractMetadata: true, GeneratePreviews: true}
}

// SetProcessingStagesReader wires the source of the processing stages in after
// construction. Without a reader every stage is enabled.
func (s *Service) SetProcessingStagesReader(reader ProcessingStagesReader) {
	s.ProcessingStagesReader = reader
}

func (s *Service) currentProcessingStages() ProcessingStages {
	if s.ProcessingStagesReader == nil {
		return allProcessingStagesEnabled()
	}
	return s.ProcessingStagesReader.ProcessingStages()
}
