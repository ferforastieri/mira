package playback

import (
	"context"

	"github.com/ferforastieri/mira/internal/library"
	"github.com/ferforastieri/mira/internal/media"
)

// AnalyzeAudio is Analyze for a track or an audiobook: its stream is an
// audio file, which Analyze refuses as it holds no video. Only versions of
// audio items are analyzed so, so that an audio-only clip standing in for
// a movie or an episode is still refused.
func (s *Service) AnalyzeAudio(ctx context.Context, version library.Version) (media.Analysis, error) {
	return s.analyze(ctx, version, media.Prober.ProbeAudio)
}
