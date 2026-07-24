package host

import (
	"context"

	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// contentServer runs in the Platform process and serves the module's callbacks
// into the Platform's real application service.
//
// Every method here forwards the Caller the module presented. It does not
// validate it: the Platform's own service authenticates and authorises on every
// call (ADR 0017), and the invocation-handle table is what makes a retained
// Caller stop resolving. Re-checking here would duplicate a gate that already
// exists and would suggest this layer is the one enforcing it.
type contentServer struct {
	modulev1.UnimplementedContentServiceServer

	impl       v1.ContentService
	categoryOf CategoryFunc
}

func (s *contentServer) AddContentWork(ctx context.Context, req *modulev1.AddContentWorkRequest) (*modulev1.AddContentWorkResponse, error) {
	out, err := s.impl.AddContentWork(ctx, v1.AddContentWorkCommand{
		Caller:      callerFromWire(req.GetCaller()),
		MediaType:   v1.MediaType(req.GetMediaType()),
		Title:       req.GetTitle(),
		ExternalIDs: req.GetExternalIds(),
		Attributes:  req.GetAttributes(),
		Artwork:     artworkFromWire(req.GetArtwork()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.AddContentWorkResponse{Work: nodeToWire(out.Work)}, nil
}

func (s *contentServer) AddContentChild(ctx context.Context, req *modulev1.AddContentChildRequest) (*modulev1.AddContentChildResponse, error) {
	out, err := s.impl.AddContentChild(ctx, v1.AddContentChildCommand{
		Caller:        callerFromWire(req.GetCaller()),
		ParentID:      v1.NodeID(req.GetParentId()),
		Kind:          nodeKindFromWire(req.GetKind()),
		ContainerType: v1.ContainerType(req.GetContainerType()),
		ItemType:      v1.ItemType(req.GetItemType()),
		Title:         req.GetTitle(),
		NaturalOrder:  req.GetNaturalOrder(),
		ExternalIDs:   req.GetExternalIds(),
		Attributes:    req.GetAttributes(),
		Artwork:       artworkFromWire(req.GetArtwork()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.AddContentChildResponse{Node: nodeToWire(out.Node)}, nil
}

func (s *contentServer) AttachContentPart(ctx context.Context, req *modulev1.AttachContentPartRequest) (*modulev1.AttachContentPartResponse, error) {
	out, err := s.impl.AttachContentPart(ctx, v1.AttachContentPartCommand{
		Caller:       callerFromWire(req.GetCaller()),
		NodeID:       v1.NodeID(req.GetNodeId()),
		Role:         partRoleFromWire(req.GetRole()),
		EditionLabel: req.GetEditionLabel(),
		NaturalOrder: req.GetNaturalOrder(),
		Location:     locationFromWire(req.GetLocation()),
		Container:    req.GetContainer(),
		VideoCodec:   req.GetVideoCodec(),
		AudioCodec:   req.GetAudioCodec(),
		Width:        int(req.GetWidth()),
		Height:       int(req.GetHeight()),
		HDRFormat:    req.GetHdrFormat(),
		Duration:     durationFromWire(req.GetDurationNanos()),
		BitrateBPS:   req.GetBitrateBps(),
		SizeBytes:    req.GetSizeBytes(),
		Attributes:   req.GetAttributes(),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.AttachContentPartResponse{Part: partToWire(out.Part)}, nil
}

func (s *contentServer) SetContentArtwork(ctx context.Context, req *modulev1.SetContentArtworkRequest) (*modulev1.SetContentArtworkResponse, error) {
	out, err := s.impl.SetContentArtwork(ctx, v1.SetContentArtworkCommand{
		Caller:  callerFromWire(req.GetCaller()),
		NodeID:  v1.NodeID(req.GetNodeId()),
		Artwork: artworkFromWire(req.GetArtwork()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.SetContentArtworkResponse{Node: nodeToWire(out.Node)}, nil
}

func (s *contentServer) RelateContent(ctx context.Context, req *modulev1.RelateContentRequest) (*modulev1.RelateContentResponse, error) {
	out, err := s.impl.RelateContent(ctx, v1.RelateContentCommand{
		Caller:     callerFromWire(req.GetCaller()),
		FromNodeID: v1.NodeID(req.GetFromNodeId()),
		ToNodeID:   v1.NodeID(req.GetToNodeId()),
		Type:       v1.RelationType(req.GetType()),
		Confidence: req.GetConfidence(),
		Origin:     relationOriginFromWire(req.GetOrigin()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.RelateContentResponse{Relation: relationToWire(out.Relation)}, nil
}

func (s *contentServer) BindContentSource(ctx context.Context, req *modulev1.BindContentSourceRequest) (*modulev1.BindContentSourceResponse, error) {
	out, err := s.impl.BindContentSource(ctx, v1.BindContentSourceCommand{
		Caller:          callerFromWire(req.GetCaller()),
		NodeID:          v1.NodeID(req.GetNodeId()),
		SourceProvider:  req.GetSourceProvider(),
		SourceRef:       req.GetSourceRef(),
		MatchConfidence: req.GetMatchConfidence(),
		MatchMethod:     matchMethodFromWire(req.GetMatchMethod()),
		Status:          bindingStatusFromWire(req.GetStatus()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.BindContentSourceResponse{Binding: bindingToWire(out.Binding)}, nil
}

func (s *contentServer) ResolveContentBinding(ctx context.Context, req *modulev1.ResolveContentBindingRequest) (*modulev1.ResolveContentBindingResponse, error) {
	out, err := s.impl.ResolveContentBinding(ctx, v1.ResolveContentBindingCommand{
		Caller:       callerFromWire(req.GetCaller()),
		BindingID:    v1.SourceBindingID(req.GetBindingId()),
		Resolution:   bindingResolutionFromWire(req.GetResolution()),
		MoveToNodeID: v1.NodeID(req.GetMoveToNodeId()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.ResolveContentBindingResponse{Binding: bindingToWire(out.Binding)}, nil
}

func (s *contentServer) SearchContent(ctx context.Context, req *modulev1.SearchContentRequest) (*modulev1.SearchContentResponse, error) {
	out, err := s.impl.SearchContent(ctx, v1.SearchContentQuery{
		Caller:            callerFromWire(req.GetCaller()),
		Title:             req.GetTitle(),
		MediaType:         v1.MediaType(req.GetMediaType()),
		Kind:              nodeKindFromWire(req.GetKind()),
		AttributesContain: req.GetAttributesContain(),
		Limit:             int(req.GetLimit()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.SearchContentResponse{Nodes: nodesToWire(out.Nodes)}, nil
}

func (s *contentServer) FindContentByExternalID(ctx context.Context, req *modulev1.FindContentByExternalIDRequest) (*modulev1.FindContentByExternalIDResponse, error) {
	out, err := s.impl.FindContentByExternalID(ctx, v1.FindContentByExternalIDQuery{
		Caller: callerFromWire(req.GetCaller()),
		Scheme: req.GetScheme(),
		Value:  req.GetValue(),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.FindContentByExternalIDResponse{Nodes: nodesToWire(out.Nodes)}, nil
}

func (s *contentServer) GetContentNode(ctx context.Context, req *modulev1.GetContentNodeRequest) (*modulev1.GetContentNodeResponse, error) {
	out, err := s.impl.GetContentNode(ctx, v1.GetContentNodeQuery{
		Caller:       callerFromWire(req.GetCaller()),
		NodeID:       v1.NodeID(req.GetNodeId()),
		WithChildren: req.GetWithChildren(),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.GetContentNodeResponse{
		Node:     nodeToWire(out.Node),
		Children: nodesToWire(out.Children),
	}, nil
}

func (s *contentServer) ListContentParts(ctx context.Context, req *modulev1.ListContentPartsRequest) (*modulev1.ListContentPartsResponse, error) {
	out, err := s.impl.ListContentParts(ctx, v1.ListContentPartsQuery{
		Caller: callerFromWire(req.GetCaller()),
		NodeID: v1.NodeID(req.GetNodeId()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	resp := &modulev1.ListContentPartsResponse{}
	for _, p := range out.Parts {
		resp.Parts = append(resp.Parts, partToWire(p))
	}
	return resp, nil
}

func (s *contentServer) RecordPlaybackProgress(ctx context.Context, req *modulev1.RecordPlaybackProgressRequest) (*modulev1.RecordPlaybackProgressResponse, error) {
	out, err := s.impl.RecordPlaybackProgress(ctx, v1.RecordPlaybackProgressCommand{
		Caller:   callerFromWire(req.GetCaller()),
		NodeID:   v1.NodeID(req.GetNodeId()),
		PartID:   v1.PartID(req.GetPartId()),
		Position: durationFromWire(req.GetPositionNanos()),
		Duration: durationFromWire(req.GetDurationNanos()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.RecordPlaybackProgressResponse{State: playbackStateToWire(out.State)}, nil
}

func (s *contentServer) SetPlaybackFinished(ctx context.Context, req *modulev1.SetPlaybackFinishedRequest) (*modulev1.SetPlaybackFinishedResponse, error) {
	out, err := s.impl.SetPlaybackFinished(ctx, v1.SetPlaybackFinishedCommand{
		Caller:   callerFromWire(req.GetCaller()),
		NodeID:   v1.NodeID(req.GetNodeId()),
		Finished: req.GetFinished(),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.SetPlaybackFinishedResponse{State: playbackStateToWire(out.State)}, nil
}

func (s *contentServer) GetPlaybackState(ctx context.Context, req *modulev1.GetPlaybackStateRequest) (*modulev1.GetPlaybackStateResponse, error) {
	out, err := s.impl.GetPlaybackState(ctx, v1.GetPlaybackStateQuery{
		Caller: callerFromWire(req.GetCaller()),
		NodeID: v1.NodeID(req.GetNodeId()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	return &modulev1.GetPlaybackStateResponse{
		State: playbackStateToWire(out.State),
		Found: out.Found,
	}, nil
}

func (s *contentServer) ListPlaybackStates(ctx context.Context, req *modulev1.ListPlaybackStatesRequest) (*modulev1.ListPlaybackStatesResponse, error) {
	ids := make([]v1.NodeID, 0, len(req.GetNodeIds()))
	for _, id := range req.GetNodeIds() {
		ids = append(ids, v1.NodeID(id))
	}
	out, err := s.impl.ListPlaybackStates(ctx, v1.ListPlaybackStatesQuery{
		Caller:  callerFromWire(req.GetCaller()),
		NodeIDs: ids,
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	resp := &modulev1.ListPlaybackStatesResponse{States: make(map[string]*modulev1.PlaybackState, len(out.States))}
	for id, st := range out.States {
		resp.States[string(id)] = playbackStateToWire(st)
	}
	return resp, nil
}

func (s *contentServer) ListInProgress(ctx context.Context, req *modulev1.ListInProgressRequest) (*modulev1.ListInProgressResponse, error) {
	out, err := s.impl.ListInProgress(ctx, v1.ListInProgressQuery{
		Caller: callerFromWire(req.GetCaller()),
		Limit:  int(req.GetLimit()),
	})
	if err != nil {
		return nil, errorToWire(err, s.categoryOf)
	}
	resp := &modulev1.ListInProgressResponse{}
	for _, i := range out.Items {
		resp.Items = append(resp.Items, &modulev1.InProgressItem{
			Node:  nodeToWire(i.Node),
			State: playbackStateToWire(i.State),
		})
	}
	return resp, nil
}
