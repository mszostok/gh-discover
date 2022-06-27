package gh

type ReactionType string

const (
	// Confused represents the :confused: emoji.
	Confused ReactionType = "CONFUSED"
	// ThumbsDown represents the :-1: emoji.
	ThumbsDown = "THUMBS_DOWN"
	// Laugh represents the :laugh: emoji.
	Laugh = "LAUGH"

	// ThumbsUp represents the :+1: emoji.
	ThumbsUp = "THUMBS_UP"
	// Hooray represents the :hooray: emoji.
	Hooray = "HOORAY"
	// Heart represents the :heart: emoji.
	Heart = "HEART"
	// Rocket represents the :rocket: emoji.
	Rocket = "ROCKET"
	// Eyes represents the :eyes: emoji.
	Eyes = "EYES"
)

func (r ReactionType) Emoji() string {
	switch r {
	case Confused:
		return ":confused:"
	case ThumbsDown:
		return ":-1:"
	case Laugh:
		return ":laugh:"
	case ThumbsUp:
		return ":+1:"
	case Hooray:
		return ":hooray:"
	case Heart:
		return ":heart:"
	case Rocket:
		return ":rocket:"
	case Eyes:
		return ":eyes:"
	}
	return ""
}

func (r ReactionType) HasNegativeVibe() bool {
	switch r {
	case Confused, ThumbsDown, Laugh:
		return true
	default:
		return false
	}
}
