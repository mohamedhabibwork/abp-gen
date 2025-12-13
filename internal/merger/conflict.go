package merger

import "github.com/mohamedhabibwork/abp-gen/internal/prompts"

// Import types from prompts package to avoid circular dependency
type ConflictType = prompts.ConflictType
type Conflict = prompts.Conflict
type ConflictResolution = prompts.ConflictResolution

const (
	ConflictTypeDuplicateClass    = prompts.ConflictTypeDuplicateClass
	ConflictTypeDuplicateMethod   = prompts.ConflictTypeDuplicateMethod
	ConflictTypeDuplicateProperty = prompts.ConflictTypeDuplicateProperty
	ConflictTypeDifferentValue    = prompts.ConflictTypeDifferentValue
	ConflictTypeStructural        = prompts.ConflictTypeStructural
)

const (
	ResolutionKeepExisting = prompts.ResolutionKeepExisting
	ResolutionUseNew       = prompts.ResolutionUseNew
	ResolutionKeepBoth     = prompts.ResolutionKeepBoth
	ResolutionSkip         = prompts.ResolutionSkip
	ResolutionShowContext  = prompts.ResolutionShowContext
)
