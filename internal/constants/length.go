package constants

var (
	MaxThreadThemeLength     = int64(200)
	MaxThreadReasoningLength = int64(4096)
	MaxThreadSummaryLength   = int64(4096)
	MaxAnalysisLength        = int64(4096)

	MaxFactCheckingExplanationLength    = int64(2000)
	MaxFactCheckingAdditionalInfoLength = int64(2000)

	MinimumConfidenceScore = float64(0.0)
	MaximumConfidenceScore = float64(1.0)

	MaxUserBioLength      = int64(300)
	MaxUserInterestLength = int64(100)
	MaxUserHobbyLength    = int64(100)
)
