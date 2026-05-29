package reddit

type Post struct {
	ID          string
	Subreddit   string
	Title       string
	SelfText    string
	Author      string
	Score       int
	UpvoteRatio float64
	NumComments int
	Permalink   string
	URL         string
	CreatedUTC  int64
}
