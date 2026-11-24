package video

import (
	"fmt"
	"sync"

	"golang.org/x/sync/singleflight"

	files "nas-go/api/internal/api/v1/files"
)

const maxCachedTrackLists = 512

type trackProbeCache[Track any] struct {
	mutex   sync.Mutex
	entries map[embeddedSubtitleVersion][]Track
	flights singleflight.Group
}

func newTrackProbeCache[Track any]() *trackProbeCache[Track] {
	return &trackProbeCache[Track]{entries: map[embeddedSubtitleVersion][]Track{}}
}

func (cache *trackProbeCache[Track]) load(videoFile files.FileDto, probe func() ([]Track, error)) ([]Track, error) {
	version := versionOfVideo(videoFile)
	if tracks, isCached := cache.cached(version); isCached {
		return tracks, nil
	}
	flightKey := fmt.Sprintf("%d:%d:%d", version.fileID, version.updatedAt, version.size)
	probedTracks, err, _ := cache.flights.Do(flightKey, func() (any, error) {
		if tracks, isCached := cache.cached(version); isCached {
			return tracks, nil
		}
		tracks, err := probe()
		if err != nil {
			return nil, err
		}
		cache.store(version, tracks)
		return tracks, nil
	})
	if err != nil {
		return nil, err
	}
	return probedTracks.([]Track), nil
}

func (cache *trackProbeCache[Track]) cached(version embeddedSubtitleVersion) ([]Track, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	tracks, isCached := cache.entries[version]
	return tracks, isCached
}

func (cache *trackProbeCache[Track]) store(version embeddedSubtitleVersion, tracks []Track) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	if len(cache.entries) >= maxCachedTrackLists {
		cache.entries = map[embeddedSubtitleVersion][]Track{}
	}
	cache.entries[version] = tracks
}
