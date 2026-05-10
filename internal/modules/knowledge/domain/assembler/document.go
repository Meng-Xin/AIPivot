package assembler

import (
	"aipivot/internal/shared/po"
	"aipivot/internal/types"
)

// ③ PO → ShowType
func DocumentPoToShow(doc *po.Document) types.ShowDocument {
	return types.ShowDocument{
		ID:          doc.ID,
		UUID:        doc.UUID,
		Name:        doc.Name,
		ContentType: doc.ContentType,
		FileSize:    doc.FileSize,
		ChunkCount:  doc.ChunkCount,
		Status:      doc.Status,
		ErrorMsg:    doc.ErrorMsg,
		CreatedAt:   doc.CreatedAt.Unix(),
		UpdatedAt:   doc.UpdatedAt.Unix(),
	}
}

// ③ PO List → ShowType List
func DocumentPoListToShowList(list []*po.Document) []types.ShowDocument {
	result := make([]types.ShowDocument, 0, len(list))
	for _, doc := range list {
		result = append(result, DocumentPoToShow(doc))
	}
	return result
}
