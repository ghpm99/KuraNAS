package scan

import (
	"nas-go/api/internal/api/v1/files"
)

func CreateFileRecord(service files.ServiceInterface, finalizedFile files.FileDto) (files.FileDto, error) {
	return service.CreateFile(finalizedFile)
}

func UpdateFileRecord(service files.ServiceInterface, finalizedFile files.FileDto, existingRecord files.FileDto) (bool, error) {
	existingRecord.Format = finalizedFile.Format
	existingRecord.Size = finalizedFile.Size
	existingRecord.UpdatedAt = finalizedFile.UpdatedAt
	existingRecord.CreatedAt = finalizedFile.CreatedAt
	existingRecord.DeletedAt = finalizedFile.DeletedAt
	existingRecord.LastInteraction = finalizedFile.LastInteraction
	existingRecord.LastBackup = finalizedFile.LastBackup
	existingRecord.CheckSum = finalizedFile.CheckSum
	existingRecord.DirectoryContentCount = finalizedFile.DirectoryContentCount
	existingRecord.Starred = finalizedFile.Starred
	existingRecord.Metadata = finalizedFile.Metadata

	return service.UpdateFile(existingRecord)
}
