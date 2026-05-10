package chunker

import (
	"strings"
	"unicode/utf8"
)

// Chunk 表示从原始文本切分出的一个片段。
type Chunk struct {
	Index   int    // 在文档中的顺序索引（0-based）
	Content string // 切块文本内容
}

// Config 切块配置。
type Config struct {
	ChunkSize int // 每个 chunk 的最大字符数（近似 token，1 token ≈ 4 chars 英文 / ≈ 1.5 chars 中文）
	Overlap   int // 相邻 chunk 之间的重叠字符数
}

// DefaultConfig 返回默认切块配置（chunk_size=1500 chars ≈ 512 tokens, overlap=200 chars ≈ 64 tokens）。
func DefaultConfig() Config {
	return Config{
		ChunkSize: 1500,
		Overlap:   200,
	}
}

// SplitText 将文本按固定窗口 + 重叠策略切分为多个 Chunk。
// 优先在段落/句子边界切分以保持语义完整性。
func SplitText(text string, cfg Config) []Chunk {
	if cfg.ChunkSize <= 0 {
		cfg = DefaultConfig()
	}
	if cfg.Overlap >= cfg.ChunkSize {
		cfg.Overlap = cfg.ChunkSize / 4
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	// 短文本直接作为单个 chunk
	if utf8.RuneCountInString(text) <= cfg.ChunkSize {
		return []Chunk{{Index: 0, Content: text}}
	}

	// 按段落分割，保留语义边界
	paragraphs := splitParagraphs(text)
	return mergeParagraphsIntoChunks(paragraphs, cfg)
}

// splitParagraphs 按双换行符分割段落，空段落会被过滤。
func splitParagraphs(text string) []string {
	raw := strings.Split(text, "\n\n")
	result := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// mergeParagraphsIntoChunks 贪心合并段落直到达到 chunk_size，然后产出一个 chunk。
// 回退 overlap 字符数以保持上下文连贯。
func mergeParagraphsIntoChunks(paragraphs []string, cfg Config) []Chunk {
	var chunks []Chunk
	var buf strings.Builder
	idx := 0

	for _, para := range paragraphs {
		paraLen := utf8.RuneCountInString(para)

		// 单个段落超过 chunk_size 时，进行硬切分
		if paraLen > cfg.ChunkSize {
			// 先 flush 已有缓冲
			if buf.Len() > 0 {
				chunks = append(chunks, Chunk{Index: idx, Content: strings.TrimSpace(buf.String())})
				idx++
				buf.Reset()
			}
			subChunks := hardSplit(para, cfg)
			for _, sc := range subChunks {
				chunks = append(chunks, Chunk{Index: idx, Content: sc})
				idx++
			}
			continue
		}

		// 合并后是否超限
		currentLen := utf8.RuneCountInString(buf.String())
		if currentLen+paraLen+1 > cfg.ChunkSize && buf.Len() > 0 {
			content := strings.TrimSpace(buf.String())
			chunks = append(chunks, Chunk{Index: idx, Content: content})
			idx++

			// 保留 overlap 部分以连接上下文
			overlap := extractOverlap(content, cfg.Overlap)
			buf.Reset()
			if overlap != "" {
				buf.WriteString(overlap)
				buf.WriteString("\n\n")
			}
		}

		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(para)
	}

	if buf.Len() > 0 {
		chunks = append(chunks, Chunk{Index: idx, Content: strings.TrimSpace(buf.String())})
	}

	return chunks
}

// hardSplit 对超长段落按字符数硬切分，尝试在句子边界切分。
func hardSplit(text string, cfg Config) []string {
	runes := []rune(text)
	var result []string

	for start := 0; start < len(runes); {
		end := start + cfg.ChunkSize
		if end > len(runes) {
			end = len(runes)
		}

		// 尝试在句子边界回退切分
		if end < len(runes) {
			boundary := findSentenceBoundary(runes[start:end])
			if boundary > cfg.ChunkSize/2 {
				end = start + boundary
			}
		}

		result = append(result, strings.TrimSpace(string(runes[start:end])))

		// 下一段起始位置需回退 overlap
		next := end - cfg.Overlap
		if next <= start {
			next = end
		}
		start = next
	}

	return result
}

// findSentenceBoundary 从右向左查找句子结束符（。！？.!?\n），返回切分位置。
func findSentenceBoundary(runes []rune) int {
	sentenceEnders := "。！？.!?\n"
	for i := len(runes) - 1; i >= 0; i-- {
		if strings.ContainsRune(sentenceEnders, runes[i]) {
			return i + 1
		}
	}
	return len(runes)
}

// extractOverlap 从 content 末尾提取 overlap 字符数作为下一个 chunk 的前缀。
func extractOverlap(content string, overlapSize int) string {
	runes := []rune(content)
	if len(runes) <= overlapSize {
		return content
	}
	return string(runes[len(runes)-overlapSize:])
}
