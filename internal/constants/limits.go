package constants

// Content limits - synced with backend utils/token.go
// Based on Gemini 1.5 Pro's 1M token context window

const (
	// MaxMemoryChars is the maximum characters allowed for a single memory (~750K tokens)
	MaxMemoryChars = 3_000_000

	// MaxAIInputChars is the maximum characters for AI operations (~800K tokens)
	MaxAIInputChars = 3_200_000

	// CharsPerToken is the average number of characters per token
	CharsPerToken = 4

	// WarningThresholdPercent is when to warn users about content length
	WarningThresholdPercent = 80
)

// EstimateTokens estimates the number of tokens in a string
func EstimateTokens(text string) int {
	return len(text) / CharsPerToken
}

// IsWithinMemoryLimit checks if content is within memory storage limit
func IsWithinMemoryLimit(content string) bool {
	return len(content) <= MaxMemoryChars
}

// GetContentStats returns content statistics
func GetContentStats(content string) (chars int, tokens int, usagePercent float64) {
	chars = len(content)
	tokens = EstimateTokens(content)
	usagePercent = float64(chars) / float64(MaxMemoryChars) * 100
	return
}

// FormatNumber formats large numbers with K/M suffix
func FormatNumber(num int) string {
	if num >= 1_000_000 {
		return formatFloat(float64(num)/1_000_000) + "M"
	}
	if num >= 1_000 {
		return formatFloat(float64(num)/1_000) + "K"
	}
	return formatInt(num)
}

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return formatInt(int(f))
	}
	return formatIntWithDecimal(f)
}

func formatInt(i int) string {
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}

func formatIntWithDecimal(f float64) string {
	return string(rune('0'+int(f))) + "." + string(rune('0'+int((f-float64(int(f)))*10)))
}

