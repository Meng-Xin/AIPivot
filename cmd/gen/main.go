package main

import (
	"aipivot/internal/shared/po"

	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:           "./internal/shared/query",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: false,
		FieldWithTypeTag:  true,
	})

	g.ApplyBasic(
		po.User{},
		po.Tenant{},
		po.ApiKey{},
		po.KnowledgeBase{},
		po.Document{},
		po.DocumentChunk{},
		po.Conversation{},
		po.Message{},
		po.Webhook{},
		po.Skill{},
		po.Flow{},
		po.FlowRun{},
	)

	g.Execute()
}
