package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/ai"
	"sort"
	"strings"
	"time"
)

var ErrInvalidPeriod = errors.New("invalid analytics period")

type Service struct {
	Repository RepositoryInterface
	AIService  ai.ServiceInterface
}

func NewService(repository RepositoryInterface, aiService ai.ServiceInterface) ServiceInterface {
	return &Service{Repository: repository, AIService: aiService}
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
		Library: LibraryDto{
			CategorizedMedia:  data.LibrarySummary.CategorizedMedia,
			AudioWithMetadata: data.LibrarySummary.AudioWithMetadata,
			VideoWithMetadata: data.LibrarySummary.VideoWithMetadata,
			ImageWithMetadata: data.LibrarySummary.ImageWithMetadata,
			ImageClassified:   data.LibrarySummary.ImageClassified,
		},
		Processing: ProcessingDto{
			MetadataPending:  data.Processing.MetadataPending,
			MetadataFailed:   data.Processing.MetadataFailed,
			ThumbnailPending: data.Processing.ThumbnailPending,
			ThumbnailFailed:  data.Processing.ThumbnailFailed,
		},
		Health: toHealthDto(data),
	}
	overview.Insights = s.generateInsights(overview)

	return overview, nil
}

func (s *Service) generateInsights(overview OverviewDto) []string {
	if s.AIService == nil {
		return []string{}
	}

	summary := buildMetricsSummary(overview)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := s.AIService.Execute(ctx, ai.Request{
		TaskType:     ai.TaskSummarization,
		SystemPrompt: "You are a storage analytics assistant for a personal NAS system. Provide actionable insights based on storage metrics. Respond ONLY with a JSON array of strings, no extra text. Write insights in the user's language (pt-BR).",
		Prompt:       fmt.Sprintf("Analyze these NAS storage metrics and provide 3-5 actionable insights:\n\n%s\n\nRespond with JSON: [\"insight 1\", \"insight 2\", ...]", summary),
		MaxTokens:    500,
		Temperature:  0.3,
	})
	if err != nil {
		log.Printf("AI insights generation failed: %v\n", err)
		return []string{}
	}

	return parseInsightsResponse(resp.Content)
}

func buildMetricsSummary(overview OverviewDto) string {
	var parts []string

	if overview.Storage.TotalBytes > 0 {
		usagePct := float64(overview.Storage.UsedBytes) / float64(overview.Storage.TotalBytes) * 100
		parts = append(parts, fmt.Sprintf("Storage: %.1f%% used (%d bytes of %d bytes)", usagePct, overview.Storage.UsedBytes, overview.Storage.TotalBytes))
	}
	parts = append(parts, fmt.Sprintf("Growth: %d bytes in period %s", overview.Storage.GrowthBytes, overview.Period))
	parts = append(parts, fmt.Sprintf("Files: %d total, %d added in period", overview.Counts.FilesTotal, overview.Counts.FilesAdded))
	parts = append(parts, fmt.Sprintf("Folders: %d", overview.Counts.Folders))
	parts = append(parts, fmt.Sprintf("Duplicates: %d groups, %d bytes reclaimable", overview.Duplicates.Groups, overview.Duplicates.ReclaimableSize))
	parts = append(parts, fmt.Sprintf("Health: %s, errors 24h: %d", overview.Health.Status, overview.Health.ErrorsLast24h))

	if len(overview.HotFolders) > 0 {
		hotNames := make([]string, 0, len(overview.HotFolders))
		for _, hf := range overview.HotFolders {
			hotNames = append(hotNames, fmt.Sprintf("%s (%d new files)", hf.Path, hf.NewFiles))
		}
		parts = append(parts, fmt.Sprintf("Hot folders: %s", strings.Join(hotNames, ", ")))
	}

	parts = append(parts, fmt.Sprintf("Processing: %d metadata pending, %d failed", overview.Processing.MetadataPending, overview.Processing.MetadataFailed))

	return strings.Join(parts, "\n")
}

func parseInsightsResponse(content string) []string {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		filtered := make([]string, 0, len(lines))
		for _, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "```") {
				filtered = append(filtered, line)
			}
		}
		content = strings.Join(filtered, "\n")
	}

	var insights []string
	if err := json.Unmarshal([]byte(content), &insights); err != nil {
		log.Printf("AI insights parse error: %v\n", err)
		return []string{}
	}

	return insights
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
