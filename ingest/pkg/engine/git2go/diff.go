package git2go

import (
	git "github.com/libgit2/git2go"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/rs/zerolog/log"
)

type diffDelta struct {
	repo *git.Repository
	*git.DiffDelta
}

func newDiffDelta(r *git.Repository, d *git.DiffDelta) *diffDelta {
	return &diffDelta{repo: r, DiffDelta: d}
}

func (d *diffDelta) Action() engine.DeltaType {
	switch d.DiffDelta.Status {
	case git.DeltaUnmodified:
		return engine.DeltaUnmodified
	case git.DeltaAdded:
		return engine.DeltaAdded
	case git.DeltaDeleted:
		return engine.DeltaDeleted
	case git.DeltaModified:
		return engine.DeltaModified
	case git.DeltaRenamed:
		return engine.DeltaRenamed
	case git.DeltaCopied:
		return engine.DeltaCopied
	case git.DeltaIgnored:
		return engine.DeltaIgnored
	case git.DeltaUntracked:
		return engine.DeltaUntracked
	case git.DeltaTypeChange:
		return engine.DeltaTypeChange
	case git.DeltaUnreadable:
		return engine.DeltaUnreadable
	case git.DeltaConflicted:
		return engine.DeltaConflicted
	default:
		log.Error().Stringer("delta_status", d.DiffDelta.Status).Msg("unexpected delta status")
		return engine.DeltaOther
	}
}

func (d *diffDelta) OldContent() ([]byte, error) {
	return d.content(d.DiffDelta.OldFile.Oid)
}

func (d *diffDelta) NewContent() ([]byte, error) {
	return d.content(d.DiffDelta.NewFile.Oid)
}

func (d *diffDelta) content(oid *git.Oid) ([]byte, error) {
	if oid == nil || oid.IsZero() {
		return nil, nil
	}

	blob, err := d.repo.LookupBlob(oid)
	if err != nil {
		return nil, err
	}
	defer blob.Free()

	return blob.Contents(), nil
}

func (d *diffDelta) FromPath() string {
	return d.DiffDelta.OldFile.Path
}

func (d *diffDelta) ToPath() string {
	return d.DiffDelta.NewFile.Path
}

func (d *diffDelta) IsBinary() bool {
	return d.DiffDelta.Flags == git.DiffFlagBinary
}

func (d *diffDelta) IsSymlink() bool {
	return git.Filemode(d.DiffDelta.OldFile.Mode) == git.FilemodeLink || git.Filemode(d.DiffDelta.NewFile.Mode) == git.FilemodeLink
}

type diffHunk struct {
	*git.DiffHunk
}

func (h *diffHunk) LinesAdded() int {
	return h.DiffHunk.NewLines
}

func (h *diffHunk) LinesDeleted() int {
	return h.DiffHunk.OldLines
}

func (h *diffHunk) AddressDeleted() engine.ChunkAddr {
	return engine.ChunkAddr{
		Start: h.DiffHunk.OldStart,
		End:   h.DiffHunk.OldStart + h.DiffHunk.OldLines - 1,
	}
}

func newDiffHunk(h *git.DiffHunk) *diffHunk {
	return &diffHunk{DiffHunk: h}
}

type diffLine struct {
	*git.DiffLine
}

func newDiffLine(l *git.DiffLine) *diffLine {
	return &diffLine{DiffLine: l}
}
