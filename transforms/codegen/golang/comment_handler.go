package gogen

// CommentHandler allows library users to define custom handling for
// comments.
type CommentHandler interface {
	// Handles is called for stand-alone comments
	HandleComment(comment string)
	// HandleRule is called for a comment preceding a rule definition.
	HandleRuleComment(comment string, cddlTypeName string, goTypeName string)
}
