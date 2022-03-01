package response

// Pos represents a position. Used to declare where an object was defined in
// the .wharf-ci.yml file. The first line and column starts at 1.
// The zero value is used to represent an undefined position.
type Pos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
