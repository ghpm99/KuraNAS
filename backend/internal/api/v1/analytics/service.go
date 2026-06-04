package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
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

func (s *Service) GetStorage(period string) (StorageStatsDto, error) {
	periodConfig, err := resolvePeriod(period)
	if err != nil {
		return StorageStatsDto{}, err
	}

	kpis, err := s.Repository.GetStorageKpis(periodConfig)
	if err != nil {
		return StorageStatsDto{}, err
	}

	totalBytes, freeBytes := getFileSystemStorage(kpis.UsedBytes)

	return StorageStatsDto{
		Storage: StorageDto{
			TotalBytes:  totalBytes,
			UsedBytes:   kpis.UsedBytes,
			FreeBytes:   freeBytes,
			GrowthBytes: kpis.GrowthBytes,
		},
		Counts: CountsDto{
			FilesTotal: kpis.FilesTotal,
			FilesAdded: kpis.FilesAdded,
			Folders:    kpis.FoldersTotal,
		},
	}, nil
}

func (s *Service) GetTimeSeries(period string) ([]TimeSeriesPointDto, error) {
	periodConfig, err := resolvePeriod(period)
	if err != nil {
		return nil, err
	}
	models, err := s.Repository.GetStorageTimeSeries(periodConfig)
	if err != nil {
		return nil, err
	}
	return toTimeSeriesDto(models), nil
}

func (s *Service) GetTypes() ([]TypeBreakdownDto, error) {
	models, err := s.Repository.GetTypeDistribution()
	if err != nil {
		return nil, err
	}
	return toTypeBreakdownDto(models), nil
}

func (s *Service) GetExtensions(limit int) ([]ExtensionDto, error) {
	models, err := s.Repository.GetExtensionDistribution(limit)
	if err != nil {
		return nil, err
	}
	return toExtensionDto(models), nil
}

func (s *Service) GetRecentFiles(limit int) ([]RecentFileDto, error) {
	models, err := s.Repository.GetRecentFiles(limit)
	if err != nil {
		return nil, err
	}
	return toRecentFilesDto(models), nil
}

func (s *Service) GetTopFolders(limit int) ([]FolderUsageDto, error) {
	models, err := s.Repository.GetTopFolders(limit)
	if err != nil {
		return nil, err
	}
	return toFolderUsageDto(models), nil
}

func (s *Service) GetHotFolders(period string, limit int) ([]HotFolderDto, error) {
	periodConfig, err := resolvePeriod(period)
	if err != nil {
		return nil, err
	}
	models, err := s.Repository.GetHotFolders(periodConfig, limit)
	if err != nil {
		return nil, err
	}
	return toHotFolderDto(models), nil
}

func (s *Service) GetDuplicatesSummary() (DuplicatesSummaryDto, error) {
	model, err := s.Repository.GetDuplicatesSummary()
	if err != nil {
		return DuplicatesSummaryDto{}, err
	}
	return DuplicatesSummaryDto{
		Groups:          model.GroupsTotal,
		Files:           model.FilesTotal,
		ReclaimableSize: model.ReclaimableBytes,
	}, nil
}

func (s *Service) GetDuplicateGroups(limit int) ([]DuplicateGroupDto, error) {
	models, err := s.Repository.GetDuplicateGroups(limit)
	if err != nil {
		return nil, err
	}
	return toDuplicateGroupDto(models), nil
}

func (s *Service) GetLibrary() (LibraryDto, error) {
	model, err := s.Repository.GetLibrarySummary()
	if err != nil {
		return LibraryDto{}, err
	}
	return LibraryDto{
		CategorizedMedia:  model.CategorizedMedia,
		AudioWithMetadata: model.AudioWithMetadata,
		VideoWithMetadata: model.VideoWithMetadata,
		ImageWithMetadata: model.ImageWithMetadata,
		ImageClassified:   model.ImageClassified,
	}, nil
}

func (s *Service) GetProcessing() (ProcessingDto, error) {
	model, err := s.Repository.GetProcessingSummary()
	if err != nil {
		return ProcessingDto{}, err
	}
	return ProcessingDto{
		MetadataPending:   model.MetadataPending,
		MetadataFailed:    model.MetadataFailed,
		ThumbnailPending:  model.ThumbnailPending,
		ThumbnailFailed:   model.ThumbnailFailed,
		RecurringTimeouts: model.RecurringTimeouts,
	}, nil
}

func (s *Service) GetHealth() (HealthDto, error) {
	model, err := s.Repository.GetHealth()
	if err != nil {
		return HealthDto{}, err
	}
	return toHealthDto(model), nil
}

func (s *Service) GetInsights(period string) ([]string, error) {
	periodConfig, err := resolvePeriod(period)
	if err != nil {
		return nil, err
	}

	if s.AIService == nil {
		return []string{}, nil
	}

	kpis, err := s.Repository.GetStorageKpis(periodConfig)
	if err != nil {
		return nil, err
	}
	duplicates, err := s.Repository.GetDuplicatesSummary()
	if err != nil {
		return nil, err
	}
	processing, err := s.Repository.GetProcessingSummary()
	if err != nil {
		return nil, err
	}
	health, err := s.Repository.GetHealth()
	if err != nil {
		return nil, err
	}
	hotFolders, err := s.Repository.GetHotFolders(periodConfig, 3)
	if err != nil {
		return nil, err
	}

	totalBytes, _ := getFileSystemStorage(kpis.UsedBytes)
	summary := buildMetricsSummary(periodConfig.Label, totalBytes, kpis, duplicates, processing, health, hotFolders)

	return s.generateInsights(summary), nil
}

func (s *Service) generateInsights(summary string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := s.AIService.Execute(ctx, ai.Request{
		TaskType:     ai.TaskSummarization,
		SystemPrompt: prompts.AnalyticsInsightsSystemPrompt(),
		Prompt:       prompts.AnalyticsInsightsUserPrompt(summary),
		MaxTokens:    500,
		Temperature:  0.3,
	})
	if err != nil {
		log.Printf("AI insights generation failed: %v\n", err)
		return []string{}
	}

	return parseInsightsResponse(resp.Content)
}

func buildMetricsSummary(
	periodLabel string,
	totalBytes int64,
	kpis StorageKpisModel,
	duplicates DuplicatesSummaryModel,
	processing ProcessingSummaryModel,
	health HealthModel,
	hotFolders []FolderHotModel,
) string {
	var parts []string

	if totalBytes > 0 {
		usagePct := float64(kpis.UsedBytes) / float64(totalBytes) * 100
		parts = append(parts, fmt.Sprintf("Storage: %.1f%% used (%d bytes of %d bytes)", usagePct, kpis.UsedBytes, totalBytes))
	}
	parts = append(parts, fmt.Sprintf("Growth: %d bytes in period %s", kpis.GrowthBytes, periodLabel))
	parts = append(parts, fmt.Sprintf("Files: %d total, %d added in period", kpis.FilesTotal, kpis.FilesAdded))
	parts = append(parts, fmt.Sprintf("Folders: %d", kpis.FoldersTotal))
	parts = append(parts, fmt.Sprintf("Duplicates: %d groups, %d bytes reclaimable", duplicates.GroupsTotal, duplicates.ReclaimableBytes))
	parts = append(parts, fmt.Sprintf("Health: %s, errors 24h: %d", resolveHealthStatus(health.Status), health.ErrorsLast24h))

	if len(hotFolders) > 0 {
		hotNames := make([]string, 0, len(hotFolders))
		for _, hf := range hotFolders {
			hotNames = append(hotNames, fmt.Sprintf("%s (%d new files)", hf.ParentPath, hf.NewFiles))
		}
		parts = append(parts, fmt.Sprintf("Hot folders: %s", strings.Join(hotNames, ", ")))
	}

	parts = append(parts, fmt.Sprintf("Processing: %d metadata pending, %d failed", processing.MetadataPending, processing.MetadataFailed))

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

func resolveHealthStatus(status sql.NullString) string {
	if !status.Valid {
		return "ok"
	}
	switch status.String {
	case "PENDING":
		return "scanning"
	case "FAILED":
		return "error"
	default:
		return "ok"
	}
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

func toHealthDto(model HealthModel) HealthDto {
	lastScanAt := ""
	lastScanSeconds := int64(0)
	if model.LastScanStart.Valid {
		lastScanAt = model.LastScanStart.Time.Format(time.RFC3339)
	}
	if model.LastScanStart.Valid && model.LastScanEnd.Valid {
		lastScanSeconds = int64(model.LastScanEnd.Time.Sub(model.LastScanStart.Time).Seconds())
	}

	errorsList := make([]string, 0, len(model.RecentErrors))
	sort.SliceStable(model.RecentErrors, func(i, j int) bool {
		return model.RecentErrors[i].CreatedAt.After(model.RecentErrors[j].CreatedAt)
	})
	for _, item := range model.RecentErrors {
		description := item.Name
		if item.Description.Valid && item.Description.String != "" {
			description = fmt.Sprintf("%s: %s", item.Name, item.Description.String)
		}
		errorsList = append(errorsList, description)
	}

	return HealthDto{
		Status:          resolveHealthStatus(model.Status),
		LastScanAt:      lastScanAt,
		LastScanSeconds: lastScanSeconds,
		IndexedFiles:    model.IndexedFiles,
		ErrorsLast24h:   model.ErrorsLast24h,
		RecentErrors:    errorsList,
	}
}
