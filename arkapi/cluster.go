package arkapi

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
)

type ClusterAPI struct {
	data *arkcluster.Data
}

type ClusterSummary struct {
	ID               string
	Path             string
	ArchiveVersion   int32
	ObjectCount      int
	ItemCount        int
	DinoCount        int
	ItemCountsByType map[string]int
	ParseErrorCount  int
}

type ClusterItemSummary struct {
	Items                   int
	DinoItems               int
	EquipmentItems          int
	OtherItems              int
	SupportedVersionItems   int
	UnsupportedVersionItems int
	CraftedItems            int
	TotalQuantity           int64
	MaxRating               float64
	MaxQuality              int32
}

type ClusterDinoSummary struct {
	Dinos                   int
	ParsedDinos             int
	ParseErrorDinos         int
	SupportedVersionDinos   int
	UnsupportedVersionDinos int
	WithStatusComponent     int
	WithAIController        int
	WithInventoryComponent  int
	WithDinoID              int
	TamedDinos              int
	FemaleDinos             int
	BabyDinos               int
	DeadDinos               int
	WithStats               int
	TotalBaseLevel          int64
	MaxBaseLevel            int32
	AverageBaseLevel        float64
	TotalCurrentLevel       int64
	MaxCurrentLevel         int32
	AverageCurrentLevel     float64
	TotalEmbeddedObjects    int
	MaxEmbeddedObjects      int
}

type ClusterDirectoryAggregate struct {
	Files       int
	Objects     int
	Items       int
	Dinos       int
	ParseErrors int
	ItemSummary ClusterItemSummary
	DinoSummary ClusterDinoSummary
}

func NewCluster(data *arkcluster.Data) *ClusterAPI {
	return &ClusterAPI{data: data}
}

func NewClusterFromPath(path string) (*ClusterAPI, error) {
	data, err := arkcluster.Open(path)
	if err != nil {
		return nil, err
	}
	return NewCluster(data), nil
}

func ClusterItemsFromPath(path string) ([]arkobject.ClusterItem, error) {
	api, err := NewClusterFromPath(path)
	if err != nil {
		return nil, err
	}
	return api.ItemsTyped(), nil
}

func ClusterDinosFromPath(path string) ([]arkobject.ClusterDino, error) {
	api, err := NewClusterFromPath(path)
	if err != nil {
		return nil, err
	}
	return api.DinosTyped(), nil
}

func (c *ClusterAPI) Items() []arkcluster.Item {
	if c == nil || c.data == nil {
		return nil
	}
	return append([]arkcluster.Item(nil), c.data.Items...)
}

func (c *ClusterAPI) ItemsTyped() []arkobject.ClusterItem {
	if c == nil || c.data == nil {
		return nil
	}
	out := make([]arkobject.ClusterItem, 0, len(c.data.Items))
	for _, item := range c.data.Items {
		out = append(out, arkobject.ClusterItemFromUpload(item, clusterItemType(item)))
	}
	return out
}

func (c *ClusterAPI) Dinos() []arkcluster.Dino {
	if c == nil || c.data == nil {
		return nil
	}
	return append([]arkcluster.Dino(nil), c.data.Dinos...)
}

func (c *ClusterAPI) DinosTyped() []arkobject.ClusterDino {
	if c == nil || c.data == nil {
		return nil
	}
	out := make([]arkobject.ClusterDino, 0, len(c.data.Dinos))
	for _, dino := range c.data.Dinos {
		var classNames []string
		if dino.Archive != nil {
			classNames = archiveClassNames(dino.Archive)
		}
		out = append(out, arkobject.ClusterDinoFromUpload(dino, classNames))
	}
	return out
}

func (c *ClusterAPI) ItemsByType(typeName string) []arkcluster.Item {
	if c == nil || c.data == nil {
		return nil
	}
	var out []arkcluster.Item
	for _, item := range c.data.Items {
		if clusterItemType(item).String() == typeName {
			out = append(out, item)
		}
	}
	return out
}

func (c *ClusterAPI) ItemsByTypeTyped(typeName string) []arkobject.ClusterItem {
	if c == nil || c.data == nil {
		return nil
	}
	var out []arkobject.ClusterItem
	for _, item := range c.data.Items {
		itemType := clusterItemType(item)
		if itemType.String() == typeName {
			out = append(out, arkobject.ClusterItemFromUpload(item, itemType))
		}
	}
	return out
}

func (c *ClusterAPI) ItemsByTypedType(itemType arkobject.ClusterItemType) []arkobject.ClusterItem {
	if c == nil || c.data == nil {
		return nil
	}
	var out []arkobject.ClusterItem
	for _, item := range c.data.Items {
		currentType := clusterItemType(item)
		if currentType == itemType {
			out = append(out, arkobject.ClusterItemFromUpload(item, currentType))
		}
	}
	return out
}

func (c *ClusterAPI) DinosByParseStatus(ok bool) []arkcluster.Dino {
	if c == nil || c.data == nil {
		return nil
	}
	var out []arkcluster.Dino
	for _, dino := range c.data.Dinos {
		parsed := dino.Archive != nil && dino.ParseError == ""
		if parsed == ok {
			out = append(out, dino)
		}
	}
	return out
}

func (c *ClusterAPI) DinosByParseStatusTyped(ok bool) []arkobject.ClusterDino {
	if c == nil || c.data == nil {
		return nil
	}
	var out []arkobject.ClusterDino
	for _, dino := range c.data.Dinos {
		parsed := dino.Archive != nil && dino.ParseError == ""
		if parsed != ok {
			continue
		}
		var classNames []string
		if dino.Archive != nil {
			classNames = archiveClassNames(dino.Archive)
		}
		out = append(out, arkobject.ClusterDinoFromUpload(dino, classNames))
	}
	return out
}

func (c *ClusterAPI) ItemCountsByType() map[string]int {
	counts := map[string]int{"dino": 0, "equipment": 0, "other": 0}
	if c == nil || c.data == nil {
		return counts
	}
	for _, item := range c.data.Items {
		counts[clusterItemType(item).String()]++
	}
	return counts
}

func (c *ClusterAPI) ItemSummary() ClusterItemSummary {
	items := c.ItemsTyped()
	summary := ClusterItemSummary{Items: len(items)}
	for _, item := range items {
		switch item.ItemType() {
		case arkobject.ClusterItemTypeDino:
			summary.DinoItems++
		case arkobject.ClusterItemTypeEquipment:
			summary.EquipmentItems++
		default:
			summary.OtherItems++
		}
		if item.SupportedVersion() {
			summary.SupportedVersionItems++
		} else if item.UnsupportedVersion() {
			summary.UnsupportedVersionItems++
		}
		if item.IsCrafted() {
			summary.CraftedItems++
		}
		summary.TotalQuantity += int64(item.Quantity)
		if item.Rating > summary.MaxRating {
			summary.MaxRating = item.Rating
		}
		if item.Quality > summary.MaxQuality {
			summary.MaxQuality = item.Quality
		}
	}
	return summary
}

func (c *ClusterAPI) ParseErrorCount() int {
	if c == nil || c.data == nil {
		return 0
	}
	var count int
	for _, dino := range c.data.Dinos {
		if dino.ParseError != "" {
			count++
		}
	}
	return count
}

func (c *ClusterAPI) DinoParseStatusCounts() map[string]int {
	counts := map[string]int{
		arkobject.ClusterDinoParseStatusParsed.String():             0,
		arkobject.ClusterDinoParseStatusUnsupportedVersion.String(): 0,
		arkobject.ClusterDinoParseStatusParseError.String():         0,
		arkobject.ClusterDinoParseStatusUnparsed.String():           0,
	}
	for _, dino := range c.DinosTyped() {
		counts[dino.ParseStatus().String()]++
	}
	return counts
}

func (c *ClusterAPI) DinoSummary() ClusterDinoSummary {
	dinos := c.DinosTyped()
	summary := ClusterDinoSummary{Dinos: len(dinos)}
	for _, dino := range dinos {
		if dino.Parsed() {
			summary.ParsedDinos++
		}
		if dino.HasParseError() {
			summary.ParseErrorDinos++
		}
		if dino.SupportedVersion() {
			summary.SupportedVersionDinos++
		} else if dino.UnsupportedVersion() {
			summary.UnsupportedVersionDinos++
		}
		if len(dino.StatusComponentClassNames) > 0 {
			summary.WithStatusComponent++
		}
		if len(dino.AIControllerClassNames) > 0 {
			summary.WithAIController++
		}
		if len(dino.InventoryComponentClassNames) > 0 {
			summary.WithInventoryComponent++
		}
		if dino.DinoID1 != 0 || dino.DinoID2 != 0 {
			summary.WithDinoID++
		}
		if dino.IsTamed {
			summary.TamedDinos++
		}
		if dino.IsFemale {
			summary.FemaleDinos++
		}
		if dino.IsBaby {
			summary.BabyDinos++
		}
		if dino.IsDead {
			summary.DeadDinos++
		}
		if dino.HasStats {
			summary.WithStats++
			summary.TotalBaseLevel += int64(dino.BaseLevel)
			if dino.BaseLevel > summary.MaxBaseLevel {
				summary.MaxBaseLevel = dino.BaseLevel
			}
			summary.TotalCurrentLevel += int64(dino.CurrentLevel)
			if dino.CurrentLevel > summary.MaxCurrentLevel {
				summary.MaxCurrentLevel = dino.CurrentLevel
			}
		}
		summary.TotalEmbeddedObjects += dino.ObjectCount
		if dino.ObjectCount > summary.MaxEmbeddedObjects {
			summary.MaxEmbeddedObjects = dino.ObjectCount
		}
	}
	if summary.WithStats > 0 {
		summary.AverageBaseLevel = float64(summary.TotalBaseLevel) / float64(summary.WithStats)
		summary.AverageCurrentLevel = float64(summary.TotalCurrentLevel) / float64(summary.WithStats)
	}
	return summary
}

func (c *ClusterAPI) Summary() ClusterSummary {
	summary := ClusterSummary{ItemCountsByType: c.ItemCountsByType()}
	if c == nil || c.data == nil {
		return summary
	}
	summary.ID = c.data.ID
	summary.Path = c.data.Path
	summary.ItemCount = len(c.data.Items)
	summary.DinoCount = len(c.data.Dinos)
	summary.ParseErrorCount = c.ParseErrorCount()
	if c.data.Archive != nil {
		summary.ArchiveVersion = c.data.Archive.Version
		summary.ObjectCount = len(c.data.Archive.Objects)
	}
	return summary
}

func ClusterDirectorySummary(entries []*arkcluster.Data) ClusterDirectoryAggregate {
	summary := ClusterDirectoryAggregate{Files: len(entries)}
	for _, entry := range entries {
		api := NewCluster(entry)
		fileSummary := api.Summary()
		summary.Objects += fileSummary.ObjectCount
		summary.Items += fileSummary.ItemCount
		summary.Dinos += fileSummary.DinoCount
		summary.ParseErrors += fileSummary.ParseErrorCount
		addClusterItemSummary(&summary.ItemSummary, api.ItemSummary())
		addClusterDinoSummary(&summary.DinoSummary, api.DinoSummary())
	}
	return summary
}

func addClusterItemSummary(total *ClusterItemSummary, next ClusterItemSummary) {
	total.Items += next.Items
	total.DinoItems += next.DinoItems
	total.EquipmentItems += next.EquipmentItems
	total.OtherItems += next.OtherItems
	total.SupportedVersionItems += next.SupportedVersionItems
	total.UnsupportedVersionItems += next.UnsupportedVersionItems
	total.CraftedItems += next.CraftedItems
	total.TotalQuantity += next.TotalQuantity
	if next.MaxRating > total.MaxRating {
		total.MaxRating = next.MaxRating
	}
	if next.MaxQuality > total.MaxQuality {
		total.MaxQuality = next.MaxQuality
	}
}

func addClusterDinoSummary(total *ClusterDinoSummary, next ClusterDinoSummary) {
	total.Dinos += next.Dinos
	total.ParsedDinos += next.ParsedDinos
	total.ParseErrorDinos += next.ParseErrorDinos
	total.SupportedVersionDinos += next.SupportedVersionDinos
	total.UnsupportedVersionDinos += next.UnsupportedVersionDinos
	total.WithStatusComponent += next.WithStatusComponent
	total.WithAIController += next.WithAIController
	total.WithInventoryComponent += next.WithInventoryComponent
	total.WithDinoID += next.WithDinoID
	total.TamedDinos += next.TamedDinos
	total.FemaleDinos += next.FemaleDinos
	total.BabyDinos += next.BabyDinos
	total.DeadDinos += next.DeadDinos
	total.WithStats += next.WithStats
	total.TotalBaseLevel += next.TotalBaseLevel
	if next.MaxBaseLevel > total.MaxBaseLevel {
		total.MaxBaseLevel = next.MaxBaseLevel
	}
	total.TotalCurrentLevel += next.TotalCurrentLevel
	if next.MaxCurrentLevel > total.MaxCurrentLevel {
		total.MaxCurrentLevel = next.MaxCurrentLevel
	}
	total.TotalEmbeddedObjects += next.TotalEmbeddedObjects
	if next.MaxEmbeddedObjects > total.MaxEmbeddedObjects {
		total.MaxEmbeddedObjects = next.MaxEmbeddedObjects
	}
	if total.WithStats > 0 {
		total.AverageBaseLevel = float64(total.TotalBaseLevel) / float64(total.WithStats)
		total.AverageCurrentLevel = float64(total.TotalCurrentLevel) / float64(total.WithStats)
	}
}
