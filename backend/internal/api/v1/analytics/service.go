package analytics

import (
	"errors"
	"fmt"
	"nas-go/api/internal/config"
	"sort"
	"time"
)

var ErrInvalidPeriod = errors.New("invalid analytics period")

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (s *Service) GetOverview(period string) (OverviewDto, error) {
	periodConfig, err := resolvePeriod(period)
	if err != nil {
		return OverviewDto{}, err
	}

	data, err := s.Repository.GetOverviewData(periodConfig, OverviewLimits{
		RecentFiles:    50,
		TopExtensions:  12,
		TopFolders:     20,
		TopHotFolders:  3,
		TopDuplicates:  20,
		RecentLogError: 5,
	})
	if err != nil {
		return OverviewDto{}, err
	}

	totalBytes, freeBytes := getFileSystemStorage(data.StorageKpis.UsedBytes)

	overview := OverviewDto{
		Period:      periodConfig.Label,
		GeneratedAt: time.Now().UTC(),
		Storage: StorageDto{
			TotalBytes:  totalBytes,
			UsedBytes:   data.StorageKpis.UsedBytes,
			FreeBytes:   freeBytes,
			GrowthBytes: data.StorageKpis.GrowthBytes,
		},
		Counts: CountsDto{
			FilesTotal: data.StorageKpis.FilesTotal,
			FilesAdded: data.StorageKpis.FilesAdded,
			Folders:    data.StorageKpis.FoldersTotal,
		},
		TimeSeries:  toTimeSeriesDto(data.TimeSeries),
		Types:       toTypeBreakdownDto(data.Types),
		Extensions:  toExtensionDto(data.Extensions),
		HotFolders:  toHotFolderDto(data.HotFolders),
		TopFolders:  toFolderUsageDto(data.TopFolders),
		RecentFiles: toRecentFilesDto(data.RecentFiles),
		Duplicates: DuplicatesDto{
			Groups:          data.Duplicates.GroupsTotal,
			Files:           data.Duplicates.FilesTotal,
			ReclaimableSize: data.Duplicates.ReclaimableBytes,
			TopGroups:       toDuplicateGroupDto(data.TopDuplicateSets),
		},
		Health: toHealthDto(data),
	}

	return overview, nil
}

func resolvePeriod(period string) (PeriodConfig, error) {
	switch period {
	case "", "7d":
		return PeriodConfig{Label: "7d", Interval: "7 days"}, nil
	case "24h":
		return PeriodConfig{Label: "24h", Interval: "24 hours"}, nil
	case "30d":
		return PeriodConfig{Label: "30d", Interval: "30 days"}, nil
	case "90d":
		return PeriodConfig{Label: "90d", Interval: "90 days"}, nil
	default:
		return PeriodConfig{}, fmt.Errorf("%w: %s", ErrInvalidPeriod, period)
	}
}

func getFileSystemStorage(usedBytes int64) (int64, int64) {
	totalBytes, freeBytes, err := getFileSystemStats(config.AppConfig.EntryPoint)
	if err != nil {
		return usedBytes, 0
	}

	if totalBytes <= 0 {
		return usedBytes, freeBytes
	}
	return totalBytes, freeBytes
}

func toTimeSeriesDto(models []StorageTimeSeriesModel) []TimeSeriesPointDto {
	response := make([]TimeSeriesPointDto, 0, len(models))
	for _, item := range models {
		response = append(response, TimeSeriesPointDto{Date: item.Date.Format(time.DateOnly), UsedBytes: item.UsedBytes})
	}
	return response
}

func toTypeBreakdownDto(models []TypeDistributionModel) []TypeBreakdownDto {
	response := make([]TypeBreakdownDto, 0, len(models))
	for _, item := range models {
		response = append(response, TypeBreakdownDto{Type: item.Category, Count: item.Count, Bytes: item.Bytes})
	}
	return response
}

func toExtensionDto(models []ExtensionDistributionModel) []ExtensionDto {
	response := make([]ExtensionDto, 0, len(models))
	for _, item := range models {
		response = append(response, ExtensionDto{Ext: item.Extension, Count: item.Count, Bytes: item.Bytes})
	}
	return response
}

func toHotFolderDto(models []FolderHotModel) []HotFolderDto {
	response := make([]HotFolderDto, 0, len(models))
	for _, item := range models {
		responseItem := HotFolderDto{Path: item.ParentPath, NewFiles: item.NewFiles, AddedBytes: item.AddedBytes}
		if item.LastEvent.Valid {
			responseItem.LastEvent = item.LastEvent.Time.Format(time.RFC3339)
		}
		response = append(response, responseItem)
	}
	return response
}

func toFolderUsageDto(models []FolderUsageModel) []FolderUsageDto {
	response := make([]FolderUsageDto, 0, len(models))
	for _, item := range models {
		responseItem := FolderUsageDto{Path: item.ParentPath, Files: item.TotalFiles, Bytes: item.TotalBytes}
		if item.LastModified.Valid {
			responseItem.LastModified = item.LastModified.Time.Format(time.RFC3339)
		}
		response = append(response, responseItem)
	}
	return response
}

func toRecentFilesDto(models []RecentFileModel) []RecentFileDto {
	response := make([]RecentFileDto, 0, len(models))
	for _, item := range models {
		response = append(response, RecentFileDto{
			ID:         item.ID,
			Name:       item.Name,
			Path:       item.Path,
			ParentPath: item.ParentPath,
			Format:     item.Format,
			SizeBytes:  item.Size,
			CreatedAt:  item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  item.UpdatedAt.Format(time.RFC3339),
		})
	}
	return response
}

func toDuplicateGroupDto(models []DuplicateGroupModel) []DuplicateGroupDto {
	response := make([]DuplicateGroupDto, 0, len(models))
	for _, item := range models {
		response = append(response, DuplicateGroupDto{
			Signature:       item.Signature,
			Copies:          item.Copies,
			SizeBytes:       item.ItemSize,
			ReclaimableSize: item.ReclaimableSize,
			Paths:           item.Paths,
		})
	}
	return response
}

func toHealthDto(data OverviewDataModel) HealthDto {
	status := "ok"
	if data.HealthStatus.Valid {
		switch data.HealthStatus.String {
		case "PENDING":
			status = "scanning"
		case "FAILED":
			status = "error"
		default:
			status = "ok"
		}
	}

	lastScanAt := ""
	lastScanSeconds := int64(0)
	if data.LastScanStart.Valid {
		lastScanAt = data.LastScanStart.Time.Format(time.RFC3339)
	}
	if data.LastScanStart.Valid && data.LastScanEnd.Valid {
		lastScanSeconds = int64(data.LastScanEnd.Time.Sub(data.LastScanStart.Time).Seconds())
	}

	errorsList := make([]string, 0, len(data.RecentErrors))
	sort.SliceStable(data.RecentErrors, func(i, j int) bool {
		return data.RecentErrors[i].CreatedAt.After(data.RecentErrors[j].CreatedAt)
	})
	for _, item := range data.RecentErrors {
		description := item.Name
		if item.Description.Valid && item.Description.String != "" {
			description = fmt.Sprintf("%s: %s", item.Name, item.Description.String)
		}
		errorsList = append(errorsList, description)
	}

	return HealthDto{
		Status:          status,
		LastScanAt:      lastScanAt,
		LastScanSeconds: lastScanSeconds,
		IndexedFiles:    data.StorageKpis.FilesTotal,
		ErrorsLast24h:   data.ErrorsLast24h,
		RecentErrors:    errorsList,
	}
}
