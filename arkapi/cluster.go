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

func NewCluster(data *arkcluster.Data) *ClusterAPI {
	return &ClusterAPI{data: data}
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

func (c *ClusterAPI) DinosByParseStatus(ok bool) []arkcluster.Dino {
	if c == nil || c.data == nil {
		return nil
	}
	var out []arkcluster.Dino
	for _, dino := range c.data.Dinos {
		parsed := dino.ParseError == ""
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
		parsed := dino.ParseError == ""
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