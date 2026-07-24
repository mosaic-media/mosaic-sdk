package host

import (
	"time"

	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// Conversions between the SDK's Go types and the wire.
//
// This file is the price of ADR 0064's central decision: the Go interfaces stay
// the contract and the proto is an implementation of them, so the two are
// separate sources of truth that must agree. Nothing enforces the agreement but
// this file and the round-trip tests beside it.
//
// Two conventions throughout, both chosen so a missing case fails loudly rather
// than silently:
//
//   - Enums map through explicit switches with a default, never through
//     integer casts. A cast would turn an unrecognised value into whatever
//     happened to share its number.
//   - Nil messages convert to zero values rather than panicking. proto3 message
//     fields are pointers and an older peer legitimately omits a field this
//     build knows about, which ADR 0064's additive-only rule makes normal.

// ─── Time and duration ──────────────────────────────────────────────────────

// Times cross as RFC 3339 in UTC. A zero time crosses as the empty string
// rather than as year 1, so the far side gets a zero time back rather than a
// date nobody meant.
func timeToWire(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func timeFromWire(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		// A malformed timestamp is a peer bug, and a zero time is the honest
		// representation of "this did not parse". Returning an error here would
		// push a decoding concern into every call site for a case that means
		// the other end is broken.
		return time.Time{}
	}
	return t
}

// Durations cross as nanoseconds, matching time.Duration exactly rather than
// approximating it in seconds — a chapter offset and a bitrate window both care.
func durationToWire(d time.Duration) int64 { return int64(d) }
func durationFromWire(n int64) time.Duration {
	return time.Duration(n)
}

// ─── Caller ─────────────────────────────────────────────────────────────────

// The Caller's Session field carries the invocation handle across the boundary
// (ADR 0064). A module forwards the Caller it was handed and never inspects it,
// so the field's in-process meaning — a session reference — and its
// across-the-boundary meaning — a handle the Platform can revoke — never have to
// be distinguished by module code.
func callerToWire(c v1.Caller) *modulev1.Caller {
	return &modulev1.Caller{Handle: c.Session}
}

func callerFromWire(c *modulev1.Caller) v1.Caller {
	if c == nil {
		return v1.Caller{}
	}
	return v1.Caller{Session: c.GetHandle()}
}

// ─── Closed vocabularies ────────────────────────────────────────────────────

func nodeKindToWire(k v1.NodeKind) modulev1.NodeKind {
	switch k {
	case v1.NodeWork:
		return modulev1.NodeKind_NODE_KIND_WORK
	case v1.NodeContainer:
		return modulev1.NodeKind_NODE_KIND_CONTAINER
	case v1.NodeItem:
		return modulev1.NodeKind_NODE_KIND_ITEM
	default:
		return modulev1.NodeKind_NODE_KIND_UNSPECIFIED
	}
}

func nodeKindFromWire(k modulev1.NodeKind) v1.NodeKind {
	switch k {
	case modulev1.NodeKind_NODE_KIND_WORK:
		return v1.NodeWork
	case modulev1.NodeKind_NODE_KIND_CONTAINER:
		return v1.NodeContainer
	case modulev1.NodeKind_NODE_KIND_ITEM:
		return v1.NodeItem
	default:
		return ""
	}
}

func nodeStatusToWire(s v1.NodeStatus) modulev1.NodeStatus {
	switch s {
	case v1.NodeActive:
		return modulev1.NodeStatus_NODE_STATUS_ACTIVE
	case v1.NodeOrphaned:
		return modulev1.NodeStatus_NODE_STATUS_ORPHANED
	default:
		return modulev1.NodeStatus_NODE_STATUS_UNSPECIFIED
	}
}

func nodeStatusFromWire(s modulev1.NodeStatus) v1.NodeStatus {
	switch s {
	case modulev1.NodeStatus_NODE_STATUS_ACTIVE:
		return v1.NodeActive
	case modulev1.NodeStatus_NODE_STATUS_ORPHANED:
		return v1.NodeOrphaned
	default:
		return ""
	}
}

func partRoleToWire(r v1.PartRole) modulev1.PartRole {
	switch r {
	case v1.PartEdition:
		return modulev1.PartRole_PART_ROLE_EDITION
	case v1.PartSegment:
		return modulev1.PartRole_PART_ROLE_SEGMENT
	default:
		return modulev1.PartRole_PART_ROLE_UNSPECIFIED
	}
}

func partRoleFromWire(r modulev1.PartRole) v1.PartRole {
	switch r {
	case modulev1.PartRole_PART_ROLE_EDITION:
		return v1.PartEdition
	case modulev1.PartRole_PART_ROLE_SEGMENT:
		return v1.PartSegment
	default:
		return ""
	}
}

func schemeToWire(s v1.LocationScheme) modulev1.LocationScheme {
	switch s {
	case v1.LocalLocation:
		return modulev1.LocationScheme_LOCATION_SCHEME_LOCAL
	case v1.RemoteLocation:
		return modulev1.LocationScheme_LOCATION_SCHEME_REMOTE
	default:
		return modulev1.LocationScheme_LOCATION_SCHEME_UNSPECIFIED
	}
}

func schemeFromWire(s modulev1.LocationScheme) v1.LocationScheme {
	switch s {
	case modulev1.LocationScheme_LOCATION_SCHEME_LOCAL:
		return v1.LocalLocation
	case modulev1.LocationScheme_LOCATION_SCHEME_REMOTE:
		return v1.RemoteLocation
	default:
		return ""
	}
}

func relationOriginToWire(o v1.RelationOrigin) modulev1.RelationOrigin {
	switch o {
	case v1.OriginSystemInferred:
		return modulev1.RelationOrigin_RELATION_ORIGIN_SYSTEM_INFERRED
	case v1.OriginProviderSupplied:
		return modulev1.RelationOrigin_RELATION_ORIGIN_PROVIDER_SUPPLIED
	case v1.OriginUserConfirmed:
		return modulev1.RelationOrigin_RELATION_ORIGIN_USER_CONFIRMED
	default:
		return modulev1.RelationOrigin_RELATION_ORIGIN_UNSPECIFIED
	}
}

func relationOriginFromWire(o modulev1.RelationOrigin) v1.RelationOrigin {
	switch o {
	case modulev1.RelationOrigin_RELATION_ORIGIN_SYSTEM_INFERRED:
		return v1.OriginSystemInferred
	case modulev1.RelationOrigin_RELATION_ORIGIN_PROVIDER_SUPPLIED:
		return v1.OriginProviderSupplied
	case modulev1.RelationOrigin_RELATION_ORIGIN_USER_CONFIRMED:
		return v1.OriginUserConfirmed
	default:
		return ""
	}
}

func matchMethodToWire(m v1.MatchMethod) modulev1.MatchMethod {
	switch m {
	case v1.MatchExternalIDExact:
		return modulev1.MatchMethod_MATCH_METHOD_EXTERNAL_ID_EXACT
	case v1.MatchFingerprint:
		return modulev1.MatchMethod_MATCH_METHOD_FINGERPRINT
	case v1.MatchFuzzyTitle:
		return modulev1.MatchMethod_MATCH_METHOD_FUZZY_TITLE
	case v1.MatchUserSelected:
		return modulev1.MatchMethod_MATCH_METHOD_USER_SELECTED
	default:
		return modulev1.MatchMethod_MATCH_METHOD_UNSPECIFIED
	}
}

func matchMethodFromWire(m modulev1.MatchMethod) v1.MatchMethod {
	switch m {
	case modulev1.MatchMethod_MATCH_METHOD_EXTERNAL_ID_EXACT:
		return v1.MatchExternalIDExact
	case modulev1.MatchMethod_MATCH_METHOD_FINGERPRINT:
		return v1.MatchFingerprint
	case modulev1.MatchMethod_MATCH_METHOD_FUZZY_TITLE:
		return v1.MatchFuzzyTitle
	case modulev1.MatchMethod_MATCH_METHOD_USER_SELECTED:
		return v1.MatchUserSelected
	default:
		return ""
	}
}

func bindingStatusToWire(s v1.BindingStatus) modulev1.BindingStatus {
	switch s {
	case v1.BindingConfirmed:
		return modulev1.BindingStatus_BINDING_STATUS_CONFIRMED
	case v1.BindingPendingReview:
		return modulev1.BindingStatus_BINDING_STATUS_PENDING_REVIEW
	case v1.BindingRejected:
		return modulev1.BindingStatus_BINDING_STATUS_REJECTED
	default:
		return modulev1.BindingStatus_BINDING_STATUS_UNSPECIFIED
	}
}

func bindingStatusFromWire(s modulev1.BindingStatus) v1.BindingStatus {
	switch s {
	case modulev1.BindingStatus_BINDING_STATUS_CONFIRMED:
		return v1.BindingConfirmed
	case modulev1.BindingStatus_BINDING_STATUS_PENDING_REVIEW:
		return v1.BindingPendingReview
	case modulev1.BindingStatus_BINDING_STATUS_REJECTED:
		return v1.BindingRejected
	default:
		return ""
	}
}

// BindingResolution has two values and not three: a split is Confirm with
// MoveToNodeID set, which is why there is no MOVE case to map here.
func bindingResolutionToWire(r v1.BindingResolution) modulev1.BindingResolution {
	switch r {
	case v1.ResolveConfirm:
		return modulev1.BindingResolution_BINDING_RESOLUTION_CONFIRM
	case v1.ResolveReject:
		return modulev1.BindingResolution_BINDING_RESOLUTION_REJECT
	default:
		return modulev1.BindingResolution_BINDING_RESOLUTION_UNSPECIFIED
	}
}

func bindingResolutionFromWire(r modulev1.BindingResolution) v1.BindingResolution {
	switch r {
	case modulev1.BindingResolution_BINDING_RESOLUTION_CONFIRM:
		return v1.ResolveConfirm
	case modulev1.BindingResolution_BINDING_RESOLUTION_REJECT:
		return v1.ResolveReject
	default:
		return ""
	}
}

func playbackKindToWire(k v1.PlaybackKind) modulev1.PlaybackKind {
	switch k {
	case v1.PlaybackDirect:
		return modulev1.PlaybackKind_PLAYBACK_KIND_DIRECT
	default:
		return modulev1.PlaybackKind_PLAYBACK_KIND_UNSPECIFIED
	}
}

func playbackKindFromWire(k modulev1.PlaybackKind) v1.PlaybackKind {
	switch k {
	case modulev1.PlaybackKind_PLAYBACK_KIND_DIRECT:
		return v1.PlaybackDirect
	default:
		return ""
	}
}

func watchOfferTypeToWire(t v1.WatchOfferType) modulev1.WatchOfferType {
	switch t {
	case v1.WatchSubscription:
		return modulev1.WatchOfferType_WATCH_OFFER_TYPE_SUBSCRIPTION
	case v1.WatchRent:
		return modulev1.WatchOfferType_WATCH_OFFER_TYPE_RENT
	case v1.WatchBuy:
		return modulev1.WatchOfferType_WATCH_OFFER_TYPE_BUY
	case v1.WatchFree:
		return modulev1.WatchOfferType_WATCH_OFFER_TYPE_FREE
	case v1.WatchAds:
		return modulev1.WatchOfferType_WATCH_OFFER_TYPE_ADS
	default:
		return modulev1.WatchOfferType_WATCH_OFFER_TYPE_UNSPECIFIED
	}
}

func watchOfferTypeFromWire(t modulev1.WatchOfferType) v1.WatchOfferType {
	switch t {
	case modulev1.WatchOfferType_WATCH_OFFER_TYPE_SUBSCRIPTION:
		return v1.WatchSubscription
	case modulev1.WatchOfferType_WATCH_OFFER_TYPE_RENT:
		return v1.WatchRent
	case modulev1.WatchOfferType_WATCH_OFFER_TYPE_BUY:
		return v1.WatchBuy
	case modulev1.WatchOfferType_WATCH_OFFER_TYPE_FREE:
		return v1.WatchFree
	case modulev1.WatchOfferType_WATCH_OFFER_TYPE_ADS:
		return v1.WatchAds
	default:
		return ""
	}
}

func redactionToWire(r v1.RedactionClass) modulev1.RedactionClass {
	switch r {
	case v1.RedactionNone:
		return modulev1.RedactionClass_REDACTION_CLASS_NONE
	case v1.RedactionSensitive:
		return modulev1.RedactionClass_REDACTION_CLASS_SENSITIVE
	case v1.RedactionSecret:
		return modulev1.RedactionClass_REDACTION_CLASS_SECRET
	case v1.RedactionIdentifier:
		return modulev1.RedactionClass_REDACTION_CLASS_IDENTIFIER
	default:
		return modulev1.RedactionClass_REDACTION_CLASS_UNSPECIFIED
	}
}

func redactionFromWire(r modulev1.RedactionClass) v1.RedactionClass {
	switch r {
	case modulev1.RedactionClass_REDACTION_CLASS_NONE:
		return v1.RedactionNone
	case modulev1.RedactionClass_REDACTION_CLASS_SENSITIVE:
		return v1.RedactionSensitive
	case modulev1.RedactionClass_REDACTION_CLASS_SECRET:
		return v1.RedactionSecret
	case modulev1.RedactionClass_REDACTION_CLASS_IDENTIFIER:
		return v1.RedactionIdentifier
	default:
		return ""
	}
}

// ─── Content types ──────────────────────────────────────────────────────────

func locationToWire(l v1.MediaLocation) *modulev1.MediaLocation {
	return &modulev1.MediaLocation{
		Scheme:   schemeToWire(l.Scheme),
		Provider: l.Provider,
		Ref:      l.Ref,
	}
}

func locationFromWire(l *modulev1.MediaLocation) v1.MediaLocation {
	if l == nil {
		return v1.MediaLocation{}
	}
	return v1.MediaLocation{
		Scheme:   schemeFromWire(l.GetScheme()),
		Provider: l.GetProvider(),
		Ref:      l.GetRef(),
	}
}

func artworkCandidateToWire(c v1.ArtworkCandidate) *modulev1.ArtworkCandidate {
	return &modulev1.ArtworkCandidate{
		Slot:     string(c.Slot),
		Url:      c.URL,
		Source:   c.Source,
		Language: c.Language,
		Rank:     c.Rank,
	}
}

func artworkCandidateFromWire(c *modulev1.ArtworkCandidate) v1.ArtworkCandidate {
	if c == nil {
		return v1.ArtworkCandidate{}
	}
	return v1.ArtworkCandidate{
		Slot:     v1.ArtworkSlot(c.GetSlot()),
		URL:      c.GetUrl(),
		Source:   c.GetSource(),
		Language: c.GetLanguage(),
		Rank:     c.GetRank(),
	}
}

func artworkToWire(a v1.Artwork) *modulev1.Artwork {
	out := &modulev1.Artwork{
		Poster:    a.Poster,
		Landscape: a.Landscape,
		Backdrop:  a.Backdrop,
		Logo:      a.Logo,
	}
	for _, c := range a.Candidates {
		out.Candidates = append(out.Candidates, artworkCandidateToWire(c))
	}
	return out
}

func artworkFromWire(a *modulev1.Artwork) v1.Artwork {
	if a == nil {
		return v1.Artwork{}
	}
	out := v1.Artwork{
		Poster:    a.GetPoster(),
		Landscape: a.GetLandscape(),
		Backdrop:  a.GetBackdrop(),
		Logo:      a.GetLogo(),
	}
	for _, c := range a.GetCandidates() {
		out.Candidates = append(out.Candidates, artworkCandidateFromWire(c))
	}
	return out
}

func nodeToWire(n v1.Node) *modulev1.Node {
	out := &modulev1.Node{
		Id:            string(n.ID),
		WorkId:        string(n.WorkID),
		Kind:          nodeKindToWire(n.Kind),
		MediaType:     string(n.MediaType),
		ContainerType: string(n.ContainerType),
		ItemType:      string(n.ItemType),
		Title:         n.Title,
		NaturalOrder:  n.NaturalOrder,
		Status:        nodeStatusToWire(n.Status),
		ExternalIds:   n.ExternalIDs,
		Attributes:    n.Attributes,
		Artwork:       artworkToWire(n.Artwork),
		CreatedAt:     timeToWire(n.CreatedAt),
		UpdatedAt:     timeToWire(n.UpdatedAt),
	}
	// A Work has no parent, and proto3 has no nullable scalar. Empty is
	// unambiguous because a node id is never the empty string.
	if n.ParentID != nil {
		out.ParentId = string(*n.ParentID)
	}
	return out
}

func nodeFromWire(n *modulev1.Node) v1.Node {
	if n == nil {
		return v1.Node{}
	}
	out := v1.Node{
		ID:            v1.NodeID(n.GetId()),
		WorkID:        v1.NodeID(n.GetWorkId()),
		Kind:          nodeKindFromWire(n.GetKind()),
		MediaType:     v1.MediaType(n.GetMediaType()),
		ContainerType: v1.ContainerType(n.GetContainerType()),
		ItemType:      v1.ItemType(n.GetItemType()),
		Title:         n.GetTitle(),
		NaturalOrder:  n.GetNaturalOrder(),
		Status:        nodeStatusFromWire(n.GetStatus()),
		ExternalIDs:   n.GetExternalIds(),
		Attributes:    n.GetAttributes(),
		Artwork:       artworkFromWire(n.GetArtwork()),
		CreatedAt:     timeFromWire(n.GetCreatedAt()),
		UpdatedAt:     timeFromWire(n.GetUpdatedAt()),
	}
	if id := n.GetParentId(); id != "" {
		pid := v1.NodeID(id)
		out.ParentID = &pid
	}
	return out
}

func nodesToWire(ns []v1.Node) []*modulev1.Node {
	out := make([]*modulev1.Node, 0, len(ns))
	for _, n := range ns {
		out = append(out, nodeToWire(n))
	}
	return out
}

func nodesFromWire(ns []*modulev1.Node) []v1.Node {
	out := make([]v1.Node, 0, len(ns))
	for _, n := range ns {
		out = append(out, nodeFromWire(n))
	}
	return out
}

func partToWire(p v1.Part) *modulev1.Part {
	return &modulev1.Part{
		Id:            string(p.ID),
		NodeId:        string(p.NodeID),
		Role:          partRoleToWire(p.Role),
		EditionLabel:  p.EditionLabel,
		NaturalOrder:  p.NaturalOrder,
		Location:      locationToWire(p.Location),
		Container:     p.Container,
		VideoCodec:    p.VideoCodec,
		AudioCodec:    p.AudioCodec,
		Width:         int32(p.Width),
		Height:        int32(p.Height),
		HdrFormat:     p.HDRFormat,
		DurationNanos: durationToWire(p.Duration),
		BitrateBps:    p.BitrateBPS,
		SizeBytes:     p.SizeBytes,
		Attributes:    p.Attributes,
		CreatedAt:     timeToWire(p.CreatedAt),
		UpdatedAt:     timeToWire(p.UpdatedAt),
	}
}

func partFromWire(p *modulev1.Part) v1.Part {
	if p == nil {
		return v1.Part{}
	}
	return v1.Part{
		ID:           v1.PartID(p.GetId()),
		NodeID:       v1.NodeID(p.GetNodeId()),
		Role:         partRoleFromWire(p.GetRole()),
		EditionLabel: p.GetEditionLabel(),
		NaturalOrder: p.GetNaturalOrder(),
		Location:     locationFromWire(p.GetLocation()),
		Container:    p.GetContainer(),
		VideoCodec:   p.GetVideoCodec(),
		AudioCodec:   p.GetAudioCodec(),
		Width:        int(p.GetWidth()),
		Height:       int(p.GetHeight()),
		HDRFormat:    p.GetHdrFormat(),
		Duration:     durationFromWire(p.GetDurationNanos()),
		BitrateBPS:   p.GetBitrateBps(),
		SizeBytes:    p.GetSizeBytes(),
		Attributes:   p.GetAttributes(),
		CreatedAt:    timeFromWire(p.GetCreatedAt()),
		UpdatedAt:    timeFromWire(p.GetUpdatedAt()),
	}
}

func relationToWire(r v1.Relation) *modulev1.Relation {
	return &modulev1.Relation{
		Id:         string(r.ID),
		FromNodeId: string(r.FromNodeID),
		ToNodeId:   string(r.ToNodeID),
		Type:       string(r.Type),
		Confidence: r.Confidence,
		Origin:     relationOriginToWire(r.Origin),
		CreatedAt:  timeToWire(r.CreatedAt),
	}
}

func relationFromWire(r *modulev1.Relation) v1.Relation {
	if r == nil {
		return v1.Relation{}
	}
	return v1.Relation{
		ID:         v1.RelationID(r.GetId()),
		FromNodeID: v1.NodeID(r.GetFromNodeId()),
		ToNodeID:   v1.NodeID(r.GetToNodeId()),
		Type:       v1.RelationType(r.GetType()),
		Confidence: r.GetConfidence(),
		Origin:     relationOriginFromWire(r.GetOrigin()),
		CreatedAt:  timeFromWire(r.GetCreatedAt()),
	}
}

func bindingToWire(b v1.SourceBinding) *modulev1.SourceBinding {
	return &modulev1.SourceBinding{
		Id:              string(b.ID),
		NodeId:          string(b.NodeID),
		SourceProvider:  b.SourceProvider,
		SourceRef:       b.SourceRef,
		MatchConfidence: b.MatchConfidence,
		MatchMethod:     matchMethodToWire(b.MatchMethod),
		Status:          bindingStatusToWire(b.Status),
		CreatedAt:       timeToWire(b.CreatedAt),
		UpdatedAt:       timeToWire(b.UpdatedAt),
	}
}

func bindingFromWire(b *modulev1.SourceBinding) v1.SourceBinding {
	if b == nil {
		return v1.SourceBinding{}
	}
	return v1.SourceBinding{
		ID:              v1.SourceBindingID(b.GetId()),
		NodeID:          v1.NodeID(b.GetNodeId()),
		SourceProvider:  b.GetSourceProvider(),
		SourceRef:       b.GetSourceRef(),
		MatchConfidence: b.GetMatchConfidence(),
		MatchMethod:     matchMethodFromWire(b.GetMatchMethod()),
		Status:          bindingStatusFromWire(b.GetStatus()),
		CreatedAt:       timeFromWire(b.GetCreatedAt()),
		UpdatedAt:       timeFromWire(b.GetUpdatedAt()),
	}
}

func refToWire(r v1.ContentRef) *modulev1.ContentRef {
	return &modulev1.ContentRef{
		Provider:       r.Provider,
		NativeId:       r.NativeID,
		NativeType:     r.NativeType,
		MediaType:      string(r.MediaType),
		ExternalScheme: r.ExternalScheme,
		ExternalId:     r.ExternalID,
	}
}

func refFromWire(r *modulev1.ContentRef) v1.ContentRef {
	if r == nil {
		return v1.ContentRef{}
	}
	return v1.ContentRef{
		Provider:       r.GetProvider(),
		NativeID:       r.GetNativeId(),
		NativeType:     r.GetNativeType(),
		MediaType:      v1.MediaType(r.GetMediaType()),
		ExternalScheme: r.GetExternalScheme(),
		ExternalID:     r.GetExternalId(),
	}
}

func playbackStateToWire(s v1.PlaybackState) *modulev1.PlaybackState {
	return &modulev1.PlaybackState{
		NodeId:           string(s.NodeID),
		PartId:           string(s.PartID),
		PositionNanos:    durationToWire(s.Position),
		DurationNanos:    durationToWire(s.Duration),
		Finished:         s.Finished,
		FinishedExplicit: s.FinishedExplicit,
		UpdatedAt:        timeToWire(s.UpdatedAt),
	}
}

func playbackStateFromWire(s *modulev1.PlaybackState) v1.PlaybackState {
	if s == nil {
		return v1.PlaybackState{}
	}
	return v1.PlaybackState{
		NodeID:           v1.NodeID(s.GetNodeId()),
		PartID:           v1.PartID(s.GetPartId()),
		Position:         durationFromWire(s.GetPositionNanos()),
		Duration:         durationFromWire(s.GetDurationNanos()),
		Finished:         s.GetFinished(),
		FinishedExplicit: s.GetFinishedExplicit(),
		UpdatedAt:        timeFromWire(s.GetUpdatedAt()),
	}
}
