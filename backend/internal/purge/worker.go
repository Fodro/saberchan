// Package purge implements the same-process background worker that
// permanently removes soft-deleted boards/threads/posts (and their S3
// attachments) once their 24h restore grace window has elapsed.
package purge

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Fodro/saberchan/internal/database"
	"github.com/Fodro/saberchan/internal/file"
)

// gracePeriod mirrors the 24h restore window enforced by board.Service.
const gracePeriod = 24 * time.Hour

// staleBumpAge is how long a thread may sit without a bump before the
// worker soft-deletes it (cascading to posts). Media is removed by the
// normal post purge pass after gracePeriod.
const staleBumpAge = 30 * 24 * time.Hour

// Run performs an immediate sweep, then repeats every interval until ctx is
// cancelled. It is intended to be launched with `go purge.Run(...)` and
// stopped by cancelling ctx during shutdown.
func Run(ctx context.Context, repo database.Repository, files file.Service, interval time.Duration) {
	Sweep(ctx, repo, files)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			Sweep(ctx, repo, files)
		}
	}
}

// Sweep soft-deletes threads that have not been bumped for staleBumpAge,
// then purges everything whose soft-delete grace period has elapsed:
// posts have their S3 attachments and attachment rows removed before being
// marked purged; threads/boards are simply marked purged (their own posts
// are purged independently by the post pass above).
func Sweep(ctx context.Context, repo database.Repository, files file.Service) {
	log.Print("purge: sweep started")

	var postsPurged, threadsPurged, boardsPurged, staleSoftDeleted, errCount int

	staleCutoff := time.Now().Add(-staleBumpAge)
	stale, err := repo.ListStaleThreads(ctx, staleCutoff)
	if err != nil {
		log.Printf("purge: failed to list stale threads: %v", err)
		errCount++
	} else {
		for _, thread := range stale {
			if err := repo.SoftDeleteThread(ctx, thread.ID); err != nil {
				log.Printf("purge: failed to soft-delete stale thread %s: %v", thread.ID, err)
				errCount++
				continue
			}
			staleSoftDeleted++
		}
	}

	cutoff := time.Now().Add(-gracePeriod)

	posts, err := repo.ListPostsDueForPurge(ctx, cutoff)
	if err != nil {
		log.Printf("purge: failed to list posts due for purge: %v", err)
		errCount++
	}
	for _, post := range posts {
		if err := purgePost(ctx, repo, files, post); err != nil {
			log.Printf("purge: failed to purge post %s: %v", post.ID, err)
			errCount++
			continue
		}
		postsPurged++
	}

	threads, err := repo.ListThreadsDueForPurge(ctx, cutoff)
	if err != nil {
		log.Printf("purge: failed to list threads due for purge: %v", err)
		errCount++
	}
	for _, thread := range threads {
		if err := repo.MarkThreadPurged(ctx, thread.ID); err != nil {
			log.Printf("purge: failed to mark thread %s purged: %v", thread.ID, err)
			errCount++
			continue
		}
		threadsPurged++
	}

	boards, err := repo.ListBoardsDueForPurge(ctx, cutoff)
	if err != nil {
		log.Printf("purge: failed to list boards due for purge: %v", err)
		errCount++
	}
	for _, b := range boards {
		if err := repo.MarkBoardPurged(ctx, b.ID); err != nil {
			log.Printf("purge: failed to mark board %s purged: %v", b.ID, err)
			errCount++
			continue
		}
		boardsPurged++
	}

	log.Printf(
		"purge: sweep completed stale_soft_deleted=%d posts=%d threads=%d boards=%d errors=%d",
		staleSoftDeleted, postsPurged, threadsPurged, boardsPurged, errCount,
	)
}

// purgePost deletes every S3 object attached to post, then the attachment
// rows, then marks the post purged. It stops (and returns an error) on the
// first failure so a retried sweep can pick up where it left off.
func purgePost(ctx context.Context, repo database.Repository, files file.Service, post database.Post) error {
	attachments, err := repo.GetAttachments(ctx, post.ID)
	if err != nil {
		return fmt.Errorf("get attachments: %w", err)
	}

	for _, a := range attachments {
		key := a.Key
		if key == "" {
			key = KeyFromLink(a.Link)
		}
		if key == "" {
			continue
		}
		if err := files.DeleteFile(ctx, key); err != nil {
			return fmt.Errorf("delete file %s: %w", key, err)
		}
		log.Printf("purge: deleted %s", key)
	}

	if err := repo.DeleteAttachmentsByPostID(ctx, post.ID); err != nil {
		return fmt.Errorf("delete attachment rows: %w", err)
	}
	if err := repo.MarkPostPurged(ctx, post.ID); err != nil {
		return fmt.Errorf("mark post purged: %w", err)
	}
	return nil
}

// KeyFromLink recovers the S3 object key from a public link when the
// attachment's key column is empty (rows written before the key column
// existed): the key is always the last path segment.
func KeyFromLink(link string) string {
	link = strings.TrimSuffix(strings.TrimSpace(link), "/")
	if link == "" {
		return ""
	}
	if idx := strings.LastIndex(link, "/"); idx != -1 {
		return link[idx+1:]
	}
	return link
}
