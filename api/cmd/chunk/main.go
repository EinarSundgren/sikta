package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/einarsundgren/sikta/internal/config"
	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		logger.Error("usage: chunk <document-id>")
		os.Exit(1)
	}
	docID := os.Args[1]

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := database.New(pool)

	parsedUUID, err := uuid.Parse(docID)
	if err != nil {
		logger.Error("invalid document ID", "error", err)
		os.Exit(1)
	}

	src, err := queries.GetSource(ctx, database.PgUUID(parsedUUID))
	if err != nil {
		logger.Error("failed to get source", "error", err)
		os.Exit(1)
	}

	content, err := os.ReadFile(src.FilePath)
	if err != nil {
		logger.Error("failed to read file", "error", err)
		os.Exit(1)
	}

	text := string(content)
	chapters := splitByChapter(text)

	logger.Info("found chapters", "count", len(chapters))

	pgSrcID := database.PgUUID(parsedUUID)

	for i, chapter := range chapters {
		title := chapter.Title
		if title == "" {
			title = fmt.Sprintf("Chapter %d", i+1)
		}

		chapterNum := int32(i + 1)
		_, err := queries.CreateChunk(ctx, database.CreateChunkParams{
			SourceID:          pgSrcID,
			ChunkIndex:        int32(i),
			Content:           chapter.Content,
			ChapterTitle:      pgtype.Text{String: title, Valid: true},
			ChapterNumber:     pgtype.Int4{Int32: chapterNum, Valid: true},
			NarrativePosition: int32(i),
		})
		if err != nil {
			logger.Error("failed to create chunk", "index", i, "error", err)
			os.Exit(1)
		}
	}

	logger.Info("chunking complete", "chunks", len(chapters))
}

type Chapter struct {
	Title   string
	Content string
}

func splitByChapter(text string) []Chapter {
	var chapters []Chapter

	lines := strings.Split(text, "\n")
	var currentChapter *Chapter
	var currentContent strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Chapter ") && len(trimmed) < 50 {
			if currentChapter != nil {
				currentChapter.Content = strings.TrimSpace(currentContent.String())
				chapters = append(chapters, *currentChapter)
			}

			currentChapter = &Chapter{
				Title: trimmed,
			}
			currentContent.Reset()
		} else {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	if currentChapter != nil {
		currentChapter.Content = strings.TrimSpace(currentContent.String())
		chapters = append(chapters, *currentChapter)
	}

	return chapters
}
