package model

import "errors"

// EmbeddingModels 当前支持的 Embedding 模型白名单
var EmbeddingModels = map[string]int{
	"text-embedding-3-small": 1536,
	"text-embedding-3-large": 3072,
	"text-embedding-ada-002": 1536,
}

type KnowledgeBase struct {
	Name        string
	Description string
	Model       string
}

func (kb *KnowledgeBase) CheckName() error {
	if kb.Name == "" {
		return errors.New("知识库名称不能为空")
	}
	if len(kb.Name) > 255 {
		return errors.New("知识库名称不能超过 255 个字符")
	}
	return nil
}

// ResolveDimension 根据模型名称解析向量维度，未知模型默认 1536
func (kb *KnowledgeBase) ResolveDimension() int {
	if dim, ok := EmbeddingModels[kb.Model]; ok {
		return dim
	}
	return 1536
}
