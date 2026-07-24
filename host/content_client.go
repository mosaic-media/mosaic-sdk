package host

import (
	"context"

	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// contentClient runs in the module process and implements [v1.ContentService]
// by calling back to the Platform. A module's own code holds it as the
// interface and cannot tell it from the in-process application service.
//
// This is the chatty direction ADR 0064 leaves open: a tree import makes one
// call per node, each now a round trip over the socket. That record says the
// cost must be measured against a real import before the protocol is fixed, and
// allows the service shape to grow coarser, batched verbs if it warrants them.
// Those would be additions to ContentService rather than changes to what is
// here.
type contentClient struct {
	client modulev1.ContentServiceClient
}

var _ v1.ContentService = (*contentClient)(nil)

func (c *contentClient) AddContentWork(ctx context.Context, cmd v1.AddContentWorkCommand) (v1.AddContentWorkResult, error) {
	resp, err := c.client.AddContentWork(ctx, &modulev1.AddContentWorkRequest{
		Caller:      callerToWire(cmd.Caller),
		MediaType:   string(cmd.MediaType),
		Title:       cmd.Title,
		ExternalIds: cmd.ExternalIDs,
		Attributes:  cmd.Attributes,
		Artwork:     artworkToWire(cmd.Artwork),
	})
	if err != nil {
		return v1.AddContentWorkResult{}, errorFromWire(err)
	}
	return v1.AddContentWorkResult{Work: nodeFromWire(resp.GetWork())}, nil
}

func (c *contentClient) AddContentChild(ctx context.Context, cmd v1.AddContentChildCommand) (v1.AddContentChildResult, error) {
	resp, err := c.client.AddContentChild(ctx, &modulev1.AddContentChildRequest{
		Caller:        callerToWire(cmd.Caller),
		ParentId:      string(cmd.ParentID),
		Kind:          nodeKindToWire(cmd.Kind),
		ContainerType: string(cmd.ContainerType),
		ItemType:      string(cmd.ItemType),
		Title:         cmd.Title,
		NaturalOrder:  cmd.NaturalOrder,
		ExternalIds:   cmd.ExternalIDs,
		Attributes:    cmd.Attributes,
		Artwork:       artworkToWire(cmd.Artwork),
	})
	if err != nil {
		return v1.AddContentChildResult{}, errorFromWire(err)
	}
	return v1.AddContentChildResult{Node: nodeFromWire(resp.GetNode())}, nil
}

func (c *contentClient) AttachContentPart(ctx context.Context, cmd v1.AttachContentPartCommand) (v1.AttachContentPartResult, error) {
	resp, err := c.client.AttachContentPart(ctx, &modulev1.AttachContentPartRequest{
		Caller:        callerToWire(cmd.Caller),
		NodeId:        string(cmd.NodeID),
		Role:          partRoleToWire(cmd.Role),
		EditionLabel:  cmd.EditionLabel,
		NaturalOrder:  cmd.NaturalOrder,
		Location:      locationToWire(cmd.Location),
		Container:     cmd.Container,
		VideoCodec:    cmd.VideoCodec,
		AudioCodec:    cmd.AudioCodec,
		Width:         int32(cmd.Width),
		Height:        int32(cmd.Height),
		HdrFormat:     cmd.HDRFormat,
		DurationNanos: durationToWire(cmd.Duration),
		BitrateBps:    cmd.BitrateBPS,
		SizeBytes:     cmd.SizeBytes,
		Attributes:    cmd.Attributes,
	})
	if err != nil {
		return v1.AttachContentPartResult{}, errorFromWire(err)
	}
	return v1.AttachContentPartResult{Part: partFromWire(resp.GetPart())}, nil
}

func (c *contentClient) SetContentArtwork(ctx context.Context, cmd v1.SetContentArtworkCommand) (v1.SetContentArtworkResult, error) {
	resp, err := c.client.SetContentArtwork(ctx, &modulev1.SetContentArtworkRequest{
		Caller:  callerToWire(cmd.Caller),
		NodeId:  string(cmd.NodeID),
		Artwork: artworkToWire(cmd.Artwork),
	})
	if err != nil {
		return v1.SetContentArtworkResult{}, errorFromWire(err)
	}
	return v1.SetContentArtworkResult{Node: nodeFromWire(resp.GetNode())}, nil
}

func (c *contentClient) RelateContent(ctx context.Context, cmd v1.RelateContentCommand) (v1.RelateContentResult, error) {
	resp, err := c.client.RelateContent(ctx, &modulev1.RelateContentRequest{
		Caller:     callerToWire(cmd.Caller),
		FromNodeId: string(cmd.FromNodeID),
		ToNodeId:   string(cmd.ToNodeID),
		Type:       string(cmd.Type),
		Confidence: cmd.Confidence,
		Origin:     relationOriginToWire(cmd.Origin),
	})
	if err != nil {
		return v1.RelateContentResult{}, errorFromWire(err)
	}
	return v1.RelateContentResult{Relation: relationFromWire(resp.GetRelation())}, nil
}

func (c *contentClient) BindContentSource(ctx context.Context, cmd v1.BindContentSourceCommand) (v1.BindContentSourceResult, error) {
	resp, err := c.client.BindContentSource(ctx, &modulev1.BindContentSourceRequest{
		Caller:          callerToWire(cmd.Caller),
		NodeId:          string(cmd.NodeID),
		SourceProvider:  cmd.SourceProvider,
		SourceRef:       cmd.SourceRef,
		MatchConfidence: cmd.MatchConfidence,
		MatchMethod:     matchMethodToWire(cmd.MatchMethod),
		Status:          bindingStatusToWire(cmd.Status),
	})
	if err != nil {
		return v1.BindContentSourceResult{}, errorFromWire(err)
	}
	return v1.BindContentSourceResult{Binding: bindingFromWire(resp.GetBinding())}, nil
}

func (c *contentClient) ResolveContentBinding(ctx context.Context, cmd v1.ResolveContentBindingCommand) (v1.ResolveContentBindingResult, error) {
	resp, err := c.client.ResolveContentBinding(ctx, &modulev1.ResolveContentBindingRequest{
		Caller:       callerToWire(cmd.Caller),
		BindingId:    string(cmd.BindingID),
		Resolution:   bindingResolutionToWire(cmd.Resolution),
		MoveToNodeId: string(cmd.MoveToNodeID),
	})
	if err != nil {
		return v1.ResolveContentBindingResult{}, errorFromWire(err)
	}
	return v1.ResolveContentBindingResult{Binding: bindingFromWire(resp.GetBinding())}, nil
}

func (c *contentClient) SearchContent(ctx context.Context, q v1.SearchContentQuery) (v1.SearchContentResult, error) {
	resp, err := c.client.SearchContent(ctx, &modulev1.SearchContentRequest{
		Caller:            callerToWire(q.Caller),
		Title:             q.Title,
		MediaType:         string(q.MediaType),
		Kind:              nodeKindToWire(q.Kind),
		AttributesContain: q.AttributesContain,
		Limit:             int32(q.Limit),
	})
	if err != nil {
		return v1.SearchContentResult{}, errorFromWire(err)
	}
	return v1.SearchContentResult{Nodes: nodesFromWire(resp.GetNodes())}, nil
}

func (c *contentClient) FindContentByExternalID(ctx context.Context, q v1.FindContentByExternalIDQuery) (v1.FindContentByExternalIDResult, error) {
	resp, err := c.client.FindContentByExternalID(ctx, &modulev1.FindContentByExternalIDRequest{
		Caller: callerToWire(q.Caller),
		Scheme: q.Scheme,
		Value:  q.Value,
	})
	if err != nil {
		return v1.FindContentByExternalIDResult{}, errorFromWire(err)
	}
	return v1.FindContentByExternalIDResult{Nodes: nodesFromWire(resp.GetNodes())}, nil
}

func (c *contentClient) GetContentNode(ctx context.Context, q v1.GetContentNodeQuery) (v1.GetContentNodeResult, error) {
	resp, err := c.client.GetContentNode(ctx, &modulev1.GetContentNodeRequest{
		Caller:       callerToWire(q.Caller),
		NodeId:       string(q.NodeID),
		WithChildren: q.WithChildren,
	})
	if err != nil {
		return v1.GetContentNodeResult{}, errorFromWire(err)
	}
	return v1.GetContentNodeResult{
		Node:     nodeFromWire(resp.GetNode()),
		Children: nodesFromWire(resp.GetChildren()),
	}, nil
}

func (c *contentClient) ListContentParts(ctx context.Context, q v1.ListContentPartsQuery) (v1.ListContentPartsResult, error) {
	resp, err := c.client.ListContentParts(ctx, &modulev1.ListContentPartsRequest{
		Caller: callerToWire(q.Caller),
		NodeId: string(q.NodeID),
	})
	if err != nil {
		return v1.ListContentPartsResult{}, errorFromWire(err)
	}
	out := v1.ListContentPartsResult{}
	for _, p := range resp.GetParts() {
		out.Parts = append(out.Parts, partFromWire(p))
	}
	return out, nil
}

func (c *contentClient) RecordPlaybackProgress(ctx context.Context, cmd v1.RecordPlaybackProgressCommand) (v1.RecordPlaybackProgressResult, error) {
	resp, err := c.client.RecordPlaybackProgress(ctx, &modulev1.RecordPlaybackProgressRequest{
		Caller:        callerToWire(cmd.Caller),
		NodeId:        string(cmd.NodeID),
		PartId:        string(cmd.PartID),
		PositionNanos: durationToWire(cmd.Position),
		DurationNanos: durationToWire(cmd.Duration),
	})
	if err != nil {
		return v1.RecordPlaybackProgressResult{}, errorFromWire(err)
	}
	return v1.RecordPlaybackProgressResult{State: playbackStateFromWire(resp.GetState())}, nil
}

func (c *contentClient) SetPlaybackFinished(ctx context.Context, cmd v1.SetPlaybackFinishedCommand) (v1.SetPlaybackFinishedResult, error) {
	resp, err := c.client.SetPlaybackFinished(ctx, &modulev1.SetPlaybackFinishedRequest{
		Caller:   callerToWire(cmd.Caller),
		NodeId:   string(cmd.NodeID),
		Finished: cmd.Finished,
	})
	if err != nil {
		return v1.SetPlaybackFinishedResult{}, errorFromWire(err)
	}
	return v1.SetPlaybackFinishedResult{State: playbackStateFromWire(resp.GetState())}, nil
}

func (c *contentClient) GetPlaybackState(ctx context.Context, q v1.GetPlaybackStateQuery) (v1.GetPlaybackStateResult, error) {
	resp, err := c.client.GetPlaybackState(ctx, &modulev1.GetPlaybackStateRequest{
		Caller: callerToWire(q.Caller),
		NodeId: string(q.NodeID),
	})
	if err != nil {
		return v1.GetPlaybackStateResult{}, errorFromWire(err)
	}
	return v1.GetPlaybackStateResult{
		State: playbackStateFromWire(resp.GetState()),
		Found: resp.GetFound(),
	}, nil
}

func (c *contentClient) ListPlaybackStates(ctx context.Context, q v1.ListPlaybackStatesQuery) (v1.ListPlaybackStatesResult, error) {
	ids := make([]string, 0, len(q.NodeIDs))
	for _, id := range q.NodeIDs {
		ids = append(ids, string(id))
	}
	resp, err := c.client.ListPlaybackStates(ctx, &modulev1.ListPlaybackStatesRequest{
		Caller:  callerToWire(q.Caller),
		NodeIds: ids,
	})
	if err != nil {
		return v1.ListPlaybackStatesResult{}, errorFromWire(err)
	}
	out := v1.ListPlaybackStatesResult{States: make(map[v1.NodeID]v1.PlaybackState, len(resp.GetStates()))}
	for id, st := range resp.GetStates() {
		out.States[v1.NodeID(id)] = playbackStateFromWire(st)
	}
	return out, nil
}

func (c *contentClient) ListInProgress(ctx context.Context, q v1.ListInProgressQuery) (v1.ListInProgressResult, error) {
	resp, err := c.client.ListInProgress(ctx, &modulev1.ListInProgressRequest{
		Caller: callerToWire(q.Caller),
		Limit:  int32(q.Limit),
	})
	if err != nil {
		return v1.ListInProgressResult{}, errorFromWire(err)
	}
	out := v1.ListInProgressResult{}
	for _, i := range resp.GetItems() {
		out.Items = append(out.Items, v1.InProgressItem{
			Node:  nodeFromWire(i.GetNode()),
			State: playbackStateFromWire(i.GetState()),
		})
	}
	return out, nil
}
