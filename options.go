package aimakeme

import "sync"

// Options are...
// TODO Ned to make a New method probably so i can wrap the wg creation
type Options struct {
	Folder *string
	Prompt *string
	N      *int
	Style  *string
	APIKey string
	Wait   *sync.WaitGroup
}
