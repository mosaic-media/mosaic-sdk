package host

import (
	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// Conversions for the virtual content plane (ADR 0028) — the transient
// projections the read roles return, which are never object-graph nodes.
//
// SearchResult, CatalogItem and RelatedItem are identical in shape and are
// converted separately rather than through one helper, for the same reason the
// SDK declares them as three types: they are produced by different roles and
// answer different questions, and a shared converter would quietly propagate a
// field added to one of them into the other two.

func searchResultToWire(r v1.SearchResult) *modulev1.SearchResult {
	return &modulev1.SearchResult{
		Ref:       refToWire(r.Ref),
		Title:     r.Title,
		Year:      int32(r.Year),
		Poster:    r.Poster,
		InLibrary: r.InLibrary,
		NodeId:    string(r.NodeID),
	}
}

func searchResultFromWire(r *modulev1.SearchResult) v1.SearchResult {
	if r == nil {
		return v1.SearchResult{}
	}
	return v1.SearchResult{
		Ref:       refFromWire(r.GetRef()),
		Title:     r.GetTitle(),
		Year:      int(r.GetYear()),
		Poster:    r.GetPoster(),
		InLibrary: r.GetInLibrary(),
		NodeID:    v1.NodeID(r.GetNodeId()),
	}
}

func catalogItemToWire(i v1.CatalogItem) *modulev1.CatalogItem {
	return &modulev1.CatalogItem{
		Ref:       refToWire(i.Ref),
		Title:     i.Title,
		Year:      int32(i.Year),
		Poster:    i.Poster,
		InLibrary: i.InLibrary,
		NodeId:    string(i.NodeID),
	}
}

func catalogItemFromWire(i *modulev1.CatalogItem) v1.CatalogItem {
	if i == nil {
		return v1.CatalogItem{}
	}
	return v1.CatalogItem{
		Ref:       refFromWire(i.GetRef()),
		Title:     i.GetTitle(),
		Year:      int(i.GetYear()),
		Poster:    i.GetPoster(),
		InLibrary: i.GetInLibrary(),
		NodeID:    v1.NodeID(i.GetNodeId()),
	}
}

func relatedItemToWire(i v1.RelatedItem) *modulev1.RelatedItem {
	return &modulev1.RelatedItem{
		Ref:       refToWire(i.Ref),
		Title:     i.Title,
		Year:      int32(i.Year),
		Poster:    i.Poster,
		InLibrary: i.InLibrary,
		NodeId:    string(i.NodeID),
	}
}

func relatedItemFromWire(i *modulev1.RelatedItem) v1.RelatedItem {
	if i == nil {
		return v1.RelatedItem{}
	}
	return v1.RelatedItem{
		Ref:       refFromWire(i.GetRef()),
		Title:     i.GetTitle(),
		Year:      int(i.GetYear()),
		Poster:    i.GetPoster(),
		InLibrary: i.GetInLibrary(),
		NodeID:    v1.NodeID(i.GetNodeId()),
	}
}

func catalogToWire(c v1.Catalog) *modulev1.Catalog {
	return &modulev1.Catalog{Id: c.ID, NativeType: c.NativeType, Name: c.Name}
}

func catalogFromWire(c *modulev1.Catalog) v1.Catalog {
	if c == nil {
		return v1.Catalog{}
	}
	return v1.Catalog{ID: c.GetId(), NativeType: c.GetNativeType(), Name: c.GetName()}
}

func personToWire(p v1.Person) *modulev1.Person {
	return &modulev1.Person{Name: p.Name, Role: p.Role, Photo: p.Photo}
}

func personFromWire(p *modulev1.Person) v1.Person {
	if p == nil {
		return v1.Person{}
	}
	return v1.Person{Name: p.GetName(), Role: p.GetRole(), Photo: p.GetPhoto()}
}

func episodePreviewToWire(e v1.EpisodePreview) *modulev1.EpisodePreview {
	return &modulev1.EpisodePreview{
		Season:    int32(e.Season),
		Episode:   int32(e.Episode),
		Title:     e.Title,
		Overview:  e.Overview,
		Thumbnail: e.Thumbnail,
		Released:  e.Released,
	}
}

func episodePreviewFromWire(e *modulev1.EpisodePreview) v1.EpisodePreview {
	if e == nil {
		return v1.EpisodePreview{}
	}
	return v1.EpisodePreview{
		Season:    int(e.GetSeason()),
		Episode:   int(e.GetEpisode()),
		Title:     e.GetTitle(),
		Overview:  e.GetOverview(),
		Thumbnail: e.GetThumbnail(),
		Released:  e.GetReleased(),
	}
}

func trailerToWire(t v1.Trailer) *modulev1.Trailer {
	return &modulev1.Trailer{Name: t.Name, Site: t.Site, Key: t.Key, Official: t.Official}
}

func trailerFromWire(t *modulev1.Trailer) v1.Trailer {
	if t == nil {
		return v1.Trailer{}
	}
	return v1.Trailer{Name: t.GetName(), Site: t.GetSite(), Key: t.GetKey(), Official: t.GetOfficial()}
}

func collectionToWire(c *v1.Collection) *modulev1.Collection {
	if c == nil {
		return nil
	}
	out := &modulev1.Collection{
		Name:     c.Name,
		Overview: c.Overview,
		Poster:   c.Poster,
		Backdrop: c.Backdrop,
	}
	for _, i := range c.Items {
		out.Items = append(out.Items, relatedItemToWire(i))
	}
	return out
}

func collectionFromWire(c *modulev1.Collection) *v1.Collection {
	if c == nil {
		return nil
	}
	out := &v1.Collection{
		Name:     c.GetName(),
		Overview: c.GetOverview(),
		Poster:   c.GetPoster(),
		Backdrop: c.GetBackdrop(),
	}
	for _, i := range c.GetItems() {
		out.Items = append(out.Items, relatedItemFromWire(i))
	}
	return out
}

func watchToWire(w *v1.WatchAvailability) *modulev1.WatchAvailability {
	if w == nil {
		return nil
	}
	out := &modulev1.WatchAvailability{
		Region:      w.Region,
		Link:        w.Link,
		Attribution: w.Attribution,
	}
	for _, o := range w.Offers {
		out.Offers = append(out.Offers, &modulev1.WatchOffer{
			Provider: o.Provider,
			Logo:     o.Logo,
			Type:     watchOfferTypeToWire(o.Type),
		})
	}
	return out
}

func watchFromWire(w *modulev1.WatchAvailability) *v1.WatchAvailability {
	if w == nil {
		return nil
	}
	out := &v1.WatchAvailability{
		Region:      w.GetRegion(),
		Link:        w.GetLink(),
		Attribution: w.GetAttribution(),
	}
	for _, o := range w.GetOffers() {
		out.Offers = append(out.Offers, v1.WatchOffer{
			Provider: o.GetProvider(),
			Logo:     o.GetLogo(),
			Type:     watchOfferTypeFromWire(o.GetType()),
		})
	}
	return out
}

func metadataToWire(m v1.ContentMetadata) *modulev1.ContentMetadata {
	out := &modulev1.ContentMetadata{
		Ref:           refToWire(m.Ref),
		Title:         m.Title,
		Year:          int32(m.Year),
		Overview:      m.Overview,
		Poster:        m.Poster,
		Backdrop:      m.Backdrop,
		Genres:        m.Genres,
		Logo:          m.Logo,
		Rating:        m.Rating,
		Runtime:       m.Runtime,
		Keywords:      m.Keywords,
		Certification: m.Certification,
		Collection:    collectionToWire(m.Collection),
		Watch:         watchToWire(m.Watch),
	}
	for _, p := range m.Cast {
		out.Cast = append(out.Cast, personToWire(p))
	}
	for _, e := range m.Episodes {
		out.Episodes = append(out.Episodes, episodePreviewToWire(e))
	}
	for _, s := range m.Similar {
		out.Similar = append(out.Similar, relatedItemToWire(s))
	}
	for _, t := range m.Trailers {
		out.Trailers = append(out.Trailers, trailerToWire(t))
	}
	return out
}

func metadataFromWire(m *modulev1.ContentMetadata) v1.ContentMetadata {
	if m == nil {
		return v1.ContentMetadata{}
	}
	out := v1.ContentMetadata{
		Ref:           refFromWire(m.GetRef()),
		Title:         m.GetTitle(),
		Year:          int(m.GetYear()),
		Overview:      m.GetOverview(),
		Poster:        m.GetPoster(),
		Backdrop:      m.GetBackdrop(),
		Genres:        m.GetGenres(),
		Logo:          m.GetLogo(),
		Rating:        m.GetRating(),
		Runtime:       m.GetRuntime(),
		Keywords:      m.GetKeywords(),
		Certification: m.GetCertification(),
		Collection:    collectionFromWire(m.GetCollection()),
		Watch:         watchFromWire(m.GetWatch()),
	}
	for _, p := range m.GetCast() {
		out.Cast = append(out.Cast, personFromWire(p))
	}
	for _, e := range m.GetEpisodes() {
		out.Episodes = append(out.Episodes, episodePreviewFromWire(e))
	}
	for _, s := range m.GetSimilar() {
		out.Similar = append(out.Similar, relatedItemFromWire(s))
	}
	for _, t := range m.GetTrailers() {
		out.Trailers = append(out.Trailers, trailerFromWire(t))
	}
	return out
}

func streamLinkToWire(s v1.StreamLink) *modulev1.StreamLink {
	return &modulev1.StreamLink{
		Label:     s.Label,
		Title:     s.Title,
		Quality:   s.Quality,
		SizeBytes: s.SizeBytes,
		Seeders:   int32(s.Seeders),
		Location:  locationToWire(s.Location),
	}
}

func streamLinkFromWire(s *modulev1.StreamLink) v1.StreamLink {
	if s == nil {
		return v1.StreamLink{}
	}
	return v1.StreamLink{
		Label:     s.GetLabel(),
		Title:     s.GetTitle(),
		Quality:   s.GetQuality(),
		SizeBytes: s.GetSizeBytes(),
		Seeders:   int(s.GetSeeders()),
		Location:  locationFromWire(s.GetLocation()),
	}
}

func subtitleToWire(s v1.Subtitle) *modulev1.Subtitle {
	return &modulev1.Subtitle{Language: s.Language, Url: s.URL, Id: s.ID}
}

func subtitleFromWire(s *modulev1.Subtitle) v1.Subtitle {
	if s == nil {
		return v1.Subtitle{}
	}
	return v1.Subtitle{Language: s.GetLanguage(), URL: s.GetUrl(), ID: s.GetId()}
}

func identityToWire(i v1.ExternalIdentity) *modulev1.ExternalIdentity {
	return &modulev1.ExternalIdentity{Scheme: i.Scheme, Id: i.ID}
}

func identityFromWire(i *modulev1.ExternalIdentity) v1.ExternalIdentity {
	if i == nil {
		return v1.ExternalIdentity{}
	}
	return v1.ExternalIdentity{Scheme: i.GetScheme(), ID: i.GetId()}
}

func manifestToWire(m v1.Manifest) *modulev1.Manifest {
	out := &modulev1.Manifest{
		Id:      m.ID,
		Version: m.Version,
		Name:    m.Name,
	}
	for _, r := range m.Provides {
		out.Provides = append(out.Provides, string(r))
	}
	return out
}

func manifestFromWire(m *modulev1.Manifest) v1.Manifest {
	if m == nil {
		return v1.Manifest{}
	}
	out := v1.Manifest{
		ID:      m.GetId(),
		Version: m.GetVersion(),
		Name:    m.GetName(),
	}
	for _, r := range m.GetProvides() {
		out.Provides = append(out.Provides, v1.Role(r))
	}
	return out
}
