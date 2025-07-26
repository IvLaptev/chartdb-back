package background

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/storage"
	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	"github.com/IvLaptev/chartdb-back/pkg/s3client"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
)

type CleanObjectStorageJob struct {
	period    time.Duration
	isRunning bool

	logger   *slog.Logger
	s3client s3client.Client
	storage  storage.Storage
}

func (j *CleanObjectStorageJob) Name() string {
	return "clean_object_storage"
}

func (j *CleanObjectStorageJob) Run(now int64) {
	if !(now%int64(j.period.Seconds()) == 0) || j.isRunning {
		return
	}

	j.isRunning = true
	defer func() { j.isRunning = false }()

	ctx := context.Background()
	runID, err := utils.GenerateID(10)
	if err != nil {
		ctxlog.Error(ctx, j.logger, "generate run id", slog.Any("error", err))
		return
	}
	ctx = ctxlog.WithFields(ctx, slog.String("run_id", runID), slog.String("job", j.Name()))

	objectKeys, err := j.fetchObjectStorageKeys(ctx)
	if err != nil {
		ctxlog.Error(ctx, j.logger, "fetch object storage keys", slog.Any("error", err))
		return
	}
	ctxlog.Info(ctx, j.logger, "fetch object storage keys", slog.Int("count", len(objectKeys)))

	diagramKeys, err := j.fetchDiagramObjectStorageKeys(ctx)
	if err != nil {
		ctxlog.Error(ctx, j.logger, "fetch diagram object storage keys", slog.Any("error", err))
		return
	}
	ctxlog.Info(ctx, j.logger, "fetch diagram object storage keys", slog.Int("count", len(diagramKeys)))

	keysToDelete := j.findKeysToDelete(objectKeys, diagramKeys)
	ctxlog.Info(ctx, j.logger, "find keys to delete", slog.Int("count", len(keysToDelete)))

	err = j.deleteUnusedObjects(ctx, keysToDelete)
	if err != nil {
		ctxlog.Error(ctx, j.logger, "delete unused objects", slog.Any("error", err))
		return
	}
	ctxlog.Info(ctx, j.logger, "delete unused objects", slog.Int("count", len(keysToDelete)), slog.Any("keys", keysToDelete))
}

func (j *CleanObjectStorageJob) fetchObjectStorageKeys(ctx context.Context) (map[string]struct{}, error) {
	objectKeys := make(map[string]struct{})

	var nextPageToken *string
	for {
		objectList, err := j.s3client.ListObjects(ctx, nextPageToken)
		if err != nil {
			return nil, fmt.Errorf("list objects: %w", err)
		}

		for _, object := range objectList.Objects {
			if object.LastModified.Before(time.Now().Add(-j.period)) {
				objectKeys[object.Key] = struct{}{}
			}
		}

		nextPageToken = objectList.NextPageToken
		if nextPageToken == nil {
			break
		}
	}

	return objectKeys, nil
}

func (j *CleanObjectStorageJob) fetchDiagramObjectStorageKeys(ctx context.Context) (map[string]struct{}, error) {
	batchSize := 1000
	diagramKeys := make(map[string]struct{})

	var nextPageToken string
	for {
		page, err := model.NewPage[model.OrderByCreatedAt](int64(batchSize), nextPageToken, model.WithDirection(model.OrderByAsc))
		if err != nil {
			return nil, fmt.Errorf("new page: %w", err)
		}

		diagrams, err := j.storage.Diagram().GetAllDiagrams(ctx, &storage.RowPolicyBackground{}, nil, page)
		if err != nil {
			return nil, fmt.Errorf("get all diagrams: %w", err)
		}

		for _, diagram := range diagrams.Diagrams {
			diagramKeys[diagram.ObjectStorageKey] = struct{}{}
		}

		if diagrams.NextPage == nil {
			break
		} else {
			nextPageToken, err = diagrams.NextPage.Token()
			if err != nil {
				return nil, fmt.Errorf("get next page token: %w", err)
			}
		}
	}

	return diagramKeys, nil
}

func (j *CleanObjectStorageJob) findKeysToDelete(objectKeys map[string]struct{}, diagramKeys map[string]struct{}) []string {
	var keysToDelete []string
	for key := range objectKeys {
		if _, ok := diagramKeys[key]; !ok {
			keysToDelete = append(keysToDelete, key)
		}
	}
	return keysToDelete
}

func (j *CleanObjectStorageJob) deleteUnusedObjects(ctx context.Context, keysToDelete []string) error {
	var batchSize = 1000
	var start = 0
	for start < len(keysToDelete) {
		finish := start + batchSize
		if finish > len(keysToDelete) {
			finish = len(keysToDelete)
		}

		batch := keysToDelete[start:finish]
		err := j.s3client.BatchDeleteObjects(ctx, batch)
		if err != nil {
			return fmt.Errorf("batch delete objects: %w", err)
		}
		start += batchSize
	}
	return nil
}
